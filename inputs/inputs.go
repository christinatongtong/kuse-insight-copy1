package inputs

import (
	"errors"

	"cloud.google.com/go/civil"
	"golang.org/x/net/context"
)

type User struct {
	Email        string          `json:"email"`
	UserId       string          `json:"user_id"`
	UserIdInt    int64           `json:"user_id_int"`
	AvatarInfo   string          `json:"avatar_info"`
	LastTaskDate *civil.DateTime `json:"last_task_date"`
	IsGuestMode  bool            `json:"is_guest_mode"`

	MixpanelUser *mixpanelUser `json:"mixpanel_user"`
	UserModel    *UserModel    `json:"user_model"`
	TasksModel   []*TaskModel  `json:"tasks_model"`
	FilesModel   []*FileModel  `json:"files_model"`
	Summary      []string      `json:"summary"`
}

func (user *User) LastTaskTime() string {
	if user.LastTaskDate == nil {
		return ""
	}
	return user.LastTaskDate.String()
}

func (user *User) Plan() string {
	if user.MixpanelUser == nil {
		return "free"
	}
	if user.MixpanelUser.Plan == "undefined" {
		return "free"
	}
	return user.MixpanelUser.Plan
}

type Inputs struct {
	gclient  *BigQuery
	pinecone *PineConeService
	users    map[string]*mixpanelUser
}

func NewInputs() *Inputs {
	return &Inputs{
		gclient:  NewBigQuery("kuse-ai"),
		pinecone: NewPineconeService(),
		users:    make(map[string]*mixpanelUser),
	}
}

func (inputs *Inputs) Load() {
	inputs.users = LoadMixpanelUsers()
}

func (inputs *Inputs) UserIds() []string {
	userIds := LoadProcessUserIds()
	return userIds
	// userIds := make([]string, 0)
	// for userId := range inputs.users {
	// 	userIds = append(userIds, userId)
	// }
	// return userIds
}

func (inputs *Inputs) Get(ctx context.Context, userId string) (*User, error) {
	isGuestMode := false
	mixpanelUser := inputs.users[userId]
	if mixpanelUser == nil {
		isGuestMode = true // mixpanel上找不到这个用户就把他划分为guest
	}
	// user information from database
	userModels := inputs.gclient.GetUser(ctx, userId)
	if len(userModels) == 0 {
		return nil, errors.New("not found in db")
	}
	user := &User{
		UserId:       userId,
		UserIdInt:    userModels[0].UserIdInt,
		Email:        userModels[0].Email,
		IsGuestMode:  isGuestMode,
		MixpanelUser: mixpanelUser,
	}
	user.UserModel = userModels[0]
	user.LastTaskDate, user.TasksModel = inputs.gclient.GetTasks(ctx, userId)
	user.FilesModel = inputs.gclient.GetFiles(ctx, user.Email)
	user.Summary = inputs.pinecone.SeatchUserSummary(ctx, user.UserIdInt)
	return user, nil
}
