package core_http_response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	"go.uber.org/zap"
)

func TestHTTPResponseHandlerErrorResponse(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		wantStatus      int
		wantPublicError string
	}{
		{
			name:            "invalid argument",
			err:             fmt.Errorf("decode request: %w", core_errors.ErrInvalidArgument),
			wantStatus:      http.StatusBadRequest,
			wantPublicError: core_errors.ErrInvalidArgument.Error(),
		},
		{
			name:            "not found",
			err:             fmt.Errorf("get task: %w", core_errors.ErrNotFound),
			wantStatus:      http.StatusNotFound,
			wantPublicError: core_errors.ErrNotFound.Error(),
		},
		{
			name:            "conflict",
			err:             fmt.Errorf("create task: %w", core_errors.ErrConflict),
			wantStatus:      http.StatusConflict,
			wantPublicError: core_errors.ErrConflict.Error(),
		},
		{
			name:            "internal error is hidden",
			err:             errors.New("password=secret database connection failed"),
			wantStatus:      http.StatusInternalServerError,
			wantPublicError: http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler := NewHTTPResponseHandler(testLogger(), recorder)

			handler.ErrorResponse(tt.err, "request failed")

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, recorder.Code)
			}

			response := decodeErrorResponse(t, recorder)
			if response.Error != tt.wantPublicError {
				t.Errorf("expected public error %q, got %q", tt.wantPublicError, response.Error)
			}
			if response.Message != "request failed" {
				t.Errorf("expected message %q, got %q", "request failed", response.Message)
			}
		})
	}
}

func TestHTTPResponseHandlerPanicResponseHidesPanicValue(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := NewHTTPResponseHandler(testLogger(), recorder)

	handler.PanicResponse("password=secret", "unexpected server error")

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	response := decodeErrorResponse(t, recorder)
	if response.Error != http.StatusText(http.StatusInternalServerError) {
		t.Errorf(
			"expected public error %q, got %q",
			http.StatusText(http.StatusInternalServerError),
			response.Error,
		)
	}
}

func testLogger() *core_logger.Logger {
	return &core_logger.Logger{Logger: zap.NewNop()}
}

func decodeErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}

	return response
}
