package inputs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"github.com/kuse-ai/kuse-insight-go/logger"
	"go.uber.org/zap"
)

type BigQuery struct {
	projectId string
}

func NewBigQuery(projId string) *BigQuery {
	return &BigQuery{
		projectId: projId,
	}
}

func (srv *BigQuery) getClient(ctx context.Context) *bigquery.Client {
	client, err := bigquery.NewClient(ctx, srv.projectId)
	if err != nil {
		panic(err)
	}
	return client
}

func (srv *BigQuery) GetUser(ctx context.Context, userId string) []*UserModel {
	client := srv.getClient(ctx)
	defer client.Close()

	users := make([]*UserModel, 0)
	if userId == "" {
		return users
	}
	sql := fmt.Sprintf(`
SELECT * FROM EXTERNAL_QUERY(
  "kuse-ai.us.kuse-ai-main",
  """
  SELECT 
    u.id,
    u.email, 
    u.status,
    u.given_name,
    u.family_name,
    u.image_url,
    u.output_language,
    u.full_name,
	u.updated_at
  FROM user AS u
  WHERE id = %s
  """
)
	`, userId)
	q := client.Query(sql)

	it, err := q.Read(ctx)
	if err != nil {
		logger.Error("QueryUserFailed", zap.String("UserId", userId), zap.Error(err))
		return users
	}

	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err != nil {
			break
		}
		user := &UserModel{}
		if row[0] != nil {
			user.UserId = fmt.Sprintf("%d", row[0].(int64))
			user.UserIdInt = row[0].(int64)
		}
		if row[1] != nil {
			user.Email = row[1].(string)
		}
		if row[2] != nil {
			user.Status = row[2].(string)
		}
		if row[3] != nil {
			user.GivenName = row[3].(string)
		}
		if row[4] != nil {
			user.FamilyName = row[4].(string)
		}
		if row[5] != nil {
			user.ImageUrl = row[5].(string)
		}
		if row[6] != nil {
			user.OutputLanguage = row[6].(string)
		}
		if row[7] != nil {
			user.FullName = row[7].(string)
		}
		if row[8] != nil {
			user.UpdateAt = row[8]
		}
		users = append(users, user)
	}
	return users
}

func (srv *BigQuery) GetTasks(ctx context.Context, userId string) (*civil.DateTime, []*TaskModel) {
	client := srv.getClient(ctx)
	defer client.Close()

	var lastTaskDate *civil.DateTime
	tasks := make([]*TaskModel, 0)
	if userId == "" {
		return lastTaskDate, tasks
	}
	sql := fmt.Sprintf(`
SELECT * FROM EXTERNAL_QUERY(
  "kuse-ai.us.kuse-ai-main",
  """
  SELECT 
	u.id,
    u.task_type,
	u.created_at,
    CONVERT(task_meta USING utf8) AS task_meta
  FROM tasks AS u
  WHERE u.task_type = 'communication'
  AND u.user_id = %s
  """
)
  ORDER BY created_at DESC;
	`, userId)
	q := client.Query(sql)

	it, err := q.Read(ctx)
	if err != nil {
		logger.Error("QueryTaskFailed", zap.String("UserId", userId), zap.Error(err))
		return lastTaskDate, tasks
	}

	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err != nil {
			break
		}
		if row[3] == nil {
			continue
		}
		if lastTaskDate == nil {
			if v, ok := row[2].(civil.DateTime); ok {
				lastTaskDate = &v
			}
		}
		taskMetaStr := row[3].(string)
		taskMeta := &TaskMeta{}
		json.Unmarshal([]byte(taskMetaStr), taskMeta)

		tasks = append(tasks, &TaskModel{
			TaskId:   fmt.Sprintf("%d", row[0].(int64)),
			TaskType: row[1].(string),
			TaskMeta: taskMeta,
		})
	}
	return lastTaskDate, tasks
}

func (srv *BigQuery) GetFiles(ctx context.Context, email string) []*FileModel {
	client := srv.getClient(ctx)
	defer client.Close()

	files := make([]*FileModel, 0)
	if email == "" {
		return files
	}
	sql := fmt.Sprintf(`
SELECT * FROM EXTERNAL_QUERY("kuse-ai.us.kuse-ai-main",
    """
    SELECT filename, filepath, md5
    FROM files f 
    JOIN user i ON f.user_id = i.id 
    WHERE i.email = '%s'
    ORDER BY f.created_at DESC
    """
)
	`, email)
	q := client.Query(sql)

	it, err := q.Read(ctx)
	if err != nil {
		logger.Error("QueryFilesFailed", zap.String("Email", email), zap.Error(err))
		return files
	}

	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err != nil {
			break
		}
		file := &FileModel{
			FileName: row[0].(string),
			FilePath: row[1].(string),
		}
		if row[2] != nil {
			file.Md5 = row[2].(string)
		}
		files = append(files, file)
	}
	return files
}

func (srv *BigQuery) GetFileMeta(email string) []*FileMetaModel {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute*5)
	defer cancel()

	client := srv.getClient(ctx)
	defer client.Close()

	files := make([]*FileMetaModel, 0)
	if email == "" {
		return files
	}
	sql := fmt.Sprintf(`
SELECT * FROM EXTERNAL_QUERY("kuse-ai.us.kuse-ai-main",
    """
    SELECT filename, filepath, md5
    FROM files f 
    JOIN user i ON f.user_id = i.id 
    WHERE i.email = '%s'
    ORDER BY f.created_at DESC
    """
)
	`, email)
	q := client.Query(sql)

	it, err := q.Read(ctx)
	if err != nil {
		logger.Error("QueryFileMetaFailed", zap.String("Email", email), zap.Error(err))
		return files
	}

	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err != nil {
			break
		}
		files = append(files, &FileMetaModel{})
	}
	return files
}
