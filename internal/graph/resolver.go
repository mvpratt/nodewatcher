package graph

//go:generate go run github.com/99designs/gqlgen generate
import (
	"github.com/mvpratt/nodewatcher/internal/db"
)

// This file will not be regenerated automatically.

// Resolver serves as dependency injection for your app, add any dependencies you require here.
type Resolver struct {
	DB db.NodewatcherDB
}
