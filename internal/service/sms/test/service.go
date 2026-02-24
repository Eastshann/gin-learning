package test

import "context"

type service struct {
}

func (s service) Send(ctx context.Context, numbers, code string) error {
	return nil
}
