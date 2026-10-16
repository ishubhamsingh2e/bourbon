package todos

import "context"

type TodoRepository struct{}

func NewTodoRepository() *TodoRepository {
	return &TodoRepository{}
}

func (r *TodoRepository) FindAll(ctx context.Context) ([]*Todo, error) {
	return []*Todo{}, nil
}

func (r *TodoRepository) FindByID(ctx context.Context, id string) (*Todo, error) {
	return &Todo{ID: id}, nil
}

func (r *TodoRepository) Save(ctx context.Context, item *Todo) error {
	return nil
}
