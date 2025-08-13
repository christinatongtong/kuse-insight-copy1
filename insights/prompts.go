package insights

import (
	"fmt"

	"github.com/kuse-ai/kuse-insight-go/inputs"
)

const (
	USER_AVATAR_PROMTP = `
The picture given to you is an avatar, describe the content of this picture
	`
	USER_INSIGHT_SYSTEM_PROMPT = `
You are an AI assistant specialized in user profiling. Based on the following user profile and task records, analyze and infer the user's attributes and return the result in a structured JSON format.


### Requirements

1. For each attribute, return your output in **JSON format**.
2. For each attribute, include a "candidates" array. Each candidate must contain:
   - "value": the predicted value
   - "confidence": a float between 0 and 1 (your estimated likelihood)
   - "evidence": a short explanation of what input led you to this conclusion
3. If no information is available for an attribute, return a candidate with:
   - ""value": null", ""confidence": 1.0", and appropriate "evidence"

The JSON should include the following top-level attributes:
- "is_student": whether the use is student
- "student_info": nested fields:
  - "school": string
  - "major": string
  - "degree_level": string (e.g., Undergraduate, Master's, PhD)
- "occupation": user's job title or role
- "industry": user's work domain (e.g., finance, tech, education)
- "primary_language": language the user input, Simplified Chinese and Traditional Chinese defined as two different language
- "gender": predicted gender

Please analyze the user data below and return a single valid JSON object following the format above.
	`
	USER_INSIGHT_USER_PROMPT_TEMPLATE = `
## Input:

>User Base Profile:
%s

>The prompt that the user had input, detect the primary_language by following prompts:
%s

>The FileName that the user uploaded:
%s

>The FileSummary that user uploaded:
%s

Please reason carefully and return only the final JSON result, with no explanation or formatting outside the JSON.

Now, output the persona JSON:
	`
)

type Candidate struct {
	Confidence any    `json:"confidence"`
	Value      any    `json:"value"`
	Evidence   string `json:"evidence"`
}
type Candidates struct {
	Candidates []*Candidate `json:"candidates"`
}

type StudentInfo struct {
	School      *Candidates `json:"school"`
	Major       *Candidates `json:"Major"`
	DegreeLevel *Candidates `json:"degree_level"`
}

type UserPredictOutput struct {
	StudentInfo     *StudentInfo `json:"student_info"`
	IsStudent       *Candidates  `json:"is_student"`
	Occupation      *Candidates  `json:"occupation"`
	Industry        *Candidates  `json:"industry"`
	PrimaryLanguage *Candidates  `json:"primary_language"`
	Gender          *Candidates  `json:"gender"`
	Satisfaction    *Candidates  `json:"satisfaction"`
	// Tags             *Candidates `json:"tags"`
}

func GenPromptForUser(user *inputs.User) string {
	var userProfilePlaceholder string
	if user.MixpanelUser != nil {
		if user.MixpanelUser.Name != "" {
			userProfilePlaceholder += fmt.Sprintf("-Name: %s\n", user.MixpanelUser.Name)
		}
		if user.MixpanelUser.Email != "" {
			userProfilePlaceholder += fmt.Sprintf("-Email: %s\n", user.MixpanelUser.Email)
		}
		if user.MixpanelUser.CountryCode != "" {
			userProfilePlaceholder += fmt.Sprintf("-CountryCode: %s\n", user.MixpanelUser.CountryCode)
		}
		if user.MixpanelUser.Region != "" {
			userProfilePlaceholder += fmt.Sprintf("-Region: %s\n", user.MixpanelUser.Region)
		}
		if user.MixpanelUser.City != "" {
			userProfilePlaceholder += fmt.Sprintf("-City: %s\n", user.MixpanelUser.City)
		}
		if user.MixpanelUser.IsEducation == "true" {
			userProfilePlaceholder += "-IsEducation: True \n"
		}
	}
	if user.UserModel != nil {
		if user.UserModel.GivenName != "" {
			userProfilePlaceholder += fmt.Sprintf("-GivenName: %s\n", user.UserModel.GivenName)
		}
		if user.UserModel.FamilyName != "" {
			userProfilePlaceholder += fmt.Sprintf("-FamilyName: %s\n", user.UserModel.FamilyName)
		}
		if user.UserModel.FullName != "" {
			userProfilePlaceholder += fmt.Sprintf("-Fullname: %s\n", user.UserModel.FullName)
		}
		if user.UserModel.OutputLanguage != "" {
			userProfilePlaceholder += fmt.Sprintf("-SettingOutputLanguage: %s\n", user.UserModel.OutputLanguage)
		}
	}
	if user.AvatarInfo != "" {
		userProfilePlaceholder += fmt.Sprintf("-The Describe of User's Avatar Image: %s\n", user.AvatarInfo)
	}
	var userTaskPlaceholder string
	if user.TasksModel != nil {
		for _, task := range user.TasksModel {
			if task.TaskType == "" || task.TaskType == "doc_extraction" {
				continue
			}
			userTaskPlaceholder += fmt.Sprintf("- %s\n", task.TaskMeta.Prompt)
		}
	}
	// fmt.Println(userTaskPlaceholder)
	var userFileNamePlaceholder string
	if user.FilesModel != nil {
		for _, file := range user.FilesModel {
			userFileNamePlaceholder += fmt.Sprintf("- %s\n", file.FileName)
		}
	}
	var userFileSummaryPlaceholder string
	for _, summary := range user.Summary {
		userFileSummaryPlaceholder += fmt.Sprintf("- %s\n", summary)
	}

	return fmt.Sprintf(USER_INSIGHT_USER_PROMPT_TEMPLATE, userProfilePlaceholder, userTaskPlaceholder, userFileNamePlaceholder, userFileSummaryPlaceholder)
}
