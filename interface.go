package mongorepo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IRepository defines a generic interface for data access operations on a specific type `T`.
// This interface supports common CRUD operations (Create, Read, Update, Delete) for entities
// of type `T`, where `T` can be any struct representing a MongoDB document.
type IRepository[T any] interface {
	// Access the repository mongo.Collection ptr
	Collection() *mongo.Collection

	// FindById retrieves a single entity by its unique MongoDB ObjectID.
	//
	// Parameters:
	//   - id: The ObjectID of the entity to retrieve.
	//
	// Returns:
	//   - A pointer to the entity of type `T`, or nil if not found.
	//   - An error if the operation fails.
	FindById(id primitive.ObjectID) *T

	// FindOne executes a query to retrieve a single entity matching the provided search criteria.
	//
	// Parameters:
	//   - query: A BSON map defining the search criteria.
	//   - opts: Optional FindOneOptions to modify the query behavior.
	//
	// Returns:
	//   - A pointer to the entity of type `T`, or nil if no entity matches the criteria.
	//   - An error if the operation fails.
	FindOne(query bson.M, opts ...*options.FindOneOptions) *T

	// Find retrieves a list of entities that match the provided search criteria.
	//
	// Parameters:
	//   - query: A BSON map defining the search criteria.
	//   - opts: Optional FindOptions to modify the query behavior, such as sorting or pagination.
	//
	// Returns:
	//   - A slice of pointers to entities of type `T` that match the criteria.
	//   - An error if the operation fails.
	Find(query bson.M, opts ...*options.FindOptions) []*T

	// Create inserts a new entity into the MongoDB collection.
	//
	// Parameters:
	//   - entity: A pointer to the entity of type `T` to be inserted.
	//
	// Returns:
	//   - An error if the insertion fails.
	Create(entity *T) error

	// Update modifies an existing entity in the MongoDB collection.
	//
	// Parameters:
	//   - entity: A pointer to the entity of type `T` with updated fields.
	//
	// Returns:
	//   - An error if the update operation fails.
	Update(entity *T) error

	// Delete removes an entity from the MongoDB collection by its ObjectID.
	//
	// Parameters:
	//   - entity: A pointer to the entity of type `T` to be deleted.
	//
	// Returns:
	//   - An error if the deletion fails.
	Delete(entity *T) error
}
