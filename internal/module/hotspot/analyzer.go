package hotspot

import "context"

type Analyzer interface {
	AnalyzeContent(ctx context.Context, content, keyword string, expandedKeywords []string) (*AnalysisResult, error)
	BatchAnalyze(ctx context.Context, contents []string, keyword string, expandedKeywords []string) ([]*AnalysisResult, error)
	ExpandKeyword(ctx context.Context, keyword string) ([]string, error)
}
