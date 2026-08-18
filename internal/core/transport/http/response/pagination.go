package core_http_response

import core_pagination "github.com/Forestvov/checklist-service/internal/core/pagination"

type PaginationMeta struct {
	Page       int   `json:"page" example:"2"`
	PerPage    int   `json:"per_page" example:"20"`
	Total      int64 `json:"total" example:"47"`
	TotalPages int64 `json:"total_pages" example:"3"`
}

func PaginationMetaFromResult[T any](
	result core_pagination.Result[T],
) PaginationMeta {
	return PaginationMeta{
		Page:       result.Params.Page,
		PerPage:    result.Params.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages(),
	}
}
