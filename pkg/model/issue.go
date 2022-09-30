package issue

import (
	"errors"

	gitlabUserAPI "github.com/aberestyak/gitlab-issue-bot/internal/gitlabAPI"
	"github.com/xanzy/go-gitlab"
)

// Issue - global structure for issues and issue comments. Need to avoid reflections
type Issue struct {
	IssueBody *BodySpec
	IssueNote *NoteSpec
}

// Attibutes - issue attributes
type Attibutes struct {
	ID                  int `json:"iid"`
	IssueBodyAuthor     int `json:"author_id"`
	IssueBodyAuthorName string
	Assignee            []int `json:"assignee_ids"`
	AssigneeNames       []string
	UpdatedBy           int `json:"updated_by_id"`
	UpdatedByName       string
	State               string   `json:"state"`
	URL                 string   `json:"url"`
	Labels              []Labels `json:"labels"`
	Action              string   `json:"action"`
	Description         string   `json:"description"`
	Title               string   `json:"title"`
}

// Labels - issue labels
type Labels struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// Author - issue author
type Author struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// BotUser - gitlab/telegram user
type BotUser struct {
	Name       string
	TelegramID int
	GitlabID   int
}

func (issue *Issue) CreateUsersList(gitlabClient *gitlab.Client) ([]BotUser, error) {
	var botUsersList []BotUser
	if issue.IssueBody != nil {
		botUsers, err := makeUniqUsersList(issue.IssueBody.GetUsersIDs(), issue.IssueBody.GetUsersNames(), gitlabClient)
		if err != nil {
			return nil, err
		}
		botUsersList = botUsers
	} else if issue.IssueNote != nil {
		botUsers, err := makeUniqUsersList(issue.IssueNote.GetUsersIDs(), issue.IssueNote.GetUsersNames(), gitlabClient)
		if err != nil {
			return nil, err
		}
		botUsersList = botUsers
	} else {
		return nil, errors.New("Can't determine event type, nor issue or comment")
	}
	return botUsersList, nil
}

func makeUniqUsersList(gitlabUsersIDs []int, gitlabUsersNames []string, gitlabClient *gitlab.Client) ([]BotUser, error) {
	var usersList []BotUser
	for _, gitlabUserID := range gitlabUsersIDs {
		if gitlabUserID == 0 {
			continue
		}
		telegramID, err := gitlabUserAPI.GetTgIDByGitlabID(gitlabUserID, gitlabClient)
		if err != nil {
			return nil, err
		}
		name, err := gitlabUserAPI.GetUserNameByID(gitlabUserID, gitlabClient)
		if err != nil {
			return nil, err
		}
		usersList = appendUniq(BotUser{GitlabID: gitlabUserID, TelegramID: telegramID, Name: name}, usersList, telegramID)
	}
	for _, gitlabUserName := range gitlabUsersNames {
		telegramID, err := gitlabUserAPI.GetTgIDByGitlabUsername(gitlabUserName, gitlabClient)
		if err != nil {
			return nil, err
		}
		usersList = appendUniq(BotUser{TelegramID: telegramID, GitlabID: -1, Name: gitlabUserName}, usersList, telegramID)
	}
	return usersList, nil
}

// appendUniq - append only uniq users
func appendUniq(user BotUser, usersList []BotUser, key int) []BotUser {
	found := false
	for _, user := range usersList {
		if user.TelegramID == key {
			found = true
		}
	}
	if !found {
		usersList = append(usersList, user)
	}
	return usersList
}
