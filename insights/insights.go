package insights

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kuse-ai/kuse-insight-go/inputs"
	"github.com/kuse-ai/kuse-insight-go/logger"
	"github.com/kuse-ai/kuse-insight-go/outputs"
	"github.com/kuse-ai/kuse-insight-go/tools"
	"github.com/kuse-ai/kuse-insight-go/types"
	"github.com/tmc/langchaingo/llms"
	"go.uber.org/zap"
)

type UserInsightsOption func(*UserInsights)

type UserInsights struct {
	input  *inputs.Inputs
	output *outputs.Outputs

	model          llms.Model
	maxConcurrency int
	isContinue     bool
	count          *atomic.Int32
}

func NewUserInsight(input *inputs.Inputs, output *outputs.Outputs, opts ...UserInsightsOption) *UserInsights {
	insights := &UserInsights{
		input:          input,
		output:         output,
		maxConcurrency: 1,
		count:          &atomic.Int32{},
		isContinue:     true,
	}
	for _, opt := range opts {
		opt(insights)
	}
	insights.input.Load()
	if insights.isContinue {
		insights.output.Load()
	}
	return insights
}

// RunBatch
// @Description: Batch Run User Data
func (insight *UserInsights) RunBatch(userIds []string) {
	if len(userIds) == 0 {
		logger.Info("No User Process...")
		return
	}

	semaphore := make(chan struct{}, insight.maxConcurrency)

	var wg sync.WaitGroup

	logger.Info("StartProcessing...",
		zap.Int("TotalUserIds", len(userIds)),
		zap.Int("MaxConcurrency", insight.maxConcurrency))

	for _, userId := range userIds {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(userId string) {
			defer func() { <-semaphore }()
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
			defer cancel()
			insight.Run(ctx, userId)
		}(userId)
	}

	wg.Wait()
}

// RunAll
// @Description: Run all data from base Mixpanel file: ./sources/mixpanel/users.csv
func (insight *UserInsights) RunAll() {
	userIds := insight.input.UserIds()
	insight.RunBatch(userIds)
	insight.Save()
}

// Save
// @Description: Save the Result
func (insight *UserInsights) Save() {
	logger.Info("UserInsightSaved", zap.Int32("InsightCount", insight.count.Load()))
	insight.output.Save()
}

// Run
// @Description: RunSingle
func (insight *UserInsights) Run(ctx context.Context, userId string) {
	local := insight.output.Get(userId)

	user, err := insight.input.Get(ctx, userId)
	if err != nil {
		return
	}
	isSkip := insight.IsSkip(user, local)
	logger.Info("-",
		zap.Int32("Index", insight.count.Add(1)),
		zap.Bool("IsSkip", isSkip),
		zap.String("UserId", user.UserId),
		zap.String("Email", user.Email),
		zap.Bool("IsGuest", user.IsGuestMode),
		zap.Int("TaskCount", len(user.TasksModel)),
		zap.String("Plan", user.Plan()),
		zap.String("LastTaskTime", user.LastTaskTime()),
	)
	if isSkip {
		if local != nil { // 有可能之前是guest后面变了, 以最新的为准
			local.IsGuestMode = tools.If(user.IsGuestMode, "true", "false")
			local.Email = user.Email
			local.Plan = user.Plan()
			insight.output.Add(local)
		}
		return
	}
	predict, err := insight.Predict(ctx, user)
	if err != nil {
		return
	}

	result := insight.FormatOutput(predict)
	result.Email = user.Email
	result.UserId = user.UserId
	result.IsGuestMode = tools.If(user.IsGuestMode, "true", "false")
	result.LastTaskTime = user.LastTaskTime()
	result.Plan = user.Plan()

	insight.output.Add(result)
}

// IsSkip
// @Description: is dirty data or not
func (insight *UserInsights) IsSkip(user *inputs.User, local *types.UserResult) bool {
	// No Info
	if user.UserModel == nil {
		return true
	}
	if user.Plan() == "pro" {
		return false
	}
	// Not heavy user
	if user.TasksModel == nil || len(user.TasksModel) < 1 {
		return true
	}
	return false
}

// UploadMixpanel
// @Description: Upload result to mixpanel
func (insight *UserInsights) UploadMixpanel() {
	insight.output.Upload()
}

// formatOutput
// @Description: Format Output Result Based on Predict
func (insight *UserInsights) FormatOutput(predict *UserPredictOutput) *types.UserResult {
	result := &types.UserResult{}

	// 提取IsStudent
	result.IsStudent = extractHighConfidenceValue(predict.IsStudent)
	if predict.StudentInfo != nil {
		result.School = extractHighConfidenceValue(predict.StudentInfo.School)
		result.Major = extractHighConfidenceValue(predict.StudentInfo.Major)
		result.DegreeLevel = extractHighConfidenceValue(predict.StudentInfo.DegreeLevel)
	}

	// 提取Occupation
	result.Occupation = extractHighConfidenceValue(predict.Occupation)

	// 提取Industry
	result.Industry = extractHighConfidenceValue(predict.Industry)

	// 提取PrimaryLanguage
	result.PrimaryLanguage = extractHighConfidenceValue(predict.PrimaryLanguage)
	if strings.Contains(result.PrimaryLanguage, "Cantonese") {
		result.PrimaryLanguage = "Cantonese"
	}
	if strings.Contains(result.PrimaryLanguage, "Traditional") || strings.Contains(result.PrimaryLanguage, "zh-TW") {
		result.PrimaryLanguage = "Traditional Chinese"
	}
	if strings.Contains(result.PrimaryLanguage, "Mandarin") || result.PrimaryLanguage == "Chinese" || strings.Contains(result.PrimaryLanguage, "zh-CN") {
		result.PrimaryLanguage = "Simplified Chinese"
	}

	// 提取Gender
	result.Gender = extractHighConfidenceValue(predict.Gender)
	result.Gender = strings.ToLower(result.Gender)

	// // 提取Satisfaction
	// result.Satisfaction = extractHighConfidenceValue(predict.Satisfaction)
	return result
}
