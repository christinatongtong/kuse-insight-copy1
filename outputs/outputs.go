package outputs

import (
	"sync"

	"github.com/kuse-ai/kuse-insight-go/logger"
	"github.com/kuse-ai/kuse-insight-go/types"
	"go.uber.org/zap"
)

type Outputs struct {
	results map[string]*types.UserResult
	savers  []types.IUserDataSaver
	sync.RWMutex
}

func NewOutputs() *Outputs {
	return &Outputs{
		results: make(map[string]*types.UserResult),
		savers: []types.IUserDataSaver{
			&UserCsvSaver{},
		},
	}
}

func (outputs *Outputs) Add(result *types.UserResult) {
	outputs.Lock()
	defer outputs.Unlock()
	outputs.results[result.UserId] = result
}

func (outputs *Outputs) Get(userId string) *types.UserResult {
	outputs.RLock()
	defer outputs.RUnlock()
	return outputs.results[userId]
}

func (outputs *Outputs) Results() []*types.UserResult {
	results := make([]*types.UserResult, 0)
	for _, result := range outputs.results {
		results = append(results, result)
	}
	return results
}

func (outputs *Outputs) Save() {
	results := outputs.Results()
	if len(results) == 0 {
		return
	}
	for _, saver := range outputs.savers {
		saver.Save(results)
	}
	logger.Info("UserResultSaved", zap.Int("AllCount", len(results)))
}

func (outputs *Outputs) Load() {
	outputs.results = LoadResults()
}

func (outputs *Outputs) Upload() {
	UploadMixpanel(outputs.Results())
}
