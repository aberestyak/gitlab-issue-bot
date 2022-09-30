package issue

import (
	"fmt"
	"strconv"
	"strings"

	gitlabUserAPI "aberestyak/gitlab-issue-bot/internal/gitlabAPI"
	utils "aberestyak/gitlab-issue-bot/utils"
	"github.com/xanzy/go-gitlab"
)

// BodySpec - issue itself spec
type BodySpec struct {
	Kind             string    `json:"object_kind"`
	User             Author    `json:"user"`
	ObjectAttributes Attibutes `json:"object_attributes"`
}

// GetUsersIDs - get gitlab IDs of all involved users
func (issueBody *BodySpec) GetUsersIDs() []int {
	return append(issueBody.getAssignee(), issueBody.getAuthor(), issueBody.getEditor())
}

// GetUsersNames - get gitlab usernames of all involved users
func (issueBody *BodySpec) GetUsersNames() []string {
	return issueBody.getMentionted()
}

func (issueBody *BodySpec) getMentionted() []string {
	var usernameList []string
	description := issueBody.ObjectAttributes.Description
	if strings.Contains(description, "@") {
		splitDescription := strings.SplitAfter(description, "@")
		username := strings.SplitAfter(splitDescription[1], " ")
		usernameList = append(usernameList, strings.Trim(username[0], " "))
	}
	return usernameList
}

func (issueBody *BodySpec) getAuthor() int {
	return issueBody.ObjectAttributes.IssueBodyAuthor
}

func (issueBody *BodySpec) getAssignee() []int {
	return issueBody.ObjectAttributes.Assignee
}

func (issueBody *BodySpec) getEditor() int {
	return issueBody.ObjectAttributes.UpdatedBy
}

// BeautifyNotification - generate beautiful markdown notification
func (issueBody *BodySpec) BeautifyNotification() string {
	var issueBodyBuilder strings.Builder
	issueID := strconv.Itoa(issueBody.ObjectAttributes.ID)
	issueBodyBuilder.Grow(32)
	switch issueBody.ObjectAttributes.Action {
	case "open":
		fmt.Fprintf(&issueBodyBuilder, "🆕 *New issue [\\#%s](%s)*\n", issueID, issueBody.ObjectAttributes.URL)
	case "update":
		fmt.Fprintf(&issueBodyBuilder, "👀 *Issue updated [\\#%s](%s)*\n", issueID, issueBody.ObjectAttributes.URL)
		fmt.Fprintf(&issueBodyBuilder, "*Updated by: * %s \n", issueBody.ObjectAttributes.UpdatedByName)
	case "close":
		fmt.Fprintf(&issueBodyBuilder, "🚫 *Issue closed [\\#%s](%s)*\n", issueID, issueBody.ObjectAttributes.URL)
	case "reopen":
		fmt.Fprintf(&issueBodyBuilder, "♾ *Issue reopened [\\#%s](%s)*\n", issueID, issueBody.ObjectAttributes.URL)
	}
	fmt.Fprintf(&issueBodyBuilder, "*Name*: %s\n", utils.SanitizeTelegramString(issueBody.ObjectAttributes.Title))
	fmt.Fprintf(&issueBodyBuilder, "*Creator*: %s\n", issueBody.ObjectAttributes.IssueBodyAuthorName)
	if len(issueBody.ObjectAttributes.AssigneeNames) > 0 {
		fmt.Fprintf(&issueBodyBuilder, "*Assignee*:\n")
		for _, AssigneeName := range issueBody.ObjectAttributes.AssigneeNames {
			fmt.Fprintf(&issueBodyBuilder, "  ◦ %s\n", AssigneeName)
		}
	}
	if len(issueBody.ObjectAttributes.Labels) > 0 {
		fmt.Fprintf(&issueBodyBuilder, "*Labels*:\n")
		for _, label := range issueBody.ObjectAttributes.Labels {
			fmt.Fprintf(&issueBodyBuilder, "  ◦ %s\n", utils.SanitizeTelegramString(label.Title))
		}
	}
	if issueBody.ObjectAttributes.Description != "" {
		fmt.Fprintf(&issueBodyBuilder, "*Description*: %s\n", utils.SanitizeTelegramString(issueBody.ObjectAttributes.Description))
	}
	return issueBodyBuilder.String()
}

// ConvIDsToNames - get users names from gitlab to print them instead of IDs
func (issueBody *BodySpec) ConvIDsToNames(gitlabClient *gitlab.Client) error {
	var err error
	for _, assignee := range issueBody.ObjectAttributes.Assignee {
		assingeeName, err := gitlabUserAPI.GetUserNameByID(assignee, gitlabClient)
		if err != nil {
			return err
		}
		issueBody.ObjectAttributes.AssigneeNames = append(issueBody.ObjectAttributes.AssigneeNames, assingeeName)
	}
	authorName, err := gitlabUserAPI.GetUserNameByID(issueBody.ObjectAttributes.IssueBodyAuthor, gitlabClient)
	if err != nil {
		return err
	}
	issueBody.ObjectAttributes.IssueBodyAuthorName = authorName
	updatedByName, err := gitlabUserAPI.GetUserNameByID(issueBody.ObjectAttributes.UpdatedBy, gitlabClient)
	if err != nil {
		return err
	}
	if len(updatedByName) > 0 {
		issueBody.ObjectAttributes.UpdatedByName = updatedByName
	}
	return nil
}
