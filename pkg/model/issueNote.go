package issue

import (
	"fmt"
	"strconv"
	"strings"

	gitlabUserAPI "aberestyak/gitlab-issue-bot/internal/gitlabAPI"
	utils "aberestyak/gitlab-issue-bot/utils"
	"github.com/xanzy/go-gitlab"
)

// NoteSpec - issue commect spec
type NoteSpec struct {
	Kind             string         `json:"object_kind"`
	User             Author         `json:"user"`
	ObjectAttributes NotesAttibutes `json:"object_attributes"`
	Issue            Attibutes      `json:"issue"`
}

// NotesAttibutes - issue comment attributes
type NotesAttibutes struct {
	Note        string `json:"note"`
	Description string `json:"description"`
	URL         string `json:"URL"`
}

// GetUsersIDs - get gitlab IDs of all involved users
func (issueNote *NoteSpec) GetUsersIDs() []int {
	return append(issueNote.getAssignee(), issueNote.getAuthor())
}

// GetUsersNames - get gitlab usernames of all involved users
func (issueNote *NoteSpec) GetUsersNames() []string {
	return issueNote.getMentionted()
}

func (issueNote *NoteSpec) getMentionted() []string {
	var usernameList []string
	noteText := issueNote.ObjectAttributes.Note
	if strings.Contains(noteText, "@") {
		splitNote := strings.SplitAfter(noteText, "@")
		username := strings.SplitAfter(splitNote[1], " ")
		usernameList = append(usernameList, strings.Trim(username[0], " "))
	}
	return usernameList
}

func (issueNote *NoteSpec) getAuthor() int {
	return issueNote.Issue.IssueBodyAuthor
}

func (issueNote *NoteSpec) getAssignee() []int {
	return issueNote.Issue.Assignee
}

// BeautifyNotification - generate beautiful markdown notification
func (issueNote *NoteSpec) BeautifyNotification() string {
	var noteTextBuilder strings.Builder
	issueID := strconv.Itoa(issueNote.Issue.ID)
	noteTextBuilder.Grow(32)
	fmt.Fprintf(&noteTextBuilder, "💬 *New comment in [\\#%s](%s)*\n", issueID, issueNote.Issue.URL)
	fmt.Fprintf(&noteTextBuilder, "*Issue*:\n")
	fmt.Fprintf(&noteTextBuilder, "*  Name*: %s\n", utils.SanitizeTelegramString(issueNote.Issue.Title))
	fmt.Fprintf(&noteTextBuilder, "*  Creator*: %s\n", issueNote.Issue.IssueBodyAuthorName)
	if len(issueNote.Issue.AssigneeNames) > 0 {
		fmt.Fprintf(&noteTextBuilder, "*  Assignee*:\n")
		for _, AssigneeName := range issueNote.Issue.AssigneeNames {
			fmt.Fprintf(&noteTextBuilder, "    ◦ %s\n", AssigneeName)
		}
	}
	if len(issueNote.Issue.Labels) > 0 {
		fmt.Fprintf(&noteTextBuilder, "*  Labels*:\n")
		for _, label := range issueNote.Issue.Labels {
			fmt.Fprintf(&noteTextBuilder, "    ◦ %s\n", utils.SanitizeTelegramString(label.Title))
		}
	}
	fmt.Fprintf(&noteTextBuilder, "*Comment author*: %s\n", issueNote.User.Name)
	fmt.Fprintf(&noteTextBuilder, "*Comment*: %s\n", utils.SanitizeTelegramString(issueNote.ObjectAttributes.Description))

	return noteTextBuilder.String()
}

// ConvIDsToNames - get users names from gitlab to print them instead of IDs
func (issueNote *NoteSpec) ConvIDsToNames(gitlabClient *gitlab.Client) error {
	var err error
	for _, assignee := range issueNote.Issue.Assignee {
		assingeeName, err := gitlabUserAPI.GetUserNameByID(assignee, gitlabClient)
		if err != nil {
			return err
		}
		issueNote.Issue.AssigneeNames = append(issueNote.Issue.AssigneeNames, assingeeName)
	}
	authorName, err := gitlabUserAPI.GetUserNameByID(issueNote.Issue.IssueBodyAuthor, gitlabClient)
	if err != nil {
		return err
	}
	issueNote.Issue.IssueBodyAuthorName = authorName
	updatedByName, err := gitlabUserAPI.GetUserNameByID(issueNote.Issue.UpdatedBy, gitlabClient)
	if err != nil {
		return err
	}
	issueNote.Issue.UpdatedByName = updatedByName
	return nil
}
