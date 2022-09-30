package main

import (
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"

	"encoding/json"

	config "aberestyak/gitlab-issue-bot/internal/config"
	parser "aberestyak/gitlab-issue-bot/internal/parser"
	logger "aberestyak/gitlab-issue-bot/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	bot          *tb.Bot
	gitlabClient *gitlab.Client
	mainLogger   = log.WithFields(log.Fields{
		"component": "Main",
	})
)

func main() {
	logger.Init()
	botConfig := config.GetConfig()
	gitlabClient = config.InitGitlabClient(botConfig.GitlabToken, botConfig.GitlabURL)

	bot, _ = tb.NewBot(tb.Settings{
		Token:  botConfig.TelegramToken,
		Poller: &tb.LongPoller{Timeout: 5 * time.Second},
	})
	bot.Handle("/start", func(m *tb.Message) {
		mainLogger.Infof("User %s with ID %d has joined", m.Chat.Username, m.Chat.ID)
		if _, err := bot.Send(m.Chat, "You are now subscribed for issues updates!"); err != nil {
			log.Printf("Error while send message: %s", err.Error())
		}
	})
	go bot.Start()

	router := gin.New()
	router.Use(gin.LoggerWithFormatter(config.GinLogger))
	router.POST(botConfig.ListenLocation, handlingPOST)
	router.GET("/health/readiness", func(c *gin.Context) { c.Status(200) })
	router.GET("/health/liveness", func(c *gin.Context) { c.Status(200) })

	if err := router.Run(botConfig.ListenPort); err != nil {
		mainLogger.Fatalln(err.Error())
	}
}

// handlingPOST - handle POST requests
func handlingPOST(c *gin.Context) {
	var notification string
	var issueID int
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		mainLogger.Fatalf(err.Error())
	}
	issue, err := parser.ParseBody(body)
	if err != nil {
		mainLogger.Errorf(err.Error())
	}
	// Marshal only for debug
	issueByte, _ := json.MarshalIndent(issue, "", "    ")
	mainLogger.Debugf("Parsed webhook body: %s", string(issueByte))

	if issue.IssueBody != nil {
		if err := issue.IssueBody.ConvIDsToNames(gitlabClient); err != nil {
			mainLogger.Errorf("Can't get gitlab user names from IDs: %s", err.Error())
		}
		notification = issue.IssueBody.BeautifyNotification()
		issueID = issue.IssueBody.ObjectAttributes.ID
	} else if issue.IssueNote != nil {
		if err := issue.IssueNote.ConvIDsToNames(gitlabClient); err != nil {
			mainLogger.Errorf("Can't get gitlab user names from IDs: %s", err.Error())
		}
		notification = issue.IssueNote.BeautifyNotification()
		issueID = issue.IssueNote.Issue.ID
	} else {
		mainLogger.Errorln("Can't determine event type, nor issue or comment")
	}

	botUsers, err := issue.CreateUsersList(gitlabClient)
	if err != nil {
		mainLogger.Fatalf(err.Error())
	}

	for _, botUser := range botUsers {
		if botUser.TelegramID != 0 {
			user := &tb.User{ID: botUser.TelegramID}
			if _, err := bot.Send(user, string(notification), &tb.SendOptions{ParseMode: tb.ModeMarkdownV2}); err != nil {
				mainLogger.Errorf("Issue #%d. Error when sending notification to user %s: %s", issueID, botUser.Name, err)
			} else {
				mainLogger.Infof("Issue #%d. Notifaction was sent to user %s", issueID, botUser.Name)
			}
		} else {
			mainLogger.Infof("Issue #%d. Can't send notifaction sent to user %s", issueID, botUser.Name)
		}
	}
}
