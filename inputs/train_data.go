package inputs

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/kuse-ai/kuse-insight-go/logger"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

func LoadTrainDatas() []string {
	userIdMap := make(map[string]int)
	userIds := make([]string, 0)

	filePath := "./sources/google/user_tasks.xlsx"

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		fmt.Println("ReadXlsxFail, err: ", err)
		return userIds
	}
	defer f.Close()
	sheetName := f.GetSheetList()[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Println("GetXlsxRowsFail, err: ", err)
		return userIds
	}
	if len(rows) <= 0 {
		fmt.Println("SheetIsEmpty")
		return userIds
	}
	for rowIndex, row := range rows {
		if rowIndex == 0 { // header
			continue
		}
		for columnIndex, value := range row {
			switch columnIndex {
			case 0:
				if _, ok := userIdMap[value]; ok {
					continue
				}
				userIds = append(userIds, value)
				userIdMap[value] = 1
			}
		}
	}

	logger.Info("LoadTrainDatas", zap.Any("Count", len(userIds)))
	return userIds
}

func LoadProcessUserIds() []string {
	filePath := "./sources/big_query/users.csv"
	userIds := make([]string, 0)

	file, err := os.Open(filePath)
	if err != nil {
		return userIds
	}
	defer file.Close()

	reader := csv.NewReader(file)

	_, err = reader.Read()
	if err != nil {
		return userIds
	}

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
		}
		for columnIndex, value := range record {
			switch columnIndex {
			case 0:
				userIds = append(userIds, value)
			}
		}
	}
	logger.Info("LoadProcessUsers", zap.Int("Count", len(userIds)))
	return userIds
}
