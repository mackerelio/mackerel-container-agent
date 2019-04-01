package ecsv3

import (
	"context"
	"errors"
)

type specGenerator struct{}

func newSpecGenerator() *specGenerator {
	return &specGenerator{}
}

func (g *specGenerator) Generate(ctx context.Context) (interface{}, error) {
	return nil, errors.New("not implemented yet")
}
