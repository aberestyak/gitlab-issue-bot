package config

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"

	log "github.com/sirupsen/logrus"
)

// BotConfig - telegram bot configuration
type BotConfig struct {
	ListenPort     string
	ListenLocation string
	TelegramToken  string
	GitlabToken    string
	GitlabURL      string
}

const (
	defaultListenPort     = ":8080"
	defaultListenLocation = "/"
	defaultGitlabURL      = "https://gitlab.com"
)

var (
	configLogger = log.WithFields(log.Fields{
		"component": "ConfigInit",
	})
)

// GetConfig - get environment variables for configuring bot
func GetConfig() BotConfig {
	config := BotConfig{}
	telegramToken, telegramTokenSet := os.LookupEnv("TELEGRAM_TOKEN")
	if !telegramTokenSet {
		configLogger.Fatalf("Environment variable TELEGRAM_TOKEN not set!")
	}
	config.TelegramToken = telegramToken

	gitlabToken, gitlabTokenSet := os.LookupEnv("GITLAB_TOKEN")
	if !gitlabTokenSet {
		configLogger.Fatalf("Environment variable GITLAB_TOKEN not set!")
	}
	config.GitlabToken = gitlabToken

	listenPort, portSet := os.LookupEnv("LISTEN_PORT")
	if !portSet {
		configLogger.Logger.Infof("Environment variable LISTEN_PORT not set, use default: %s", defaultListenPort)
		config.ListenPort = defaultListenPort
	} else {
		config.ListenPort = ":" + listenPort
	}

	gitlabURL, gitlabURLSet := os.LookupEnv("GITLAB_URL")
	if !gitlabURLSet {
		configLogger.Logger.Infof("Environment variable GITLAB_URL not set, use default: %s", defaultGitlabURL)
		config.GitlabURL = defaultGitlabURL
	} else {
		config.GitlabURL = gitlabURL
	}

	listenLocation, locationSet := os.LookupEnv("LISTEN_LOCATION")
	if !locationSet {
		configLogger.Logger.Infof("Environment variable LISTEN_LOCATION not set, use default: %s", defaultListenLocation)
		config.ListenLocation = defaultListenLocation
	} else {
		config.ListenLocation = listenLocation
	}
	return config
}

// InitGitlabClient - initialize gitlab client
func InitGitlabClient(token string, gitlabURL string) *gitlab.Client {
	gitlabClient, err := gitlab.NewClient(token, gitlab.WithBaseURL(gitlabURL))
	if err != nil {
		configLogger.Fatalf("Can't initialize gitlabClient: %s", err.Error())
	}
	return gitlabClient
}

// GinLogger - log only POST requests with logrus
func GinLogger(param gin.LogFormatterParams) string {
	ginLogger := log.WithFields(log.Fields{
		"component": "GIN",
	})
	if param.Method == "POST" {
		ginLogger.Debugf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}
	return ""
}
