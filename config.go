package mongorepo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config holds the configuration necessary for connecting and interacting with a MongoDB collection.
type Config struct {
	MongoClient       *mongo.Client              // The MongoDB client instance used for database connections.
	DatabaseOptions   *options.DatabaseOptions   // The MongoDb Database options, default: nil
	CollectionOptions *options.CollectionOptions // The MongoDb Collection options, default: nil
	DbName            string                     // The name of the database where the collection resides.
	CollectionName    string                     // The name of the collection representing the entity.
	Context           context.Context            // The context to manage request lifecycle (e.g., timeouts, cancellations) during MongoDB operations.
	IdField           string                     // The field in the entity struct that represents the "_id" field in MongoDB, which must be a primitive.ObjectID.
	DeletedAtField    string                     // The field in the entity struct to track soft deletes, indicating when a document is marked as deleted.
	CreatedAtField    string                     // The field in the entity struct to store the timestamp of when the document was created; must be of type time.Time.
	UpdatedAtField    string                     // The field in the entity struct to store the timestamp of when the document was last updated; must be of type time.Time.
}
