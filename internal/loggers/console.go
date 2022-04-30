package loggers

import (
	"fmt"
	"github.com/scientistnik/invest-agents/internal/app/domain"
	"time"
)

const (
	colorReset = "\033[0m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

type ConstructorConsoleLogger struct {
	Color bool
}

func (ccl ConstructorConsoleLogger) New(agentId int64) domain.Logger {
	return ConsoleLogger{agentId: agentId, color: ccl.Color}
}

type ConsoleLogger struct {
	agentId int64
	color   bool
}

type loggerLevels = string

const (
	INFO  loggerLevels = "INF"
	WARN  loggerLevels = "WRN"
	ERROR loggerLevels = "ERR"
	DEBUG loggerLevels = "DEB"
)

func (l ConsoleLogger) formatMessage(level loggerLevels, message string) {
	datetime := time.Now().Format(time.RFC3339)
	if l.color {
		datetime = colorGreen + datetime + colorReset
	}

	agent := fmt.Sprintf("%04d", l.agentId)
	if l.color {
		agent = colorBlue + "agent" + colorReset + ":" + colorYellow + agent + colorReset
	} else {
		agent = "agent:" + agent
	}

	if l.color {
		level = colorRed + level + colorReset
	}

	fmt.Printf("%s %s %s %s\n", datetime, agent, level, message)
}

func (l ConsoleLogger) Info(message string) {
	l.formatMessage(INFO, message)
}

func (l ConsoleLogger) Warn(message string) {
	l.formatMessage(WARN, message)
}

func (l ConsoleLogger) Error(message string) {
	l.formatMessage(ERROR, message)
}

func (l ConsoleLogger) Debug(message string) {
	l.formatMessage(DEBUG, message)
}
