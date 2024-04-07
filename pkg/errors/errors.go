package errors

import (
	"fmt"
)

func LogAndReturnError(err error, message string, severity string, statusCode int) error {
	errorMessage := fmt.Sprintf("%s: %v", message, err)
	switch severity {
	case "DEBUG":
		fmt.Println(fmt.Sprintf(`{"message": "%s", "severity": "debug"}`, errorMessage))
	case "INFO":
		fmt.Println(fmt.Sprintf(`{"message": "%s", "severity": "info"}`, errorMessage))
	case "NOTICE":
		fmt.Println(fmt.Sprintf(`{"message": "%s", "severity": "notice"}`, errorMessage))
	case "WARNING":
		fmt.Println(fmt.Sprintf(`{"message": "%s", "severity": "warning"}`, errorMessage))
	case "ERROR":
		fmt.Println(fmt.Sprintf(`{"message": "%s", "severity": "error"}`, errorMessage))
	case "CRITICAL":
		fmt.Println(fmt.Sprintf(`{"message": "%s", "severity": "critical"}`, errorMessage))
	case "ALERT":
		fmt.Println(fmt.Sprintf(`{"message": "%s", "severity": "alert"}`, errorMessage))
	case "EMERGENCY":
		fmt.Println(fmt.Sprintf(`{"message": "%s", "severity": "emergency"}`, errorMessage))
	default:
		fmt.Println(fmt.Sprintf(`{"message": "%s", "severity": "default"}`, errorMessage))
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
