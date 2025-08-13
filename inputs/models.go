package inputs

type UserModel struct {
	UserId         string `json:"id"`
	UserIdInt      int64  `json:"user_id"`
	Email          string `json:"email"`
	Status         string `json:"status"`
	GivenName      string `json:"given_name"`
	FamilyName     string `json:"family_name"`
	ImageUrl       string `json:"image_url"`
	OutputLanguage string `json:"output_language"`
	FullName       string `json:"full_name"`
	UpdateAt       any    `json:"update_at"`
}

type TaskMeta struct {
	Prompt  string  `json:"prompt"`
	FileIds []int64 `json:"file_ids"`
}

type TaskModel struct {
	TaskId   string    `json:"task_id"`
	TaskType string    `json:"task_type"`
	TaskMeta *TaskMeta `json:"task_meta"`
}

type FileMetaModel struct {
	FileId   string `json:"file_id"`
	MetaData string `json:"meta_data"`
}
type FileModel struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Md5      string `json:"md5"`
}
