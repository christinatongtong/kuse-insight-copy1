package outputs

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kuse-ai/kuse-insight-go/logger"
	"github.com/kuse-ai/kuse-insight-go/tools"
	"github.com/kuse-ai/kuse-insight-go/types"
	"github.com/mixpanel/mixpanel-go"
	"go.uber.org/zap"
)

func UploadMixpanel(results []*types.UserResult) error {
	mp := mixpanel.NewApiClient(os.Getenv("MIXPANEL_TOKEN"),
		mixpanel.ServiceAccount(0,
			"kuse-analyst-prod.5bfe42.mp-service-account",
			"2BPTE5P3KMdNHGNgFryPKTDlvC0u6uWF",
		))
	for index, user := range results {
		// if user.IsStudent == "<nil>" || user.IsStudent == "" {
		// 	continue
		// }
		// isStudent := strings.ToLower(user.IsStudent) == "true"

		properties := map[string]any{
			// "predict_is_student":       isStudent,
			"predict_school":           strings.ToLower(tools.If(user.School != "<nil>" && user.School != "", user.School, "unknown")),
			"predict_major":            strings.ToLower(tools.If(user.Major != "<nil>" && user.Major != "", user.Major, "unknown")),
			"predict_degree_level":     strings.ToLower(tools.If(user.DegreeLevel != "<nil>" && user.DegreeLevel != "", user.DegreeLevel, "unknown")),
			"predict_occupation":       strings.ToLower(tools.If(user.Occupation != "<nil>" && user.Occupation != "", user.Occupation, "unknown")),
			"predict_industry":         strings.ToLower(tools.If(user.Industry != "<nil>" && user.Industry != "", user.Industry, "unknown")),
			"predict_primary_language": strings.ToLower(tools.If(user.PrimaryLanguage != "<nil>" && user.PrimaryLanguage != "", user.PrimaryLanguage, "unknown")),
			"predict_gender":           strings.ToLower(tools.If(user.Gender != "<nil>" && user.Gender != "", user.Gender, "unknown")),
		}
		// People Set
		if err := mp.PeopleSet(context.Background(), []*mixpanel.PeopleProperties{{
			DistinctID:   user.UserId,
			Properties:   properties,
			UseRequestIp: false,
		}}); err != nil {
			logger.Error("PeopleSetError", zap.String("UserID", user.UserId), zap.Error(err))
			continue
		}
		logger.Info("UploadMixpanel", zap.Any("Index", index), zap.Any("IsStudent", user.IsStudent), zap.String("UserId", user.UserId), zap.Any("Property", properties))
	}

	fmt.Printf("UploadMixpanelFinished Count: %d\n", len(results))
	return nil
}
