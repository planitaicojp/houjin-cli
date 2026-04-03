package errors

import "fmt"

const (
	ExitOK         = 0
	ExitGeneral    = 1
	ExitConfig     = 2
	ExitAPI        = 3
	ExitValidation = 4
)

// ExitCoder is implemented by errors that carry a process exit code.
type ExitCoder interface {
	ExitCode() int
}

// APIError represents an error returned by the 法人番号 API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

func (e *APIError) ExitCode() int { return ExitAPI }

// ConfigError represents a configuration problem.
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error: %s", e.Message)
}

func (e *ConfigError) ExitCode() int { return ExitConfig }

// ValidationError represents invalid user input.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) ExitCode() int { return ExitValidation }

// GetExitCode returns the exit code for the given error.
func GetExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	if ec, ok := err.(ExitCoder); ok {
		return ec.ExitCode()
	}
	return ExitGeneral
}
