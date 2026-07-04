// Package rules implements the diagnostic rule engine.
package rules

import (
	"strings"

	"github.com/spark-debug-mcp/spark-debug-mcp/internal/models"
)

// Rule defines a diagnostic rule.
type Rule struct {
	ID             string
	Name           string
	Description    string
	Severity       models.RuleSeverity
	Recommendation string
	DocLinks       []string
	Score          float64
	Detect         func(ctx *models.AnalysisContext) []string
}

// Engine evaluates diagnostic rules against analysis context.
type Engine struct {
	rules []Rule
}

// NewEngine creates a rule engine with all registered rules.
func NewEngine() *Engine {
	return &Engine{rules: AllRules()}
}

// Evaluate runs all rules and returns matches sorted by score.
func (e *Engine) Evaluate(ctx *models.AnalysisContext) []models.RuleMatch {
	matches := make([]models.RuleMatch, 0)
	for _, rule := range e.rules {
		evidence := rule.Detect(ctx)
		if len(evidence) > 0 {
			matches = append(matches, models.RuleMatch{
				ID:             rule.ID,
				Name:           rule.Name,
				Description:    rule.Description,
				Severity:       rule.Severity,
				Evidence:       evidence,
				Recommendation: rule.Recommendation,
				Score:          rule.Score,
				DocLinks:       rule.DocLinks,
			})
		}
	}
	sortMatches(matches)
	return matches
}

func sortMatches(matches []models.RuleMatch) {
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Score > matches[i].Score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

// Register adds a custom rule to the engine.
func (e *Engine) Register(rule Rule) {
	e.rules = append(e.rules, rule)
}

// --- Helper functions for rule detection ---

func containsAny(s string, patterns ...string) bool {
	lower := strings.ToLower(s)
	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

func allLogs(ctx *models.AnalysisContext) string {
	var sb strings.Builder
	sb.WriteString(ctx.DriverLogs)
	for _, el := range ctx.ExecutorLogs {
		sb.WriteString(el.Logs)
	}
	return sb.String()
}

func allDescs(ctx *models.AnalysisContext) string {
	var sb strings.Builder
	sb.WriteString(ctx.DriverDesc)
	for _, d := range ctx.ExecutorDescs {
		sb.WriteString(d)
	}
	return sb.String()
}

func eventReasons(ctx *models.AnalysisContext, reasons ...string) []string {
	var evidence []string
	for _, ev := range ctx.Events {
		for _, r := range reasons {
			if ev.Reason == r || containsAny(ev.Message, r) {
				evidence = append(evidence, ev.Object+": "+ev.Reason+" - "+ev.Message)
			}
		}
	}
	return evidence
}

func logPatterns(ctx *models.AnalysisContext, patterns ...string) []string {
	logs := allLogs(ctx)
	var evidence []string
	for _, p := range patterns {
		if containsAny(logs, p) {
			evidence = append(evidence, "Log match: "+p)
		}
	}
	return evidence
}

func descPatterns(ctx *models.AnalysisContext, patterns ...string) []string {
	desc := allDescs(ctx)
	var evidence []string
	for _, p := range patterns {
		if containsAny(desc, p) {
			evidence = append(evidence, "Pod describe: "+p)
		}
	}
	return evidence
}

func confMissing(ctx *models.AnalysisContext, key string) []string {
	if _, ok := ctx.SparkConf.AllConf[key]; !ok {
		return []string{"Missing spark config: " + key}
	}
	return nil
}

func containerWaiting(ctx *models.AnalysisContext, reason string) []string {
	var evidence []string
	check := func(res models.PodResources) {
		for _, c := range res.Containers {
			if containsAny(c.State, reason) {
				evidence = append(evidence, res.PodName+"/"+c.Name+": "+c.State)
			}
		}
	}
	check(ctx.DriverRes)
	for _, r := range ctx.ExecutorRes {
		check(r)
	}
	return evidence
}
