package errors_test

import (
	"fmt"
	"testing"

	cerrors "github.com/planitaicojp/houjin-cli/internal/errors"
)

func TestAPIError(t *testing.T) {
	err := &cerrors.APIError{StatusCode: 403, Message: "rate limited"}
	if err.Error() != "API error (HTTP 403): rate limited" {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err.ExitCode() != cerrors.ExitAPI {
		t.Errorf("expected exit code %d, got %d", cerrors.ExitAPI, err.ExitCode())
	}
}

func TestConfigError(t *testing.T) {
	err := &cerrors.ConfigError{Message: "app_id not set"}
	if err.Error() != "config error: app_id not set" {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err.ExitCode() != cerrors.ExitConfig {
		t.Errorf("expected exit code %d, got %d", cerrors.ExitConfig, err.ExitCode())
	}
}

func TestValidationError(t *testing.T) {
	err := &cerrors.ValidationError{Field: "corporate_number", Message: "invalid check digit"}
	if err.Error() != "validation error on corporate_number: invalid check digit" {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err.ExitCode() != cerrors.ExitValidation {
		t.Errorf("expected exit code %d, got %d", cerrors.ExitValidation, err.ExitCode())
	}
}

func TestValidationError_noField(t *testing.T) {
	err := &cerrors.ValidationError{Message: "bad input"}
	if err.Error() != "validation error: bad input" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestGetExitCode_nil(t *testing.T) {
	if cerrors.GetExitCode(nil) != cerrors.ExitOK {
		t.Error("expected ExitOK for nil error")
	}
}

func TestGetExitCode_generic(t *testing.T) {
	err := fmt.Errorf("generic")
	if cerrors.GetExitCode(err) != cerrors.ExitGeneral {
		t.Error("expected ExitGeneral for generic error")
	}
}
