package outputs

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/kuse-ai/kuse-insight-go/tools"
	"github.com/kuse-ai/kuse-insight-go/types"
)

type UserCsvSaver struct{}

var _ types.IUserDataSaver = (*UserCsvSaver)(nil)

func (saver *UserCsvSaver) Save(results []*types.UserResult) error {
	file, err := os.Create("./results/results.csv")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"user_id",
		"email",
		"is_student",
		"school",
		"major",
		"degree_level",
		"occupation",
		"industry",
		"primary_language",
		"gender",
		"last_task_time",
		"is_guest_mode",
		"plan",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, userResult := range results {
		record := []string{
			userResult.UserId,
			tools.If(userResult.Email != "", userResult.Email, "-"),
			userResult.IsStudent,
			userResult.School,
			userResult.Major,
			userResult.DegreeLevel,
			userResult.Occupation,
			userResult.Industry,
			userResult.PrimaryLanguage,
			userResult.Gender,
			userResult.LastTaskTime,
			userResult.IsGuestMode,
			tools.If(userResult.Plan != "", userResult.Plan, "free"),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	fmt.Printf("UserResultCsvSaverFinished Count: %d\n", len(results))
	return nil
}
