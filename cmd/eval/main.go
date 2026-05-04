package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/safebites/backend-go/internal/config"
	"github.com/safebites/backend-go/internal/observability"
)

func main() {
	var (
		agentFlag      = flag.String("agent", "all", "vision|search|scorer|recommender|all")
		goldenDir      = flag.String("golden-dir", "evals/golden", "root golden cases dir")
		baselinePath   = flag.String("baseline", "evals/baselines/main.json", "baseline JSON for gating")
		updateBaseline = flag.Bool("update-baseline", false, "write current run as new baseline (use only on main)")
		gate           = flag.Bool("gate", true, "fail process on gate violations")
	)
	flag.Parse()

	cfg := config.Load()

	ctx := context.Background()
	shutdown := observability.InitTracer(ctx, cfg.Langfuse)
	defer func() {
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(sctx)
	}()

	runner := NewRunner(cfg, *goldenDir)
	report, err := runner.RunAgent(ctx, *agentFlag)
	if err != nil {
		log.Fatalf("eval run failed: %v", err)
	}

	fmt.Println(report.MarkdownTable())

	if *updateBaseline {
		if err := report.WriteBaseline(*baselinePath); err != nil {
			log.Fatalf("write baseline: %v", err)
		}
		fmt.Printf("baseline written to %s\n", *baselinePath)
		return
	}

	if !*gate {
		return
	}

	failures := EvaluateGates(report, *baselinePath)
	if len(failures) > 0 {
		fmt.Fprintln(os.Stderr, "GATES FAILED:")
		for _, f := range failures {
			fmt.Fprintln(os.Stderr, "  - "+f)
		}
		os.Exit(1)
	}
	fmt.Println("All gates passed.")
}
