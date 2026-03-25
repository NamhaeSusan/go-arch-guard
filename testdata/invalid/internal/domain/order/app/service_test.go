package app

import "context"

type mockOrderRepo struct {
	findByID func(ctx context.Context, id string) (string, error)
}

func (m *mockOrderRepo) FindByID(ctx context.Context, id string) (string, error) {
	return m.findByID(ctx, id)
}
