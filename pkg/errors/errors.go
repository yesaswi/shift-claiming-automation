package errors

import (
	"fmt"

	"go.uber.org/zap"
)

func LogAndReturnError(log *zap.Logger, err error, message string, severity string, statusCode int) error {
	errorMessage := fmt.Sprintf("%s: %v", message, err)
	switch severity {
	case "DEBUG":
		log.Debug(fmt.Sprintf(`{"message": "%s", "severity": "debug"}`, errorMessage))
	case "INFO":
		log.Info(fmt.Sprintf(`{"message": "%s", "severity": "info"}`, errorMessage))
	case "NOTICE":
		log.Info(fmt.Sprintf(`{"message": "%s", "severity": "notice"}`, errorMessage))
	case "WARNING":
		log.Warn(fmt.Sprintf(`{"message": "%s", "severity": "warning"}`, errorMessage))
	case "ERROR":
		log.Error(fmt.Sprintf(`{"message": "%s", "severity": "error"}`, errorMessage))
	case "CRITICAL":
		log.Error(fmt.Sprintf(`{"message": "%s", "severity": "critical"}`, errorMessage))
	case "ALERT":
		log.Error(fmt.Sprintf(`{"message": "%s", "severity": "alert"}`, errorMessage))
	case "EMERGENCY":
		log.Error(fmt.Sprintf(`{"message": "%s", "severity": "emergency"}`, errorMessage))
	default:
		log.Error(fmt.Sprintf(`{"message": "%s", "severity": "default"}`, errorMessage))
	}
	return HTTPError{
		StatusCode: statusCode,
		Message:    errorMessage,
	}
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e HTTPError) Error() string {
	return e.Message
}
