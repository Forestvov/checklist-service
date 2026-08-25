package core_http_request

import (
	"fmt"
	"net/http"

	core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"
)

func GetPaginationParams(r *http.Request) (core_pagination.Params, error) {
	page, err := GetIntQueryParam(r, "page")
	if err != nil {
		return core_pagination.Params{}, fmt.Errorf(
			"get page query parameter: %w",
			err,
		)
	}

	perPage, err := GetIntQueryParam(r, "per_page")
	if err != nil {
		return core_pagination.Params{}, fmt.Errorf(
			"get per_page query parameter: %w",
			err,
		)
	}

	params, err := core_pagination.NewParams(page, perPage)
	if err != nil {
		return core_pagination.Params{}, fmt.Errorf(
			"validate pagination parameters: %w",
			err,
		)
	}

	return params, nil
}
