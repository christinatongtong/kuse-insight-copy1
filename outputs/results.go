package outputs

import (
	"encoding/csv"
	"os"

	"github.com/kuse-ai/kuse-insight-go/logger"
	"github.com/kuse-ai/kuse-insight-go/types"
	"go.uber.org/zap"
)

func LoadResults() map[string]*types.UserResult {
	filePath := "./results/results.csv"
	users := make(map[string]*types.UserResult)

	file, err := os.Open(filePath)
	if err != nil {
		return users
	}
	defer file.Close()

	reader := csv.NewReader(file)

	_, err = reader.Read()
	if err != nil {
		return users
	}

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
		}
		// email := record[2]
		// if email == "" || email == "undefined" {
		// 	continue
		// }
		user := &types.UserResult{}
		for columnIndex, value := range record {
			switch columnIndex {
			case 0:
				user.UserId = value
			case 1:
				user.Email = value
			case 2:
				user.IsStudent = value
			case 3:
				user.School = value
			case 4:
				user.Major = value
			case 5:
				user.DegreeLevel = value
			case 6:
				user.Occupation = value
			case 7:
				user.Industry = value
			case 8:
				user.PrimaryLanguage = value
			case 9:
				user.Gender = value
			case 10:
				user.LastTaskTime = value
			case 11:
				user.IsGuestMode = value
			case 12:
				user.Plan = value
			}
		}

		users[user.UserId] = user
	}
	logger.Info("LoadResults", zap.Int("Count", len(users)))
	return users
}
