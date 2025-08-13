package inputs

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/kuse-ai/kuse-insight-go/logger"
	"go.uber.org/zap"
)

type mixpanelUser struct {
	UserId      string `json:"user_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	LastSeen    string `json:"last_seen"`
	CountryCode string `json:"country_code"`
	Region      string `json:"region"`
	City        string `json:"city"`
	IsEducation string `json:"is_education"`
	Plan        string `json:"plan"`
}

func LoadMixpanelUsers() map[string]*mixpanelUser {
	filePath := "./sources/mixpanel/users.csv"
	users := make(map[string]*mixpanelUser)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Cannot load mixpanel users, err: ", err)
		return users
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		fmt.Println("Cannot read mixpanel users, err: ", err)
		return users
	}
	_ = header
	// logger.Info("MixPanelUserHeader", zap.Any("Header", header))

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println("Read mixpanel users record fail, err: ", err)
		}
		email := record[2]
		if email == "undefined" {
			continue
		}
		user := &mixpanelUser{}
		for columnIndex, value := range record {
			switch columnIndex {
			case 0:
				user.UserId = value
			case 1:
				user.Name = value
			case 2:
				user.Email = value
			case 3:
				user.LastSeen = value
			case 4:
				user.CountryCode = value
			case 5:
				user.Region = value
			case 6:
				user.City = value
			case 7:
				user.IsEducation = value
			case 8:
				user.Plan = value
			}
		}

		users[user.UserId] = user
	}
	logger.Info("MixPanelUsersDataLoadFinished", zap.Int("Count", len(users)))
	return users
}
