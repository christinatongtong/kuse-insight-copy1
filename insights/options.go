package insights

import "github.com/tmc/langchaingo/llms"

func WithModel(model llms.Model) UserInsightsOption {
	return func(insights *UserInsights) { insights.model = model }
}

func WithMaxConcurrency(n int) UserInsightsOption {
	return func(insights *UserInsights) { insights.maxConcurrency = n }
}

func WithContinue(isContinue bool) UserInsightsOption {
	return func(insights *UserInsights) { insights.isContinue = isContinue }
}
