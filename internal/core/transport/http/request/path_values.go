package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

func GetInt64PathValue(r *http.Request, key string) (int64, error) {
	pathValue := r.PathValue(key)
	if pathValue == "" {
		return 0, fmt.Errorf(
			"no key='%s' in path values: %w",
			key,
			core_errors.ErrInvalidArgument,
		)
	}

	value, err := strconv.ParseInt(pathValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"path value='%s' by key='%s' is not a valid int64: %v: %w",
			pathValue,
			key,
			err,
			core_errors.ErrInvalidArgument,
		)
	}
	if value <= 0 {
		return 0, fmt.Errorf(
			"path value='%d' by key='%s' must be greater than zero: %w",
			value,
			key,
			core_errors.ErrInvalidArgument,
		)
	}

	return value, nil
}
