package core_http_response

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	"go.uber.org/zap"
)

type HTTPResponseHandler struct {
	log *core_logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponseHandler(log *core_logger.Logger, rw http.ResponseWriter) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		log: log,
		rw:  rw,
	}
}

func (h *HTTPResponseHandler) JSONResponse(
	responseBody any,
	statusCode int,
) {
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(responseBody); err != nil {
		h.log.Error("encode HTTP response", zap.Error(err))
		h.writeEncodingErrorResponse()
		return
	}

	h.rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	h.rw.WriteHeader(statusCode)

	if _, err := h.rw.Write(body.Bytes()); err != nil {
		h.log.Error("write HTTP response body", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) writeEncodingErrorResponse() {
	h.rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	h.rw.WriteHeader(http.StatusInternalServerError)

	body := fmt.Sprintf(
		"{\"error\":%q,\"message\":%q}\n",
		http.StatusText(http.StatusInternalServerError),
		"failed to encode response",
	)
	if _, err := h.rw.Write([]byte(body)); err != nil {
		h.log.Error("write HTTP encoding error response", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) errorResponse(
	statusCode int,
	publicError string,
	msg string,
) {
	response := ErrorResponse{
		Message: msg,
		Error:   publicError,
	}

	h.JSONResponse(
		response,
		statusCode,
	)
}

func (h *HTTPResponseHandler) PanicResponse(p any, msg string) {
	statusCode := http.StatusInternalServerError
	err := fmt.Errorf("unexpected panic: %v", p)

	h.log.Error(msg, zap.Error(err))

	h.errorResponse(
		statusCode,
		http.StatusText(statusCode),
		msg,
	)
}

func (h *HTTPResponseHandler) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

func (h *HTTPResponseHandler) ErrorResponse(err error, msg string) {
	var (
		statusCode  int
		publicError string
		logFunc     func(string, ...zap.Field)
	)

	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		publicError = core_errors.ErrInvalidArgument.Error()
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrNotFound):
		statusCode = http.StatusNotFound
		publicError = core_errors.ErrNotFound.Error()
		logFunc = h.log.Debug
	case errors.Is(err, core_errors.ErrConflict):
		statusCode = http.StatusConflict
		publicError = core_errors.ErrConflict.Error()
		logFunc = h.log.Warn
	default:
		statusCode = http.StatusInternalServerError
		publicError = http.StatusText(statusCode)
		logFunc = h.log.Error
	}

	logFunc(msg, zap.Error(err))

	h.errorResponse(
		statusCode,
		publicError,
		msg,
	)
}
