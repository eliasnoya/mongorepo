package mongorepo

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository defines a generic interface for a repository that interacts with MongoDB collections.
// It supports CRUD operations for entities of type T, along with other utility methods.
type IRepository[T any] interface {
	// SetDatabase sets the database name in the repository configuration and returns the updated repository.
	SetDatabase(name string) *Repository[T]

	// Collection returns the MongoDB collection instance associated with the repository.
	Collection() *mongo.Collection

	// Database returns the MongoDB database instance associated with the repository.
	Database() *mongo.Database

	// Aggregate performs an aggregation pipeline query on the collection.
	// It returns a cursor that can be iterated to access the resulting documents.
	Aggregate(pipeline *Pipe, opts ...*AggrOpts) (*mongo.Cursor, error)

	// FindById finds a document by its ID. The ID is provided as a string, and the document is returned
	// as a pointer to the entity of type T.
	FindById(id string) *T

	// Find performs a query with the provided filter and options, returning a slice of entities of type T.
	// It uses the FindOne method internally for querying multiple results.
	Find(find Find, opts ...*FindOpts) []*T

	// FindOne performs a query with the provided filter and options, returning a single entity of type T.
	// If no result is found, it returns nil.
	FindOne(find Find, opts ...*FindOneOpts) *T

	// Create inserts a new document into the collection representing the entity of type T.
	// If the entity is valid, it returns nil. It may also update the entity with the inserted values.
	Create(entity *T) error

	// Update updates an existing document in the collection based on the provided ID and entity of type T.
	// If the entity is valid, it returns nil, and the entity is updated with the new values.
	Update(id string, entity *T) error

	// Delete removes a document from the collection based on the provided ID.
	// If the 'soft' flag is true, it performs a soft delete by setting the "deleted_at" field.
	Delete(id string, soft bool) error
}
