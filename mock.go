package mongorepo

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MockRepository[T any] struct {
	MemoryDb map[string]T
	config   *RepositoryConfig
}

func NewMockRepository[T any](config *RepositoryConfig) IRepository[T] {
	if config.IdField == "" {
		config.IdField = "ID"
	}
	if config.CreatedAtField == "" {
		config.CreatedAtField = "CreatedAt"
	}
	if config.UpdatedAtField == "" {
		config.UpdatedAtField = "UpdatedAt"
	}

	return &MockRepository[T]{
		MemoryDb: make(map[string]T),
		config:   config,
	}
}

// getEntityObjectID retrieves the ObjectID from the entity's ID field.
// Returns primitive.NilObjectID if the field is not found or is not of type primitive.ObjectID.
func (r *MockRepository[T]) getEntityObjectID(entity *T) primitive.ObjectID {
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

// setNewObjectID assigns a new ObjectID to the entity's ID field if it is not already set.
// The ID field must be of type primitive.ObjectID.
func (r *MockRepository[T]) setNewObjectID(entity *T) error {
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

// setEntityTimestamp sets the current timestamp to the specified field in the entity.
// The field must be of type time.Time.
func (r *MockRepository[T]) setEntityTimestamp(entity *T, field string) {
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

// FindById retrieves a single entity by its ID (primitive.ObjectID).
// Returns a pointer to the entity or nil if not found.
func (m *MockRepository[T]) FindById(id primitive.ObjectID) *T {
	idStr := id.Hex() // Convert ObjectID to string
	if entity, exists := m.MemoryDb[idStr]; exists {
		return &entity
	}
	return nil
}

// FindOne executes a find operation using the provided search criteria (`bson.M`).
// Returns a pointer to the found entity or nil if not found.
func (m *MockRepository[T]) FindOne(query bson.M, opts ...*options.FindOneOptions) *T {
	for _, entity := range m.MemoryDb {
		if matchesQuery(entity, query) {
			return &entity
		}
	}
	return nil
}

// Find retrieves a list of entities matching the provided search criteria (`bson.M`).
// Returns a slice of pointers to the found entities.
func (m *MockRepository[T]) Find(query bson.M, opts ...*options.FindOptions) []*T {
	var results []*T
	for _, entity := range m.MemoryDb {
		if matchesQuery(entity, query) {
			results = append(results, &entity)
		}
	}
	return results
}

// Create persists a new entity in the repository.
// Requires a pointer to the entity object. Returns nil if successful.
func (m *MockRepository[T]) Create(entity *T) error {
	m.setNewObjectID(entity)
	m.setEntityTimestamp(entity, m.config.CreatedAtField)
	id := m.getEntityObjectID(entity)
	m.MemoryDb[id.Hex()] = *entity
	return nil
}

// Update updates an existing entity in the repository.
// Requires a pointer to the modified entity object. Returns nil if successful.
func (m *MockRepository[T]) Update(entity *T) error {
	id := m.getEntityObjectID(entity)

	m.setEntityTimestamp(entity, m.config.UpdatedAtField)

	if _, exists := m.MemoryDb[id.String()]; exists {
		m.MemoryDb[id.String()] = *entity
		return nil
	}
	return fmt.Errorf("entity not found")
}

// Delete removes an entity from the repository by its ID (primitive.ObjectID).
// Returns nil if successful.
func (m *MockRepository[T]) Delete(entity *T) error {
	id := primitive.NewObjectID().Hex()
	delete(m.MemoryDb, id)
	return nil
}

// getBsonTagName returns the BSON tag name for a given struct field.
func getBsonTagName(structField reflect.StructField) string {
	tag := structField.Tag.Get("bson")
	if tag == "" {
		return structField.Name
	}
	// BSON tag might be comma-separated, return the first part
	return tag
}

// matchesQuery checks if an entity matches the provided query.
func matchesQuery[T any](entity T, query bson.M) bool {
	v := reflect.ValueOf(entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for key, value := range query {
		found := false
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			structField := t.Field(i)
			bsonTag := getBsonTagName(structField)

			if bsonTag == key {
				if !reflect.DeepEqual(field.Interface(), value) {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
