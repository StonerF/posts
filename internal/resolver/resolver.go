package resolver

//go:generate go run github.com/99designs/gqlgen generate

import (
	"sync"

	"github.com/StonerF/posts/internal/model"
	"github.com/StonerF/posts/internal/storage"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Repo             storage.Repository
	CommentObservers map[string]chan *model.Comment
	mu               sync.Mutex
}

func NewResolver(Repo storage.Repository) *Resolver {
	return &Resolver{
		Repo:             Repo,
		CommentObservers: make(map[string]chan *model.Comment),
	}
}
