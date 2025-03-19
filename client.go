package mongorepo

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewClient initializes and returns a *mongo.Client instance for connecting to MongoDB.
//
// If the connection fails, the function will panic.
//
// Parameters:
// - uri (string): The MongoDB connection URI.
// - ctx (optional, context.Context): A custom context for the connection. If not provided, `context.Background()` is used.
//
// Returns:
// - *mongo.Client: The connected MongoDB client instance.
//
// Example usage:
//
// client := mongorepo.NewClient("mongodb://localhost:27017/")          // Uses context.Background()
// client := mongorepo.NewClient("mongodb://localhost:27017/", ctx)     // Uses the provided context
func NewClient(uri string, ctx ...context.Context) *mongo.Client {
	// Use the provided context or default to context.Background()
	var c context.Context
	if len(ctx) > 0 && ctx[0] != nil {
		c = ctx[0]
	} else {
		c = context.Background()
	}

	client, err := mongo.Connect(c, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err.Error())
	}

	return client
}
