package core_pagination

import (
	"fmt"

	core_errors "github.com/Forestvov/checklist-service/internal/core/errors"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

type Params struct {
	Page    int
	PerPage int
}

func NewParams(page, perPage *int) (Params, error) {
	params := Params{
		Page:    DefaultPage,
		PerPage: DefaultPerPage,
	}

	if page != nil {
		params.Page = *page
	}

	if perPage != nil {
		params.PerPage = *perPage
	}

	if params.Page < 1 {
		return Params{}, fmt.Errorf(
			"page must be greater than zero: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	if params.PerPage < 1 || params.PerPage > MaxPerPage {
		return Params{}, fmt.Errorf(
			"per_page must be between 1 and %d: %w",
			MaxPerPage,
			core_errors.ErrInvalidArgument,
		)
	}

	return params, nil
}

func (p Params) Limit() int {
	return p.PerPage
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}
