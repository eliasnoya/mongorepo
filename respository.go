package mongorepo

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository provides a generic implementation for data access operations on a specific type `T`.
// It utilizes MongoDB as the underlying database and supports CRUD operations with built-in reflection
// for dynamic field access and management of common fields like ID, CreatedAt, UpdatedAt, and DeletedAt.
type Repository[T any] struct {
	config *Config
}

// NewDefault initializes a new Repository instance with default configuration settings
// for a given MongoDB collection.
//
// Parameters:
//   - collection: A pointer to the MongoDB collection where the entities of type `T` are stored.
//
// Returns:
//   - A pointer to a newly created Repository instance.
func NewDefault[T any](collection *mongo.Collection) *Repository[T] {
	return New[T](&Config{
		Collection:     collection,
		IdField:        "ID",
		CreatedAtField: "CreatedAt",
		UpdatedAtField: "UpdatedAt",
		DeletedAtField: "DeletedAt",
	})
}

// NewRepository initializes a new Repository instance with the specified configuration.
// If not provided, it assigns default values to common field names like ID, CreatedAt, and UpdatedAt.
//
// Parameters:
//   - config: A pointer to a Config object containing the repository's settings.
//
// Returns:
//   - A pointer to a newly created Repository instance.
//
// Panics:
//   - If the MongoDB collection in the configuration is not set.
func New[T any](config *Config) *Repository[T] {
	if config.IdField == "" {
		config.IdField = "ID"
	}

	if config.Context == nil {
		config.Context = context.Background()
	}

	if config.Collection == nil {
		panic("Configuration error: The *mongo.Collection in the Collection property of mongorepo.Config is not set.")
	}

	return &Repository[T]{config: config}
}

// collection retrieves the MongoDB collection from the repository's configuration.
//
// Returns:
//   - A pointer to the MongoDB collection.
func (r *Repository[T]) collection() *mongo.Collection {
	return r.config.Collection
}

// FindById retrieves an entity by its unique MongoDB ObjectID.
// This method is a convenience wrapper around FindOne.
//
// Parameters:
//   - id: The ObjectID of the entity to retrieve.
//
// Returns:
//   - A pointer to the entity of type `T`, or nil if not found.
func (r *Repository[T]) FindById(id primitive.ObjectID) *T {
	return r.FindOne(bson.M{"_id": id})
}

// FindOne retrieves a single entity matching the provided query filter.
//
// Parameters:
//   - query: A BSON map defining the search criteria.
//   - opts: Optional FindOneOptions to modify the query behavior.
//
// Returns:
//   - A pointer to the entity of type `T`, or nil if no document matches the query.
func (r *Repository[T]) FindOne(query bson.M, opts ...*options.FindOneOptions) *T {
	var entity T

	err := r.collection().FindOne(r.config.Context, query, opts...).Decode(&entity)

	if err != nil {
		log.Printf("FindOne error: %s", err.Error())
		return nil
	}

	return &entity
}

// Find retrieves all entities matching the provided query filter.
//
// Parameters:
//   - query: A BSON map defining the search criteria.
//   - opts: Optional FindOptions to modify the query behavior (e.g., sorting, pagination).
//
// Returns:
//   - A slice of pointers to entities of type `T` that match the query, or nil if an error occurs.
func (r *Repository[T]) Find(query bson.M, opts ...*options.FindOptions) []*T {
	var entities []*T

	cursor, err := r.collection().Find(r.config.Context, query, opts...)
	if err != nil {
		log.Printf("Find error: %s", err.Error())
		return nil
	}

	if err := cursor.All(r.config.Context, &entities); err != nil {
		log.Printf("Find cursor error: %s", err.Error())
		return nil
	}

	return entities
}

// Create inserts a new entity into the MongoDB collection.
// The method automatically sets the ID and CreatedAt fields if they are present in the entity.
//
// Parameters:
//   - entity: A pointer to the entity of type `T` to be inserted.
//
// Returns:
//   - An error if the insertion fails.
func (r *Repository[T]) Create(entity *T) error {
	er := NewEntityReflection(r.config, entity)
	er.SetNewID()

	// only update CreatedAtField if is configured
	if r.config.CreatedAtField != "" {
		er.SetCreatedAt()
	}

	_, err := r.collection().InsertOne(r.config.Context, entity)
	return err
}

// Update modifies an existing entity in the MongoDB collection.
// The method automatically sets the UpdatedAt field to the current time before performing the update.
//
// Parameters:
//   - entity: A pointer to the entity of type `T` with updated data.
//
// Returns:
//   - An error if the update operation fails.
func (r *Repository[T]) Update(entity *T) error {
	er := NewEntityReflection(r.config, entity)

	// only update UpdatedAtField if is configured
	if r.config.UpdatedAtField != "" {
		er.SetUpdateAt()
	}

	_, err := r.collection().UpdateByID(r.config.Context, er.GetID(), bson.M{"$set": entity})
	return err
}

// Delete removes an entity from the MongoDB collection.
// If the configuration supports soft deletes, it sets the DeletedAt field instead of permanently deleting the document.
//
// Parameters:
//   - entity: A pointer to the entity of type `T` to be deleted.
//
// Returns:
//   - An error if the deletion fails.
func (r *Repository[T]) Delete(entity *T) error {
	er := NewEntityReflection(r.config, entity)

	// make update with timestamp over DeletedAtField if is set
	if r.config.DeletedAtField != "" {
		er.SetDeletedAt()
		return r.Update(entity)
	}

	_, err := r.collection().DeleteOne(r.config.Context, bson.M{"_id": er.GetID()})
	return err
}
