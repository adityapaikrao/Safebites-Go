package observability

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/safebites/backend-go/internal/config"
)

const (
	serviceName = "safebites-go"
	tracerName  = "safebites-go/agent"
)

var (
	providerMu      sync.RWMutex
	currentProvider *sdktrace.TracerProvider
)

// SetTracerProvider lets tests inject a recorder-backed provider.
// Pass nil to revert to the global no-op provider.
func SetTracerProvider(tp *sdktrace.TracerProvider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	currentProvider = tp
	if tp != nil {
		otel.SetTracerProvider(tp)
	} else {
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
	}
}

// InitTracer wires Langfuse via OTLP/HTTP. Returns a shutdown func that is
// always non-nil. On any error, logs and returns a no-op shutdown.
func InitTracer(ctx context.Context, cfg config.LangfuseConfig) func(context.Context) error {
	noop := func(context.Context) error { return nil }

	if !cfg.Enabled() {
		log.Printf("observability: Langfuse keys not set, tracing disabled")
		return noop
	}

	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "https://us.cloud.langfuse.com"
	}

	// Normalize scheme-less hosts (e.g. "us.cloud.langfuse.com" → "https://us.cloud.langfuse.com").
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}

	u, err := url.Parse(host)
	if err != nil || u.Host == "" {
		log.Printf("observability: invalid LANGFUSE_BASE_URL %q — tracing disabled", host)
		return noop
	}

	auth := base64.StdEncoding.EncodeToString([]byte(cfg.PublicKey + ":" + cfg.SecretKey))

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(u.Host),
		otlptracehttp.WithURLPath("/api/public/otel/v1/traces"),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization": "Basic " + auth,
		}),
	}
	if u.Scheme == "http" {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		log.Printf("observability: OTLP exporter init failed: %v — tracing disabled", err)
		return noop
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		log.Printf("observability: resource init failed: %v — tracing disabled", err)
		return noop
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Route async export errors through our log stream (default handler uses log.Print but
	// doesn't prefix — this makes auth/network failures easy to spot).
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Printf("observability: OTLP export error: %v", err)
	}))

	SetTracerProvider(tp)
	log.Printf("observability: Langfuse tracing enabled (host=%s)", host)

	// Connectivity test: flush a ping span synchronously so auth/network
	// failures surface at startup rather than silently 5 s later.
	// Use a short-lived context so a bad host can't block startup indefinitely.
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	_, pingSpan := tp.Tracer(tracerName).Start(pingCtx, "safebites.startup.ping")
	pingSpan.End()
	if pingErr := tp.ForceFlush(pingCtx); pingErr != nil {
		log.Printf("observability: OTLP connectivity test failed: %v — check LANGFUSE_PUBLIC_KEY/SECRET_KEY", pingErr)
	} else {
		log.Printf("observability: OTLP connectivity test passed — traces flowing to Langfuse")
	}

	return func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("tracer shutdown: %w", err)
		}
		return nil
	}
}
