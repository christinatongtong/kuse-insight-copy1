package inputs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

func prettifyStruct(obj interface{}) string {
	bytes, _ := json.MarshalIndent(obj, "", "  ")
	return string(bytes)
}

func TestPinecone(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	var (
		indexHost = "documents-1ucbdkk.svc.gcp-us-central1-4a9f.pinecone.io"
		apiKey    = "00844b4c-14fc-4ab5-a700-cb90ad18b20e"
		namespace = "documents"
		userId    = 108
	)

	// prod
	indexHost = "documents-dso1lmi.svc.gcp-us-central1-4a9f.pinecone.io"
	apiKey = "pcsk_gjeH4_c1hkEiKeiHAWZZot2pDLdJES6HMcnzX1mqP8xSsEeYBEXmPxPQruHZ1tKnq8Xj"
	namespace = "documents"
	userId = 55308
	userId = 1841

	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create Client: %v", err)
	}

	idxConnection, err := pc.Index(
		pinecone.NewIndexConnParams{
			Host:      indexHost,
			Namespace: namespace,
		})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v", err)
	}

	metadataMap := map[string]interface{}{
		"user_id": map[string]interface{}{
			"$eq": userId,
		},
		"is_summary": map[string]interface{}{
			"$eq": 1,
		},
	}
	// inputs := map[string]interface{}{
	// 	"text": "",
	// }
	zeros := make([]float32, 1536)
	res, err := idxConnection.SearchRecords(ctx, &pinecone.SearchRecordsRequest{
		Query: pinecone.SearchRecordsQuery{
			TopK:   10,
			Filter: &metadataMap,
			// Inputs: &inputs,
			Vector: &pinecone.SearchRecordsVector{
				Values: &zeros,
			},
		},
		// Fields: &[]string{"chunk_text", "category"},
	})
	if err != nil {
		log.Fatalf("Failed to search records: %v, %v", err, res)
	}
	fmt.Printf("搜索结果数量: %d\n", len(res.Result.Hits))
	if len(res.Result.Hits) > 0 {
		fmt.Printf(prettifyStruct(res.Result.Hits[0]))
	}
}

func testBasicQuery(ctx context.Context, idx *pinecone.IndexConnection) {
	zeros := make([]float32, 1536)
	res, err := idx.SearchRecords(ctx, &pinecone.SearchRecordsRequest{
		Query: pinecone.SearchRecordsQuery{
			TopK: 5,
			Vector: &pinecone.SearchRecordsVector{
				Values: &zeros,
			},
		},
	})
	if err != nil {
		log.Printf("Basic query failed: %v", err)
		return
	}
	fmt.Printf("基本查询结果数量: %d\n", len(res.Result.Hits))
	if len(res.Result.Hits) > 0 {
		fmt.Printf(prettifyStruct(res.Result.Hits[0]))
	}
}

func TestFileMetaBigQuery(t *testing.T) {
	ctx := context.Background()
	projId := "kuse-ai"
	client, err := bigquery.NewClient(ctx, projId)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	sql := `
SELECT * FROM EXTERNAL_QUERY(
  "kuse-ai.us.kuse-ai-main",
  """
  SELECT 
    u.id,
    CONVERT(metadata USING utf8) AS metadata
  FROM file_meta AS u
  WHERE u.id = 285093
  """
)
	`
	q := client.Query(sql)

	it, err := q.Read(ctx)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err != nil {
			break
		}
		fmt.Println(row)
	}
}

func TestTasksBigQuery(t *testing.T) {
	ctx := context.Background()
	projId := "kuse-ai"
	client, err := bigquery.NewClient(ctx, projId)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
	sql := `
SELECT * FROM EXTERNAL_QUERY(
  "kuse-ai.us.kuse-ai-main",
  """
  SELECT 
    u.user_id,
    u.task_type,
	u.created_at,
    CONVERT(task_meta USING utf8) AS task_meta,
    CONVERT(result_meta USING utf8) AS result_meta
  FROM tasks AS u
  WHERE u.task_type = 'communication'
  AND u.user_id = 55907
  """
)
  ORDER BY created_at DESC;
	`
	q := client.Query(sql)

	it, err := q.Read(ctx)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	today := time.Now()
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err != nil {
			break
		}
		fmt.Println(row[2])
		if v, ok := row[2].(civil.DateTime); ok {
			date := v
			fmt.Println("Raw:", date, date.Date == civil.DateOf(today))
		} else {
			// 类型断言失败处理
			fmt.Println("unexpected type for created_at")
		}
	}
}
