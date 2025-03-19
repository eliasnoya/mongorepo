package mongorepo

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Aliases

// D is an alias for bson.D (a slice of BSON elements).
type D = bson.D

// Find is an alias for bson.M (a map of BSON elements used for queries).
type Find = bson.M

// FindOneOpts is an alias for options.FindOneOptions, used to customize the FindOne query.
type FindOneOpts = options.FindOneOptions

// FindOpts is an alias for options.FindOptions, used to customize the Find query.
type FindOpts = options.FindOptions

// Pipe is an alias for mongo.Pipeline, used for aggregation pipelines.
type Pipe = mongo.Pipeline

// AggrOpts is an alias for options.AggregateOptions, used to customize aggregation queries.
type AggrOpts = options.AggregateOptions

// ValidationResult holds the result of validating an entity.
type ValidationResult struct {
	Valid  bool                // True if the entity is valid, false otherwise.
	Errors map[string][]string // Map of field names to validation error messages.
}

// Config holds the configuration for the Repository, including MongoDB client, database, and collection.
type Config struct {
	Client     *mongo.Client   // The MongoDB client instance used for database connections.
	Database   string          // The name of the database where the collection resides.
	Collection string          // The name of the collection representing the entity.
	Context    context.Context // The context to manage request lifecycle (e.g., timeouts, cancellations) during MongoDB operations.
}

// Repository provides a generic implementation for data access operations on a specific type `T`.
// It utilizes MongoDB as the underlying database and supports CRUD operations with built-in reflection
// for dynamic field access and management of common fields like ID, CreatedAt, UpdatedAt, and DeletedAt.
type Repository[T any] struct {
	config *Config // Configuration for the repository, including MongoDB settings.
}

// New creates a new Repository instance with the given configuration.
// It performs validation on the configuration (ensuring context, client, and collection are set).
func New[T any](config *Config) *Repository[T] {
	if config.Context == nil {
		config.Context = context.Background()
	}

	if config.Client == nil {
		panic("Configuration error: The *mongo.Client is not set.")
	}

	// Detect collection name if not set
	if config.Collection == "" {
		panic("Configuration error: The Collection name is not set.")
	}

	return &Repository[T]{config: config}
}

// SetDatabase sets the database name for the repository.
func (r *Repository[T]) SetDatabase(name string) *Repository[T] {
	r.config.Database = name
	return r
}

// Collection returns the MongoDB collection associated with the repository.
func (r *Repository[T]) Collection() *mongo.Collection {
	return r.Database().Collection(r.config.Collection)
}

// Database returns the MongoDB database associated with the repository.
func (r *Repository[T]) Database() *mongo.Database {
	if r.config.Database == "" {
		panic("Configuration error: The Database name is not set. Set in New or use SetDatabase(name)")
	}

	return r.config.Client.Database(r.config.Database)
}

// Aggregate performs an aggregation query on the MongoDB collection.
func (r *Repository[T]) Aggregate(pipeline *Pipe, opts ...*AggrOpts) (*mongo.Cursor, error) {
	return r.Database().Aggregate(r.config.Context, pipeline, opts...)
}

// FindById retrieves an entity by its ID from the MongoDB collection.
func (r *Repository[T]) FindById(id string) *T {
	return r.FindOne(bson.M{"_id": r.createObjectId(id)})
}

// FindOne retrieves a single entity from the MongoDB collection based on the given query.
func (r *Repository[T]) FindOne(find Find, opts ...*FindOneOpts) *T {
	var entity T

	err := r.Collection().FindOne(r.config.Context, find, opts...).Decode(&entity)

	if err != nil {
		log.Printf("FindOne error: %s", err.Error())
		return nil
	}

	return &entity
}

// Find retrieves multiple entities from the MongoDB collection based on the given query.
func (r *Repository[T]) Find(find Find, opts ...*FindOpts) []*T {
	var entities []*T

	cursor, err := r.Collection().Find(r.config.Context, find, opts...)
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
// It validates the entity and generates a new ObjectId and created_at timestamp.
func (r *Repository[T]) Create(entity *T) error {
	if validateErr := r.validate(entity); validateErr != nil {
		return validateErr
	}

	// Parse the entity to BSON data
	data := r.parseEntity(entity)

	// Generate new ObjectId and set the "created_at" field
	newId := primitive.NewObjectID()
	data["_id"] = newId
	data["created_at"] = time.Now()

	// Create an options.FindOneAndUpdateOptions struct
	opts := options.FindOneAndUpdate().
		SetReturnDocument(1). // Directly set options.After here
		SetUpsert(true)       // Set upsert to true, meaning insert if not found

	// Insert and return the inserted document, or update if it already exists
	var result T
	err := r.Collection().FindOneAndUpdate(
		r.config.Context,
		bson.M{"_id": newId}, // Match condition (insert or update if matching ID)
		bson.M{"$set": data}, // Set fields
		opts,
	).Decode(&result)

	if err != nil {
		return err
	}

	// Update the original entity with the values from the inserted/updated document
	*entity = result

	return nil
}

// Update modifies an existing entity in the MongoDB collection based on its ID.
// It validates the entity and sets the updated_at timestamp.
func (r *Repository[T]) Update(id string, entity *T) error {
	if validateErr := r.validate(entity); validateErr != nil {
		return validateErr
	}

	// Prepare Data
	data := r.parseEntity(entity)
	data["updated_at"] = time.Now().UTC()
	delete(data, "_id")

	opts := options.FindOneAndUpdate().SetReturnDocument(1) // Return updated document

	var updatedEntity T

	err := r.Collection().FindOneAndUpdate(
		r.config.Context,
		bson.M{"_id": r.createObjectId(id)},
		bson.M{"$set": data},
		opts,
	).Decode(&updatedEntity)

	if err != nil {
		log.Printf("Update error: %s", err.Error())
		return nil
	}

	// Update the original entity with the values from the inserted/updated document
	*entity = updatedEntity

	return nil
}

// Delete removes an entity from the MongoDB collection based on its ID.
// It can perform a soft delete (set the deleted_at field) or a hard delete (remove the document).
func (r *Repository[T]) Delete(id string, soft bool) error {
	if soft {
		err := r.Collection().FindOneAndUpdate(
			r.config.Context,
			bson.M{"_id": r.createObjectId(id)},
			bson.M{"$set": bson.M{"deleted_at": time.Now()}},
		)
		return err.Err()
	}

	_, err := r.Collection().DeleteOne(r.config.Context, bson.M{"_id": r.createObjectId(id)})
	return err
}

// validate checks if the given entity passes validation using the go-playground/validator library.
func (r *Repository[T]) validate(entity *T) error {
	validate := validator.New()

	// Perform the validation
	validateErr := validate.Struct(entity)

	if validateErr != nil {
		// Map of field names to slices of validation errors
		errorMap := map[string][]string{}

		// Convert validation errors to map
		if validationErrors, ok := validateErr.(validator.ValidationErrors); ok {
			for _, ve := range validationErrors {
				// Add each error to the map under the appropriate field
				fieldName := ve.Field()
				errorMessage := ve.Tag() // Or you can get more info like ve.ActualTag(), etc.

				// Append error message to the map entry
				errorMap[fieldName] = append(errorMap[fieldName], errorMessage)
			}
		}

		errorsJson, _ := json.Marshal(map[string]map[string][]string{
			"errors": errorMap,
		})

		return errors.New(string(errorsJson))
	}

	// Return nil if no validation errors
	return nil
}

// createObjectId converts a string ID to a MongoDB ObjectId.
func (r *Repository[T]) createObjectId(id string) *primitive.ObjectID {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("FindByHexId error: %s", err.Error())
		return nil
	}

	return &objectID
}

// parseEntity converts an entity to a BSON map for storage in MongoDB.
func (r *Repository[T]) parseEntity(obj any) bson.M {
	val := reflect.ValueOf(obj)
	typ := val.Type()
	result := bson.M{}

	// Check if the object is a pointer, and get the underlying value if so
	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	// Iterate over struct fields
	for i := range val.NumField() {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)
		bsonTag := fieldType.Tag.Get("bson")

		// Skip fields with no bson tag or with a "-" tag (meaning ignore this field)
		if bsonTag == "" || bsonTag == "-" {
			continue
		}

		tags := strings.Split(bsonTag, ",")
		tag := tags[0]

		// Skip zero time.Time and ObjectID values
		if fieldVal.Kind() == reflect.Struct {
			if fieldVal.Type() == reflect.TypeOf(time.Time{}) && fieldVal.IsZero() {
				continue // Skip zero time values
			} else if fieldVal.Type() == reflect.TypeOf(primitive.ObjectID{}) && fieldVal.IsZero() {
				continue // Skip zero ObjectID values
			} else {
				// Process nested structs recursively
				subMap := r.parseEntity(fieldVal.Interface())
				if len(subMap) > 0 {
					result[tag] = subMap
				}
			}
		} else if (fieldVal.Kind() == reflect.Bool && !fieldVal.Bool()) || r.isNumeric(fieldVal.Kind()) {
			// keep false or 0 values as valuefull
			result[tag] = fieldVal.Interface()
		} else if !fieldVal.IsZero() { // Only add non-zero fields to the BSON map
			// Handle case where a field's bson tag might have multiple names separated by commas
			result[tag] = fieldVal.Interface()
		}
	}

	return result
}

// isNumeric checks if the given reflect.Kind is a numeric type.
func (r *Repository[T]) isNumeric(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}
