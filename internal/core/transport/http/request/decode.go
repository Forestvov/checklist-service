package core_http_request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
	"github.com/go-playground/validator/v10"
)

var requestValidator = validator.New()

const MaxRequestBodySize int64 = 1 << 20

type validatable interface {
	Validate() error
}

func DecodeValidateRequest(r *http.Request, dest any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, MaxRequestBodySize+1))
	if err != nil {
		return fmt.Errorf(
			"read request body: %v: %w",
			err,
			core_errors.ErrInvalidArgument,
		)
	}
	if int64(len(body)) > MaxRequestBodySize {
		return fmt.Errorf(
			"request body exceeds the %d byte limit: %w",
			MaxRequestBodySize,
			core_errors.ErrInvalidArgument,
		)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return fmt.Errorf("request body is empty: %w", core_errors.ErrInvalidArgument)
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf(
			"decode JSON: %v: %w",
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("request body must contain a single JSON value")
		}

		return fmt.Errorf(
			"decode trailing JSON: %v: %w",
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	v, ok := dest.(validatable)
	if ok {
		err = v.Validate()
	} else {
		err = requestValidator.Struct(dest)
	}

	if err != nil {
		return fmt.Errorf(
			"request validation: %v: %w",
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return nil
}
