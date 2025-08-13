package insights

import (
	"github.com/kuse-ai/kuse-insight-go/inputs"
)

// Train
// @Description: Local Train
func (insights *UserInsights) Train() {
	trainUserIds := inputs.LoadTrainDatas()
	insights.RunBatch(trainUserIds)
	insights.Save()
}
