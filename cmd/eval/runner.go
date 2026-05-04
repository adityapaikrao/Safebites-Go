package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/safebites/backend-go/evals/metrics"
	"github.com/safebites/backend-go/internal/agent"
	"github.com/safebites/backend-go/internal/config"
	sbmodel "github.com/safebites/backend-go/internal/model"
)

type Runner struct {
	cfg       *config.Config
	goldenDir string
}

func NewRunner(cfg *config.Config, goldenDir string) *Runner {
	return &Runner{cfg: cfg, goldenDir: goldenDir}
}

// Report is what we print + persist.
type Report struct {
	Agents map[string]map[string]float64 `json:"agents"` // agent → metric → value
	Cases  map[string][]metrics.CaseResult `json:"cases"`
}

func newReport() *Report {
	return &Report{
		Agents: map[string]map[string]float64{},
		Cases:  map[string][]metrics.CaseResult{},
	}
}

// VisionCase / SearchCase / ScorerCase / RecommenderCase mirror the
// JSON file shape per agent. Keep them simple — no nested optionality.
type VisionCase struct {
	ID    string `json:"id"`
	Input struct {
		ImagePath string `json:"image_path"`
		Mime      string `json:"mime"`
	} `json:"input"`
	Expected struct {
		ProductName string `json:"product_name"`
	} `json:"expected"`
}

type SearchCase struct {
	ID    string `json:"id"`
	Input struct {
		ProductName string `json:"product_name"`
	} `json:"input"`
	Expected struct {
		Ingredients []string `json:"ingredients"`
	} `json:"expected"`
}

type ScorerCase struct {
	ID    string `json:"id"`
	Input struct {
		Ingredients []sbmodel.Ingredient    `json:"ingredients"`
		Prefs       sbmodel.UserPreferences `json:"prefs"`
	} `json:"input"`
	Expected struct {
		AllergyRespected     bool              `json:"allergy_respected"`
		IngredientDirections map[string]string `json:"ingredient_directions"` // name → "good"|"neutral"|"bad"
		OverallScoreRange    [2]float64        `json:"overall_score_range"`
	} `json:"expected"`
}

type RecommenderCase struct {
	ID    string `json:"id"`
	Input struct {
		ProductName  string  `json:"product_name"`
		CurrentScore float64 `json:"current_score"`
	} `json:"input"`
	Expected struct {
		// Distinctness/improvement/count are checked unconditionally.
	} `json:"expected"`
}

func (r *Runner) RunAgent(ctx context.Context, name string) (*Report, error) {
	rep := newReport()
	name = strings.ToLower(name)

	if name == "all" || name == "vision" {
		if err := r.runVision(ctx, rep); err != nil { return nil, err }
	}
	if name == "all" || name == "search" {
		if err := r.runSearch(ctx, rep); err != nil { return nil, err }
	}
	if name == "all" || name == "scorer" {
		if err := r.runScorer(ctx, rep); err != nil { return nil, err }
	}
	if name == "all" || name == "recommender" {
		if err := r.runRecommender(ctx, rep); err != nil { return nil, err }
	}
	return rep, nil
}

// ---------------- Vision ----------------

func (r *Runner) runVision(ctx context.Context, rep *Report) error {
	cases, err := loadCases[VisionCase](filepath.Join(r.goldenDir, "vision"))
	if err != nil { return err }

	vis, err := agent.NewVisionOCRFromAPIKey(r.cfg.GoogleAPIKey)
	if err != nil { return fmt.Errorf("vision init: %w", err) }

	var exact, fuzzy, empty []float64
	var results []metrics.CaseResult

	for _, c := range cases {
		imgBytes, err := os.ReadFile(c.Input.ImagePath)
		if err != nil {
			results = append(results, metrics.CaseResult{ID: c.ID, Err: "read image: " + err.Error()})
			empty = append(empty, 1.0)
			exact = append(exact, 0.0)
			fuzzy = append(fuzzy, 0.0)
			continue
		}
		start := time.Now()
		out, err := vis.ExtractProductName(ctx, imgBytes, c.Input.Mime)
		latency := time.Since(start).Milliseconds()
		cr := metrics.CaseResult{ID: c.ID, LatencyMs: latency, Metrics: map[string]float64{}}

		if err != nil {
			cr.Err = err.Error()
			empty = append(empty, 1.0); exact = append(exact, 0.0); fuzzy = append(fuzzy, 0.0)
		} else {
			isEmpty := metrics.VisionEmpty(out)
			isExact := metrics.VisionExactMatch(out, c.Expected.ProductName)
			isFuzzy := metrics.VisionFuzzyMatch(out, c.Expected.ProductName, 0.85)
			empty = append(empty, boolToFloat(isEmpty))
			exact = append(exact, boolToFloat(isExact))
			fuzzy = append(fuzzy, boolToFloat(isFuzzy))
			cr.Pass = isExact || isFuzzy
			cr.Metrics["exact_match"] = boolToFloat(isExact)
			cr.Metrics["fuzzy_match"] = boolToFloat(isFuzzy)
		}
		results = append(results, cr)
	}

	rep.Agents["vision"] = map[string]float64{
		"exact_match": metrics.AggregateRate(exact),
		"fuzzy_match": metrics.AggregateRate(fuzzy),
		"empty_rate":  metrics.AggregateRate(empty),
	}
	rep.Cases["vision"] = results
	return nil
}

// ---------------- Search ----------------

func (r *Runner) runSearch(ctx context.Context, rep *Report) error {
	cases, err := loadCases[SearchCase](filepath.Join(r.goldenDir, "search"))
	if err != nil { return err }

	llm, err := agent.NewGeminiModel(ctx, r.cfg.GoogleAPIKey, "")
	if err != nil { return fmt.Errorf("llm init: %w", err) }
	sa, err := agent.NewSearchAgent(llm)
	if err != nil { return fmt.Errorf("search agent init: %w", err) }

	var recall, precision, jsonValid []float64
	var results []metrics.CaseResult

	for _, c := range cases {
		start := time.Now()
		out, err := sa.Search(ctx, c.Input.ProductName)
		latency := time.Since(start).Milliseconds()

		cr := metrics.CaseResult{ID: c.ID, LatencyMs: latency, Metrics: map[string]float64{}}
		if err != nil {
			cr.Err = err.Error()
			recall = append(recall, 0); precision = append(precision, 0); jsonValid = append(jsonValid, 0)
			results = append(results, cr)
			continue
		}

		var ingNames []string
		for _, ing := range out.ListOfIngredients {
			ingNames = append(ingNames, ing.Name)
		}
		rec := metrics.IngredientRecall(ingNames, c.Expected.Ingredients)
		prec := metrics.IngredientPrecision(ingNames, c.Expected.Ingredients)
		recall = append(recall, rec)
		precision = append(precision, prec)
		jsonValid = append(jsonValid, 1.0) // succeeded → JSON parsed
		cr.Metrics["ingredient_recall"] = rec
		cr.Metrics["ingredient_precision"] = prec
		cr.Pass = rec >= 0.5 && prec >= 0.5
		results = append(results, cr)
	}

	rep.Agents["search"] = map[string]float64{
		"ingredient_recall":    metrics.AggregateRate(recall),
		"ingredient_precision": metrics.AggregateRate(precision),
		"json_valid":           metrics.AggregateRate(jsonValid),
	}
	rep.Cases["search"] = results
	return nil
}

// ---------------- Scorer ----------------

func (r *Runner) runScorer(ctx context.Context, rep *Report) error {
	cases, err := loadCases[ScorerCase](filepath.Join(r.goldenDir, "scorer"))
	if err != nil { return err }

	llm, err := agent.NewGeminiModel(ctx, r.cfg.GoogleAPIKey, "")
	if err != nil { return fmt.Errorf("llm init: %w", err) }
	sc, err := agent.NewScorerAgent(llm)
	if err != nil { return fmt.Errorf("scorer init: %w", err) }

	var allergy, direction, jsonValid []float64
	var maePairs []metrics.Pair
	var results []metrics.CaseResult

	for _, c := range cases {
		start := time.Now()
		out, err := sc.ScoreIngredients(ctx, c.Input.Ingredients, &c.Input.Prefs)
		latency := time.Since(start).Milliseconds()

		cr := metrics.CaseResult{ID: c.ID, LatencyMs: latency, Metrics: map[string]float64{}}
		if err != nil {
			cr.Err = err.Error()
			allergy = append(allergy, 0); direction = append(direction, 0); jsonValid = append(jsonValid, 0)
			results = append(results, cr); continue
		}
		jsonValid = append(jsonValid, 1.0)

		var got []metrics.IngredientScore
		for _, s := range out.IngredientScores {
			score, _ := strconv.ParseFloat(string(s.SafetyScore), 64)
			got = append(got, metrics.IngredientScore{Name: s.IngredientName, Score: score, Reasoning: s.Reasoning})
		}

		ar := metrics.AllergyRespected(got, c.Input.Prefs.Allergies)
		da := metrics.ScoreDirectionAgreement(got, c.Expected.IngredientDirections)
		allergy = append(allergy, boolToFloat(ar))
		direction = append(direction, da)

		// MAE: only score if expected range was provided (range mid-point is the target).
		if c.Expected.OverallScoreRange[1] > c.Expected.OverallScoreRange[0] {
			mid := (c.Expected.OverallScoreRange[0] + c.Expected.OverallScoreRange[1]) / 2
			maePairs = append(maePairs, metrics.Pair{Got: out.OverallScore, Want: mid})
		}

		cr.Pass = ar
		cr.Metrics["allergy_respected"] = boolToFloat(ar)
		cr.Metrics["score_direction_agreement"] = da
		results = append(results, cr)
	}

	rep.Agents["scorer"] = map[string]float64{
		"allergy_respect_rate":      metrics.AggregateRate(allergy),
		"score_direction_agreement": metrics.AggregateRate(direction),
		"overall_score_mae":         metrics.MeanAbsoluteError(maePairs),
		"json_valid":                metrics.AggregateRate(jsonValid),
	}
	rep.Cases["scorer"] = results
	return nil
}

// ---------------- Recommender ----------------

func (r *Runner) runRecommender(ctx context.Context, rep *Report) error {
	cases, err := loadCases[RecommenderCase](filepath.Join(r.goldenDir, "recommender"))
	if err != nil { return err }

	llm, err := agent.NewGeminiModel(ctx, r.cfg.GoogleAPIKey, "")
	if err != nil { return fmt.Errorf("llm init: %w", err) }
	re, err := agent.NewRecommenderAgent(llm)
	if err != nil { return fmt.Errorf("recommender init: %w", err) }

	var distinct, improved, countOk, jsonValid []float64
	var results []metrics.CaseResult

	for _, c := range cases {
		start := time.Now()
		out, err := re.Recommend(ctx, c.Input.ProductName, c.Input.CurrentScore)
		latency := time.Since(start).Milliseconds()

		cr := metrics.CaseResult{ID: c.ID, LatencyMs: latency, Metrics: map[string]float64{}}
		if err != nil {
			cr.Err = err.Error()
			distinct = append(distinct, 0); improved = append(improved, 0); countOk = append(countOk, 0); jsonValid = append(jsonValid, 0)
			results = append(results, cr); continue
		}
		jsonValid = append(jsonValid, 1.0)

		var recs []metrics.Recommendation
		for _, r := range out.Recommendations {
			score, _ := strconv.ParseFloat(string(r.HealthScore), 64)
			recs = append(recs, metrics.Recommendation{Name: r.ProductName, Score: score})
		}
		d := metrics.DistinctFromInput(recs, c.Input.ProductName)
		si := metrics.AllScoresImproved(recs, c.Input.CurrentScore, 1.0)
		co := metrics.CountEqualsThree(recs)
		distinct = append(distinct, boolToFloat(d))
		improved = append(improved, boolToFloat(si))
		countOk = append(countOk, boolToFloat(co))

		cr.Pass = d && si && co
		cr.Metrics["distinct"] = boolToFloat(d)
		cr.Metrics["score_improvement"] = boolToFloat(si)
		cr.Metrics["count_equals_three"] = boolToFloat(co)
		results = append(results, cr)
	}

	rep.Agents["recommender"] = map[string]float64{
		"distinct_from_input":    metrics.AggregateRate(distinct),
		"score_improvement_rate": metrics.AggregateRate(improved),
		"count_equals_three":     metrics.AggregateRate(countOk),
		"json_valid":             metrics.AggregateRate(jsonValid),
	}
	rep.Cases["recommender"] = results
	return nil
}

// ---------------- helpers ----------------

func loadCases[T any](dir string) ([]T, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}
	var out []T
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") { continue }
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil { return nil, fmt.Errorf("read %s: %w", e.Name(), err) }
		var c T
		if err := json.Unmarshal(b, &c); err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		out = append(out, c)
	}
	return out, nil
}

func boolToFloat(b bool) float64 { if b { return 1.0 }; return 0.0 }

// MarkdownTable renders the per-agent aggregates as a markdown table.
func (r *Report) MarkdownTable() string {
	var sb strings.Builder
	sb.WriteString("| Agent | Metric | Value |\n|---|---|---|\n")
	agentNames := make([]string, 0, len(r.Agents))
	for k := range r.Agents { agentNames = append(agentNames, k) }
	sort.Strings(agentNames)
	for _, a := range agentNames {
		metricNames := make([]string, 0, len(r.Agents[a]))
		for k := range r.Agents[a] { metricNames = append(metricNames, k) }
		sort.Strings(metricNames)
		for _, m := range metricNames {
			fmt.Fprintf(&sb, "| %s | %s | %.4f |\n", a, m, r.Agents[a][m])
		}
	}
	return sb.String()
}

// WriteBaseline persists agent aggregates only (not per-case detail).
func (r *Report) WriteBaseline(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(map[string]any{"agents": r.Agents}, "", "  ")
	if err != nil { return err }
	return os.WriteFile(path, b, 0o644)
}
