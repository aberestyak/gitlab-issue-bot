package gitlabuserapi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

var (
	gitlabAPILogger = log.WithFields(log.Fields{
		"component": "gitlabAPI",
	})
)

const (
	telegramIDKey = "Telegram_ID"
)

// GetTgIDByGitlabID - get user telegram ID from gitlab BIO by user ID
func GetTgIDByGitlabID(gitlabUserID int, gitlabClient *gitlab.Client) (int, error) {
	var telegramUserID int
	// Check if there is no ID
	if gitlabUserID == 0 {
		return 0, nil
	}
	gitlabUser, _, err := gitlabClient.Users.GetUser(gitlabUserID, gitlab.GetUsersOptions{}, nil)
	if err != nil {
		gitlabAPILogger.Errorf("Error when trying GetUser: %s", err.Error())
		return 0, err
	}
	if strings.Contains(gitlabUser.Bio, telegramIDKey) {
		telegramUserID, err = getTelegramIDFromBIO(gitlabUser.Bio)
		if err != nil {
			return 0, fmt.Errorf("%s for user with gitlab id %d", err.Error(), gitlabUserID)
		}
	} else {
		gitlabAPILogger.Warnf("Can't find user's Telegram_ID with gitlab ID %d", gitlabUserID)
	}
	return telegramUserID, nil
}

// GetTgIDByGitlabUsername - get user telegram ID from gitlab BIO by username
func GetTgIDByGitlabUsername(gitlabUsername string, gitlabClient *gitlab.Client) (int, error) {
	var telegramUserID int
	gitlabUser, _, err := gitlabClient.Users.ListUsers(&gitlab.ListUsersOptions{Username: &gitlabUsername}, nil)
	if err != nil {
		gitlabAPILogger.Errorf("Error when trying ListUsers: %s", err.Error())
		return 0, err
	}
	if len(gitlabUser) > 0 {
		if strings.Contains(gitlabUser[0].Bio, telegramIDKey) {
			telegramUserID, err = getTelegramIDFromBIO(gitlabUser[0].Bio)
			if err != nil {
				return 0, fmt.Errorf("%s for user with gitlab username %s", err.Error(), gitlabUsername)
			}
		} else {
			gitlabAPILogger.Warnf("Can't find Telegram_ID in user's %s BIO", gitlabUsername)
		}
	}
	return telegramUserID, nil
}

// GetUserNameByID - get gitlab user name by it's ID
func GetUserNameByID(gitlabUserID int, gitlabClient *gitlab.Client) (string, error) {
	// Check if there is no ID
	if gitlabUserID == 0 {
		return "", nil
	}
	gitlabUser, _, err := gitlabClient.Users.GetUser(gitlabUserID, gitlab.GetUsersOptions{}, nil)
	if err != nil {
		gitlabAPILogger.Errorf("Error when trying GetUser: %s", err.Error())
		return "", err
	}
	return gitlabUser.Name, nil
}

func getTelegramIDFromBIO(BIO string) (int, error) {
	var telegramUserID int
	// Replace, because users unterstood literally "<telegram_id>". Sigh
	numbersRegexp := regexp.MustCompile("[0-9]+")
	telegramUserIDString := numbersRegexp.FindAllString(strings.SplitAfter(BIO, telegramIDKey+": ")[1], -1)[0]
	telegramUserID, err := strconv.Atoi(telegramUserIDString)
	if err != nil {
		gitlabAPILogger.Errorf("Can't convert into integer value %s", strings.Split(telegramUserIDString, " ")[0])
		return 0, err
	}

	return telegramUserID, nil
}
