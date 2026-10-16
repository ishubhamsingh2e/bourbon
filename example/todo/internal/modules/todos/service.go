package todos

import "context"

type TodoService struct {
	repo *TodoRepository
}

func NewTodoService(r *TodoRepository) *TodoService {
	if r == nil {
		r = NewTodoRepository()
	}
	return &TodoService{repo: r}
}

func (s *TodoService) List(ctx context.Context) ([]*Todo, error) {
	return s.repo.FindAll(ctx)
}

func (s *TodoService) GetByID(ctx context.Context, id string) (*Todo, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *TodoService) Create(ctx context.Context, item *Todo) error {
	return s.repo.Save(ctx, item)
}
