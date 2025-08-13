package insights

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kuse-ai/kuse-insight-go/logger"
	"github.com/kuse-ai/kuse-insight-go/types"
	"github.com/tmc/langchaingo/llms"
	"go.uber.org/zap"
)

const (
	USER_CLUSTER_SYSTEM_PROMPT = `
You are a classification expert. Your task is to categorize users into predefined high-level occupation and industry groups.

I will provide you a JSON object with a user's occupation and industry. Based on that, assign the user to one of the following occupation categories and one of the following industry categories.

### Occupation Categories:
- Data Analysis
- Student
- Teacher
- Designer
- Marketing
- Healthcare
- Tech Engineer
- Other

### Industry Categories:
- Technology & Software
- Education
- Healthcare
- Finance & Business Services
- Media & Design
- Government & Non-Profit
- Science & Research
- Manufacturing & Hardware
- Other

Respond in JSON format like this:

{
  "occupation": "<one of the occupation categories>",
  "industry": "<one of the industry categories>",
}

Here is the user data:
{"occupation": "xxx", "industry": "xxx"}

	`
	USER_CLUSTER_TEMPLATE = `
Here is the user data:
{"occupation": %s, "industry": %s}
	`
)

type UserCluseter struct {
	Occupation string `json:"occupation"`
	Industry   string `json:"industry"`
}

func (insight *UserInsights) Cluster() {
	users := insight.output.Results()
	if len(users) == 0 {
		return
	}

	semaphore := make(chan struct{}, insight.maxConcurrency)

	var wg sync.WaitGroup

	logger.Info("StartClustering...",
		zap.Int("TotalUserIds", len(users)),
		zap.Int("MaxConcurrency", insight.maxConcurrency))

	for _, user := range users {
		if user == nil {
			logger.Error("UserIsNil")
			continue
		}
		semaphore <- struct{}{}
		wg.Add(1)
		go func(user *types.UserResult) {
			defer func() { <-semaphore }()
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
			defer cancel()
			user, err := insight.cluster(ctx, user)
			if err != nil {
				logger.Error("ClusterUserError", zap.String("UserID", user.UserId), zap.String("Email", user.Email), zap.Error(err))
				return
			}
			insight.output.Add(user)
		}(user)
	}

	wg.Wait()

	insight.Save()
}

func (insight *UserInsights) cluster(ctx context.Context, user *types.UserResult) (*types.UserResult, error) {
	messages := make([]llms.MessageContent, 0)
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(USER_CLUSTER_SYSTEM_PROMPT),
		},
	})
	prompt := fmt.Sprintf(USER_CLUSTER_TEMPLATE, user.Occupation, user.Industry)
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(prompt),
		},
	})

	// logger.Info("LLMRequest", zap.Any("Prompt", prompt))
	reply, err := insight.model.GenerateContent(ctx, messages)
	if err != nil {
		logger.Error("ClusterInvokeLLM Error", zap.Error(err))
		return user, err
	}
	predict := &UserCluseter{}
	content := reply.Choices[0].Content
	err = json.Unmarshal([]byte(content), predict)
	if err != nil {
		logger.Error("UnmarshalClusterError", zap.Any("Content", content), zap.Error(err))
		return user, err
	}
	user.Occupation = predict.Occupation
	user.Industry = predict.Industry
	logger.Info("UserClusterResult",
		zap.String("UserId", user.UserId),
		zap.String("Email", user.Email),
		zap.Any("Industry", user.Industry),
		zap.Any("Occupation", user.Occupation))
	return user, nil
}
