package types

type UserResult struct {
	UserId          string `json:"user_id"`
	Email           string `json:"email"`
	IsStudent       string `json:"is_student"`
	Occupation      string `json:"occupation"`
	Industry        string `json:"industry"`
	PrimaryLanguage string `json:"primary_language"`
	Satisfaction    string `json:"satisfaction"`
	Gender          string `json:"gender"`
	School          string `json:"school"`
	Major           string `json:"Major"`
	DegreeLevel     string `json:"degree_level"`

	IsGuestMode  string `json:"is_guest_mode"`
	Plan         string `json:"plan"`
	LastTaskTime string `json:"last_task_time"`
}

type IUserDataSaver interface {
	Save([]*UserResult) error
}
