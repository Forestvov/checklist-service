package core_http_middleware

import (
	"net/http"
	"strings"
	"time"

	core_logger "github.com/Forestvov/checklist-service/internal/core/logger"
	core_http_response "github.com/Forestvov/checklist-service/internal/core/transport/http/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

const (
	allowedCORSMethods = "GET, HEAD, POST, PATCH, DELETE, OPTIONS"
	allowedCORSHeaders = "Content-Type, Authorization, X-Request-ID"
)

func CORS(allowedOriginsList []string) Middleware {
	allowedOrigins := make(map[string]struct{})
	for _, origin := range allowedOriginsList {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowedOrigins[origin] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("Vary", "Origin")

			responseOrigin := origin
			_, originAllowed := allowedOrigins[origin]
			if _, wildcardAllowed := allowedOrigins["*"]; wildcardAllowed {
				originAllowed = true
				responseOrigin = "*"
			}

			isPreflight := r.Method == http.MethodOptions &&
				r.Header.Get("Access-Control-Request-Method") != ""
			if isPreflight {
				w.Header().Add("Vary", "Access-Control-Request-Method")
				w.Header().Add("Vary", "Access-Control-Request-Headers")
			}

			if !originAllowed ||
				(isPreflight && !isAllowedCORSMethod(r.Header.Get("Access-Control-Request-Method"))) ||
				(isPreflight && !areAllowedCORSHeaders(r.Header.Get("Access-Control-Request-Headers"))) {
				if isPreflight {
					http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", responseOrigin)
			w.Header().Set("Access-Control-Expose-Headers", requestIDHeader)

			if isPreflight {
				w.Header().Set("Access-Control-Allow-Methods", allowedCORSMethods)
				w.Header().Set("Access-Control-Allow-Headers", allowedCORSHeaders)
				w.Header().Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedCORSMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions:
		return true
	default:
		return false
	}
}

func areAllowedCORSHeaders(headers string) bool {
	for _, header := range strings.Split(headers, ",") {
		switch strings.ToLower(strings.TrimSpace(header)) {
		case "", "content-type", "authorization", "x-request-id":
			continue
		default:
			return false
		}
	}

	return true
}

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)

			if requestID == "" {
				requestID = uuid.NewString()
			}

			r.Header.Set(requestIDHeader, requestID)
			w.Header().Set(requestIDHeader, requestID)

			next.ServeHTTP(w, r)
		})
	}
}

func Logger(log *core_logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)

			logger := log.With(
				zap.String("request_id", requestID),
				zap.String("http_method", r.Method),
				zap.String("url", r.URL.String()),
			)

			ctx := core_logger.ToContext(r.Context(), logger)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := core_logger.FromContext(ctx)
			rw := core_http_response.NewResponseWriter(w)

			before := time.Now()

			logger.Debug(
				">>> incoming HTTP request",
				zap.Time("time", before.UTC()),
			)

			next.ServeHTTP(rw, r)

			logger.Debug(
				"<<< done HTTP request",
				zap.Int("status_code", rw.GetStatusCode()),
				zap.Duration("latency", time.Since(before)),
			)
		})
	}
}

func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := core_logger.FromContext(ctx)
			responseHandler := core_http_response.NewHTTPResponseHandler(logger, w)

			defer func() {
				if p := recover(); p != nil {
					responseHandler.PanicResponse(
						p,
						"during handle HTTP request got unexpected panic",
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
