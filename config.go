package mongorepo

import "go.mongodb.org/mongo-driver/mongo"

type RepositoryConfig struct {
	Collection     *mongo.Collection
	IdField        string
	DeletedAtField string
	CreatedAtField string
	UpdatedAtField string
}
