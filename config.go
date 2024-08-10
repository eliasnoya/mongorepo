package mongorepo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	Collection     *mongo.Collection
	Context        context.Context
	IdField        string
	DeletedAtField string
	CreatedAtField string
	UpdatedAtField string
}
