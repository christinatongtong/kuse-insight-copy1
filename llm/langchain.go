package llm

import (
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func NewGPT4Dot1Model() llms.Model {
	apiKey := os.Getenv("OPENAI_API_KEY")
	// responseFormat := &openai.ResponseFormat{Type: "json_object"}
	llm, err := openai.New(
		openai.WithModel("gpt-4.1"),
		openai.WithToken(apiKey),
		// openai.WithResponseFormat(responseFormat),
	)
	if err != nil {
		log.Panic(err)
	}
	return llm
}
