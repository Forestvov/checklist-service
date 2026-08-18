package core_pagination

type Result[T any] struct {
	Items  []T
	Total  int64
	Params Params
}

func NewResult[T any](
	items []T,
	total int64,
	params Params,
) Result[T] {
	if items == nil {
		items = make([]T, 0)
	}

	return Result[T]{
		Items:  items,
		Total:  total,
		Params: params,
	}
}

func (r Result[T]) TotalPages() int64 {
	if r.Total <= 0 || r.Params.PerPage <= 0 {
		return 0
	}

	perPage := int64(r.Params.PerPage)
	totalPages := r.Total / perPage

	if r.Total%perPage != 0 {
		totalPages++
	}

	return totalPages
}
