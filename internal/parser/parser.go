package parser

import (
	"encoding/json"
	"errors"

	issue "aberestyak/gitlab-issue-bot/pkg/model"
	log "github.com/sirupsen/logrus"
)

var parserLogger = log.WithFields(log.Fields{
	"component": "Parser",
})

type issueType struct {
	Kind string `json:"object_kind"`
}

// ParseBody - parse http body with issue or issue comment
func ParseBody(body []byte) (issue.Issue, error) {
	issueKind := &issueType{}
	parserLogger.Debugf("Body: %s", string(body))
	if err := json.Unmarshal(body, issueKind); err != nil {
		return issue.Issue{}, err
	}

	generatedIssue := &issue.Issue{}
	switch issueKind.Kind {
	case "issue":
		issueBody := &issue.BodySpec{}
		if err := json.Unmarshal(body, issueBody); err != nil {
			return issue.Issue{}, err
		}
		generatedIssue.IssueBody = issueBody
	case "note":
		issueNote := &issue.NoteSpec{}
		if err := json.Unmarshal(body, issueNote); err != nil {
			return issue.Issue{}, err
		}
		generatedIssue.IssueNote = issueNote
	default:
		return issue.Issue{}, errors.New("Not issue/note")
	}
	return *generatedIssue, nil
}
