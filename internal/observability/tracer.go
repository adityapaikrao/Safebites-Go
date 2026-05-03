package observability

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"

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
		otel.SetTracerProvider(nil)
	}
}

// InitTracer wires Langfuse via OTLP/HTTP. Returns a shutdown func that is
// always non-nil. On any error, returns a no-op shutdown and logs.
func InitTracer(ctx context.Context, cfg config.LangfuseConfig) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }

	if !cfg.Enabled() {
		log.Printf("observability: Langfuse keys not set, tracing disabled")
		return noop, nil
	}

	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "https://us.cloud.langfuse.com"
	}

	u, err := url.Parse(host)
	if err != nil {
		log.Printf("observability: invalid LANGFUSE_BASE_URL %q: %v — tracing disabled", host, err)
		return noop, nil
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
		return noop, nil
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
		return noop, nil
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
	_, pingSpan := tp.Tracer(tracerName).Start(ctx, "safebites.startup.ping")
	pingSpan.End()
	if pingErr := tp.ForceFlush(ctx); pingErr != nil {
		log.Printf("observability: OTLP connectivity test failed: %v — check LANGFUSE_PUBLIC_KEY/SECRET_KEY", pingErr)
	} else {
		log.Printf("observability: OTLP connectivity test passed — traces flowing to Langfuse")
	}

	return func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("tracer shutdown: %w", err)
		}
		return nil
	}, nil
}
