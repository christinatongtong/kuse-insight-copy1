package insights

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/kuse-ai/kuse-insight-go/inputs"
	"github.com/kuse-ai/kuse-insight-go/logger"
	"github.com/tmc/langchaingo/llms"
	"go.uber.org/zap"
)

// Predict
// @Description: predict
func (insight *UserInsights) Predict(ctx context.Context, user *inputs.User) (*UserPredictOutput, error) {
	messages := make([]llms.MessageContent, 0)
	if user.UserModel != nil {
		user.AvatarInfo, _ = insight.Image2Text(ctx, USER_AVATAR_PROMTP, user.UserModel.ImageUrl)
	}
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextPart(USER_INSIGHT_SYSTEM_PROMPT),
		},
	})
	prompt := GenPromptForUser(user)
	messages = append(messages, llms.MessageContent{
		Role: llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{
			llms.TextPart(prompt),
		},
	})

	reply, err := insight.model.GenerateContent(ctx, messages)
	if err != nil {
		// logger.Error("InvokeLLM Error", zap.Error(err))
		return nil, err
	}
	predict := &UserPredictOutput{}
	content := reply.Choices[0].Content
	err = json.Unmarshal([]byte(content), predict)
	if err != nil {
		logger.Error("UnmarshalPredictError", zap.Any("Content", content), zap.Error(err))
		return nil, err
	}

	// logger.Info("Predict", zap.Any("Email", user.Email), zap.Any("Reply", content))
	// fmt.Println(user.Email, content)
	return predict, nil
}

// Image2Text
// @Description: GetInfomationFromImage
func (insights *UserInsights) Image2Text(ctx context.Context, prompt, imageUrl string) (string, error) {
	if imageUrl == "" {
		return "", errors.New("image_url empty")
	}
	content := llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{},
	}
	content.Parts = append(content.Parts, llms.ImageURLPart(imageUrl))
	content.Parts = append(content.Parts, llms.TextPart(prompt))
	rsp, err := insights.model.GenerateContent(ctx, []llms.MessageContent{
		content,
	})
	if err != nil {
		// logger.Error("Image2TextError", zap.String("Url", imageUrl), zap.Error(err))
		return "", err
	}
	reply := rsp.Choices[0].Content
	// logger.Info("Image2TextSuccess", zap.String("Url", imageUrl), zap.String("Result", reply))
	return reply, err
}
