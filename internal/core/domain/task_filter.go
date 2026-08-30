package core_domain

type TaskFilter struct {
	Done *bool
}

func NewTaskFilter(done *bool) TaskFilter {
	return TaskFilter{
		Done: done,
	}
}
