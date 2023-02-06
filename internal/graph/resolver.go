package graph

//go:generate go run github.com/99designs/gqlgen generate
import (
	"github.com/mvpratt/nodewatcher/internal/db"
	"github.com/mvpratt/nodewatcher/internal/graph/model"
)

// This file will not be regenerated automatically.

// Resolver serves as dependency injection for your app, add any dependencies you require here.
type Resolver struct {
	todos []*model.Todo
	nodes []*model.Node
	Node  db.NodeIF
	DB    db.NodewatcherDB
}
