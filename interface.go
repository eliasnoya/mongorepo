package mongorepo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IRepository defines a generic interface for data access operations on a specific type `T`.
// This interface can be used with any type `T` to perform CRUD (Create, Read, Update, Delete)
type IRepository[T any] interface {
	// FindById retrieves a single entity by its ID (string).
	// Returns a pointer to the entity and an error if any occurs.
	FindById(id primitive.ObjectID) *T

	// FindOne executes a find operation using the provided search criteria (`db.TxParams`).
	// Returns a pointer to the found entity and an error if any occurs.
	FindOne(query bson.M, opts ...*options.FindOneOptions) *T

	// Find retrieves a list of entities matching the provided search criteria (`db.TxParams`).
	// Returns a slice of pointers to the found entities and an error if any occurs.
	Find(query bson.M, opts ...*options.FindOptions) []*T

	// Create persists a new entity in the database.
	// Requires a pointer to the entity object. Returns an error if any occurs during insertion.
	Create(entity *T) error

	// Update updates an existing entity in the database.
	// Requires a pointer to the modified entity object. Returns an error if any occurs during update.
	Update(entity *T) error

	// Delete removes an entity from the database by its ID (string or primitive.ObjectID).
	// Returns an error if any occurs during deletion.
	Delete(entity *T) error
}
