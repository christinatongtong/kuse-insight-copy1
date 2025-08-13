package inputs

import (
	"context"
	"os"

	"github.com/kuse-ai/kuse-insight-go/logger"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"go.uber.org/zap"
)

type UserPineconeDocSchema struct {
	BoardId int64  `json:"board_id"`
	Text    string `json:"text"`
	FileId  int64  `json:"file_id"`
	UserId  int64  `json:"user_id"`
	Type    string `json:"type"`
}

type PineConeService struct {
	conn *pinecone.IndexConnection
}

func NewPineconeService() *PineConeService {
	host := os.Getenv("PINECONE_INDEX_HOST")
	namespace := os.Getenv("PINECONE_NAMESPACE")
	apiKey := os.Getenv("PINECONE_API_KEY")
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		logger.Error("CreatePineconeClientError", zap.Error(err))
		return nil
	}
	idxConnection, err := pc.Index(
		pinecone.NewIndexConnParams{
			Host:      host,
			Namespace: namespace,
		})
	if err != nil {
		logger.Error("CreatePineconeConnectionError", zap.Error(err))
		return nil
	}
	return &PineConeService{
		conn: idxConnection,
	}
}

func (svc *PineConeService) SeatchUserSummary(ctx context.Context, userId int64) []string {
	result := make([]string, 0)
	if userId == 0 {
		return result
	}
	metadataMap := map[string]interface{}{
		"user_id": map[string]interface{}{
			"$eq": userId,
		},
		"is_summary": map[string]interface{}{
			"$eq": 1,
		},
	}
	inputs := map[string]interface{}{
		"text": "",
	}
	zeros := make([]float32, 1536)
	res, err := svc.conn.SearchRecords(ctx, &pinecone.SearchRecordsRequest{
		Query: pinecone.SearchRecordsQuery{
			TopK:   10,
			Filter: &metadataMap,
			Vector: &pinecone.SearchRecordsVector{
				Values: &zeros,
			},
			Inputs: &inputs,
		},
		// Fields: &[]string{"chunk_text", "category"},
	})
	if err != nil {
		logger.Error("SearchUserPineconeRecordsError", zap.Int64("UserId", userId), zap.Error(err))
		return result
	}
	for _, doc := range res.Result.Hits {
		result = append(result, doc.Fields["text"].(string))
	}
	// logger.Info("SearchUserPineconeRecordsError", zap.Int64("UserId", userId), zap.Int("Count", len(result)))
	return result
}
