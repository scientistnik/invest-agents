package loggers

import (
	"fmt"
	"github.com/scientistnik/invest-agents/internal/app/domain"
	"time"
)

type ConstructorConsoleLogger struct{}

func (ccl ConstructorConsoleLogger) New(agentId int64) domain.Logger {
	return ConsoleLogger{agentId: agentId}
}

type ConsoleLogger struct {
	agentId int64
}

type loggerLevels = string

const (
	INFO  loggerLevels = "INF"
	WARN  loggerLevels = "WRN"
	ERROR loggerLevels = "ERR"
	DEBUG loggerLevels = "DEB"
)

func formatMessage(agentId int64, level loggerLevels, message string) {
	fmt.Println(time.Now().Format(time.RFC3339) + " agent:" + fmt.Sprintf("%4d", agentId) + " " + level + " " + message)
}

func (l ConsoleLogger) Info(message string) {
	formatMessage(l.agentId, INFO, message)
}

func (l ConsoleLogger) Warn(message string) {
	formatMessage(l.agentId, WARN, message)
}

func (l ConsoleLogger) Error(message string) {
	formatMessage(l.agentId, ERROR, message)
}

func (l ConsoleLogger) Debug(message string) {
	formatMessage(l.agentId, DEBUG, message)
}
