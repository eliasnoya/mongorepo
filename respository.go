package mongorepo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository provides a generic implementation for data access operations on a specific type `T`.
// It leverages MongoDB as the underlying database and uses reflection for dynamic field access.
type Repository[T any] struct {
	config *Config
}

func NewDefault[T any](collection *mongo.Collection) *Repository[T] {
	return New[T](&Config{
		Collection: collection,
	})
}

// NewRepository initializes a new Repository instance for the specified collection and configuration.
// It also sets default field names for the ID, CreatedAt, and UpdatedAt fields if not provided.
func New[T any](config *Config) *Repository[T] {
	if config.IdField == "" {
		config.IdField = "ID"
	}

	if config.CreatedAtField == "" {
		config.CreatedAtField = "CreatedAt"
	}

	if config.UpdatedAtField == "" {
		config.UpdatedAtField = "UpdatedAt"
	}

	if config.Context == nil {
		config.Context = context.Background()
	}

	if config.Collection == nil {
		panic("Configuration error: The *mongo.Collection in the Collection property of mongorepo.Config is not set.")
	}

	return &Repository[T]{config: config}
}

// collection retrieves the MongoDB collection from the repository configuration.
func (r *Repository[T]) collection() *mongo.Collection {
	return r.config.Collection
}

// setNewObjectID assigns a new ObjectID to the entity's ID field if it is not already set.
// The ID field must be of type primitive.ObjectID.
func (r *Repository[T]) setNewObjectID(entity *T) error {
	entityElem := reflect.ValueOf(entity).Elem()
	idField := entityElem.FieldByName(r.config.IdField)

	if idField.IsValid() && idField.CanSet() && idField.Type() == reflect.TypeOf(primitive.ObjectID{}) {
		idField.Set(reflect.ValueOf(primitive.NewObjectID()))
		return nil
	}

	errorStr := fmt.Sprintf("Error: ID field %q is either not found or cannot be set. Ensure it is defined as primitive.ObjectID", r.config.IdField)
	log.Println(errorStr)
	return errors.New(errorStr)
}

// getEntityObjectID retrieves the ObjectID from the entity's ID field.
// Returns primitive.NilObjectID if the field is not found or is not of type primitive.ObjectID.
func (r *Repository[T]) getEntityObjectID(entity *T) primitive.ObjectID {
	entityElem := reflect.ValueOf(entity).Elem()
	idField := entityElem.FieldByName(r.config.IdField)

	if !idField.IsValid() {
		log.Printf("Error: Field %q not found in entity. Check if %q is the correct field name in the entity struct.", r.config.IdField, r.config.IdField)
		return primitive.NilObjectID
	}

	if idField.Type() != reflect.TypeOf(primitive.ObjectID{}) {
		log.Printf("Error: Field %q in entity is not of type primitive.ObjectID. Actual type: %s", r.config.IdField, idField.Type().String())
		return primitive.NilObjectID
	}

	return idField.Interface().(primitive.ObjectID)
}

// setEntityTimestamp sets the current timestamp to the specified field in the entity.
// The field must be of type time.Time.
func (r *Repository[T]) setEntityTimestamp(entity *T, field string) {
	entityElem := reflect.ValueOf(entity).Elem()
	timeField := entityElem.FieldByName(field)

	if !timeField.IsValid() {
		log.Printf("Error: Field %q not found in entity. Ensure the field name is correct.", field)
		return
	}

	if timeField.Type() != reflect.TypeOf(time.Time{}) {
		log.Printf("Error: Field %q in entity is not of type time.Time. Actual type: %s", field, timeField.Type().String())
		return
	}

	timeField.Set(reflect.ValueOf(time.Now()))
}

// FindById retrieves an entity by its ObjectID. It is a convenience method for FindOne.
func (r *Repository[T]) FindById(id primitive.ObjectID) *T {
	return r.FindOne(bson.M{"_id": id})
}

// FindOne retrieves a single entity matching the provided query filter.
// Returns nil if no document is found or if an error occurs during retrieval.
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
// Returns a slice of pointers to the retrieved entities or nil if an error occurs.
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

// Create inserts a new entity into the database.
// Automatically sets the ID and CreatedAt fields if they are present in the entity struct.
func (r *Repository[T]) Create(entity *T) error {
	if err := r.setNewObjectID(entity); err != nil {
		return err
	}
	r.setEntityTimestamp(entity, r.config.CreatedAtField)

	_, err := r.collection().InsertOne(r.config.Context, entity)
	return err
}

// Update modifies an existing entity in the database.
// Automatically sets the UpdatedAt field to the current time before performing the update.
func (r *Repository[T]) Update(entity *T) error {
	r.setEntityTimestamp(entity, r.config.UpdatedAtField)

	_, err := r.collection().UpdateByID(r.config.Context, r.getEntityObjectID(entity), bson.M{"$set": entity})
	return err
}

// Delete removes an entity from the database.
// If soft delete is enabled, it sets the DeletedAt field instead of permanently deleting the document.
func (r *Repository[T]) Delete(entity *T) error {
	if r.config.DeletedAtField != "" {
		r.setEntityTimestamp(entity, r.config.DeletedAtField)
		return r.Update(entity)
	}

	_, err := r.collection().DeleteOne(r.config.Context, bson.M{"_id": r.getEntityObjectID(entity)})
	return err
}
