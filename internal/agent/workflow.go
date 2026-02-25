package agent

import (
	"context"
	"fmt"
	"iter"
	"log"

	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/workflowagents/loopagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/session"

	sbmodel "github.com/safebites/backend-go/internal/model"
)

const (
	DefaultMinAcceptableScore  = 7.0
	DefaultMaxRecommendationTx = 2
)

type WorkflowConfig struct {
	MinAcceptableScore  float64
	MaxRecommendationTx int
}

type LoopTurn struct {
	Recommendations sbmodel.RecommenderResult `json:"recommendations"`
	Score           sbmodel.ScorerResult      `json:"score"`
}

type WorkflowResult struct {
	InitialSearch sbmodel.WebSearchResult `json:"initialSearch"`
	InitialScore  sbmodel.ScorerResult    `json:"initialScore"`
	FinalScore    sbmodel.ScorerResult    `json:"finalScore"`
	Turns         []LoopTurn              `json:"turns"`
}

type Orchestrator struct {
	searcher    *SearchAgent
	scorer      *ScorerAgent
	recommender *RecommenderAgent
	cfg         WorkflowConfig
}

func NewOrchestrator(searcher *SearchAgent, scorer *ScorerAgent, recommender *RecommenderAgent, cfg WorkflowConfig) *Orchestrator {
	if cfg.MinAcceptableScore <= 0 {
		cfg.MinAcceptableScore = DefaultMinAcceptableScore
	}
	if cfg.MaxRecommendationTx <= 0 {
		cfg.MaxRecommendationTx = DefaultMaxRecommendationTx
	}

	return &Orchestrator{
		searcher:    searcher,
		scorer:      scorer,
		recommender: recommender,
		cfg:         cfg,
	}
}

func NewOrchestratorFromModel(llm adkmodel.LLM, cfg WorkflowConfig) (*Orchestrator, error) {
	searcher, err := NewSearchAgent(llm)
	if err != nil {
		return nil, err
	}
	scorer, err := NewScorerAgent(llm)
	if err != nil {
		return nil, err
	}
	recommender, err := NewRecommenderAgent(llm)
	if err != nil {
		return nil, err
	}
	return NewOrchestrator(searcher, scorer, recommender, cfg), nil
}

// AnalyzeOnly executes the search + score steps and returns immediately.
// The recommendation loop is intentionally omitted; callers trigger it separately.
func (o *Orchestrator) AnalyzeOnly(ctx context.Context, productName string, prefs *sbmodel.UserPreferences) (*sbmodel.WebSearchResult, *sbmodel.ScorerResult, error) {
	if o.searcher == nil || o.scorer == nil {
		return nil, nil, fmt.Errorf("orchestrator requires searcher and scorer")
	}

	log.Printf("analyze_only start product=%q has_prefs=%t", productName, prefs != nil)

	var (
		searchRes    *sbmodel.WebSearchResult
		initialScore *sbmodel.ScorerResult
	)

	searchStep, err := agent.New(agent.Config{
		Name:        "search_step",
		Description: "Searches product ingredients.",
		Run: func(ic agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				log.Printf("analyze_only step=search product=%q", productName)
				result, searchErr := o.searcher.Search(ic, productName)
				if searchErr != nil {
					yield(nil, searchErr)
					return
				}
				searchRes = result
				log.Printf("analyze_only step=search complete ingredients=%d", len(searchRes.ListOfIngredients))
				yield(&session.Event{LLMResponse: adkmodel.LLMResponse{Content: genai.NewContentFromText("search_complete", genai.RoleModel)}}, nil)
			}
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create adk search step: %w", err)
	}

	scoreStep, err := agent.New(agent.Config{
		Name:        "score_step",
		Description: "Scores product ingredients.",
		Run: func(ic agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				if searchRes == nil {
					yield(nil, fmt.Errorf("search step did not produce results"))
					return
				}
				log.Printf("analyze_only step=score ingredients=%d", len(searchRes.ListOfIngredients))
				result, scoreErr := o.scorer.ScoreIngredients(ic, searchRes.ListOfIngredients, prefs)
				if scoreErr != nil {
					yield(nil, scoreErr)
					return
				}
				initialScore = result
				log.Printf("analyze_only step=score complete overall_score=%.2f", initialScore.OverallScore)
				yield(&session.Event{LLMResponse: adkmodel.LLMResponse{Content: genai.NewContentFromText("score_complete", genai.RoleModel)}}, nil)
			}
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create adk score step: %w", err)
	}

	sequential, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:      "analysis_sequence",
			SubAgents: []agent.Agent{searchStep, scoreStep},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create adk sequential workflow: %w", err)
	}

	_, err = runAgentOnce(ctx, "safebites-analysis-only", sequential, productName)
	if err != nil {
		log.Printf("analyze_only failed err=%v", err)
		return nil, nil, err
	}

	if searchRes == nil || initialScore == nil {
		return nil, nil, fmt.Errorf("analysis workflow did not produce search and score results")
	}

	log.Printf("analyze_only complete product=%q overall_score=%.2f", productName, initialScore.OverallScore)
	return searchRes, initialScore, nil
}

// AnalyzeAndImprove executes:
// OCR (outside this orchestrator) -> search -> scorer -> (recommender -> scorer) loop up to max turns.
func (o *Orchestrator) AnalyzeAndImprove(ctx context.Context, productName string, prefs *sbmodel.UserPreferences) (*WorkflowResult, error) {
	if o.searcher == nil || o.scorer == nil || o.recommender == nil {
		return nil, fmt.Errorf("orchestrator requires searcher, scorer, and recommender")
	}

	log.Printf("workflow analyze start product=%q min_score=%.2f max_turns=%d has_prefs=%t", productName, o.cfg.MinAcceptableScore, o.cfg.MaxRecommendationTx, prefs != nil)

	var (
		searchRes    *sbmodel.WebSearchResult
		initialScore *sbmodel.ScorerResult
	)

	searchStep, err := agent.New(agent.Config{
		Name:        "search_step",
		Description: "Searches product ingredients.",
		Run: func(ic agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				log.Printf("workflow step start step=search product=%q", productName)
				result, searchErr := o.searcher.Search(ic, productName)
				if searchErr != nil {
					log.Printf("workflow step failed step=search product=%q err=%v", productName, searchErr)
					yield(nil, searchErr)
					return
				}
				searchRes = result
				log.Printf("workflow step complete step=search product=%q ingredients=%d", productName, len(searchRes.ListOfIngredients))
				yield(&session.Event{LLMResponse: adkmodel.LLMResponse{Content: genai.NewContentFromText("search_complete", genai.RoleModel)}}, nil)
			}
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create adk search step: %w", err)
	}

	scoreStep, err := agent.New(agent.Config{
		Name:        "score_step",
		Description: "Scores product ingredients.",
		Run: func(ic agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				if searchRes == nil {
					yield(nil, fmt.Errorf("search step did not produce results"))
					return
				}
				log.Printf("workflow step start step=score ingredients=%d", len(searchRes.ListOfIngredients))
				result, scoreErr := o.scorer.ScoreIngredients(ic, searchRes.ListOfIngredients, prefs)
				if scoreErr != nil {
					log.Printf("workflow step failed step=score err=%v", scoreErr)
					yield(nil, scoreErr)
					return
				}
				initialScore = result
				log.Printf("workflow step complete step=score overall_score=%.2f", initialScore.OverallScore)
				yield(&session.Event{LLMResponse: adkmodel.LLMResponse{Content: genai.NewContentFromText("score_complete", genai.RoleModel)}}, nil)
			}
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create adk score step: %w", err)
	}

	sequential, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:      "analysis_sequence",
			SubAgents: []agent.Agent{searchStep, scoreStep},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create adk sequential workflow: %w", err)
	}

	_, err = runAgentOnce(ctx, "safebites-analysis-workflow", sequential, productName)
	if err != nil {
		log.Printf("workflow analyze failed stage=initial_workflow err=%v", err)
		return nil, err
	}

	if searchRes == nil || initialScore == nil {
		return nil, fmt.Errorf("analysis workflow did not produce search and score results")
	}

	result := &WorkflowResult{
		InitialSearch: *searchRes,
		InitialScore:  *initialScore,
		FinalScore:    *initialScore,
		Turns:         make([]LoopTurn, 0, o.cfg.MaxRecommendationTx),
	}

	if initialScore.OverallScore >= o.cfg.MinAcceptableScore {
		log.Printf("workflow analyze complete status=accepted_without_recommendation final_score=%.2f", initialScore.OverallScore)
		return result, nil
	}

	currentScore := initialScore.OverallScore
	var (
		latestRec   *sbmodel.RecommenderResult
		latestScore *sbmodel.ScorerResult
	)

	recommendStep, err := agent.New(agent.Config{
		Name:        "recommend_step",
		Description: "Recommends alternative products.",
		Run: func(ic agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				log.Printf("workflow step start step=recommend product=%q current_score=%.2f", productName, currentScore)
				result, recErr := o.recommender.Recommend(ic, productName, currentScore)
				if recErr != nil {
					log.Printf("workflow step failed step=recommend err=%v", recErr)
					yield(nil, recErr)
					return
				}
				latestRec = result
				log.Printf("workflow step complete step=recommend recommendations=%d", len(latestRec.Recommendations))
				yield(&session.Event{LLMResponse: adkmodel.LLMResponse{Content: genai.NewContentFromText("recommend_complete", genai.RoleModel)}}, nil)
			}
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create adk recommend step: %w", err)
	}

	rescoreStep, err := agent.New(agent.Config{
		Name:        "rescore_step",
		Description: "Scores recommended products.",
		Run: func(ic agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				if latestRec == nil {
					yield(nil, fmt.Errorf("recommend step did not produce recommendations"))
					return
				}
				log.Printf("workflow step start step=rescore recommendations=%d", len(latestRec.Recommendations))
				scoreResult, scoreErr := o.scorer.ScoreRecommendations(ic, latestRec.Recommendations, prefs)
				if scoreErr != nil {
					log.Printf("workflow step failed step=rescore err=%v", scoreErr)
					yield(nil, scoreErr)
					return
				}
				latestScore = scoreResult

				result.Turns = append(result.Turns, LoopTurn{Recommendations: *latestRec, Score: *latestScore})
				result.FinalScore = *latestScore
				currentScore = latestScore.OverallScore
				log.Printf("workflow step complete step=rescore turn=%d overall_score=%.2f threshold=%.2f", len(result.Turns), latestScore.OverallScore, o.cfg.MinAcceptableScore)

				event := &session.Event{LLMResponse: adkmodel.LLMResponse{Content: genai.NewContentFromText("rescore_complete", genai.RoleModel)}}
				if latestScore.OverallScore >= o.cfg.MinAcceptableScore {
					event.Actions.Escalate = true
					log.Printf("workflow loop early_stop=true reason=threshold_reached score=%.2f", latestScore.OverallScore)
				}
				yield(event, nil)
			}
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create adk rescore step: %w", err)
	}

	loopWorkflow, err := loopagent.New(loopagent.Config{
		MaxIterations: uint(o.cfg.MaxRecommendationTx),
		AgentConfig: agent.Config{
			Name:      "recommendation_loop",
			SubAgents: []agent.Agent{recommendStep, rescoreStep},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create adk loop workflow: %w", err)
	}

	_, err = runAgentOnce(ctx, "safebites-loop-workflow", loopWorkflow, productName)
	if err != nil {
		log.Printf("workflow analyze failed stage=loop_workflow err=%v", err)
		return nil, err
	}

	log.Printf("workflow analyze complete status=finished_with_recommendations turns=%d final_score=%.2f", len(result.Turns), result.FinalScore.OverallScore)

	return result, nil
}
