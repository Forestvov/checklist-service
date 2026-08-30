package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func GetIntQueryParam(r *http.Request, key string) (*int, error) {
	param := r.URL.Query().Get(key)
	if param == "" {
		return nil, nil
	}

	val, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf(
			"param='%s' by key='%s' not a valid integer: %v: %w",
			param,
			key,
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return &val, nil
}

func GetBoolQueryParam(r *http.Request, key string) (*bool, error) {
	params, exists := r.URL.Query()[key]
	if !exists {
		return nil, nil
	}
	if len(params) != 1 || params[0] == "" {
		return nil, fmt.Errorf(
			"query parameter by key='%s' must contain one non-empty boolean value: %w",
			key,
			core_errors.ErrInvalidArgument,
		)
	}

	param := params[0]

	value, err := strconv.ParseBool(param)
	if err != nil {
		return nil, fmt.Errorf(
			"param='%s' by key='%s' not a valid boolean: %v: %w",
			param,
			key,
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return &value, nil
}

func GetDateQueryParam(r *http.Request, key string) (*time.Time, error) {
	param := r.URL.Query().Get(key)
	if param == "" {
		return nil, nil
	}

	layout := "2006-01-02"

	date, err := time.Parse(layout, param)
	if err != nil {
		return nil, fmt.Errorf(
			"param='%s' by key='%s' not a valid date: %v: %w",
			param,
			key,
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return &date, nil
}
