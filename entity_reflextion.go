package mongorepo

import (
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EntityReflection provides reflection-based operations on a MongoDB entity.
type EntityReflection struct {
	entity any
	config *Config
}

// NewEntityReflection creates a new EntityReflection object, ensuring the entity is either a struct or a pointer to a struct.
// If the entity is not valid, it panics.
//
// Parameters:
//   - config: A pointer to the Config object.
//   - entity: The entity to be reflected upon. Must be a struct or a pointer to a struct.
//
// Returns:
//   - A pointer to an EntityReflection object.
func NewEntityReflection(config *Config, entity any) *EntityReflection {
	entityType := reflect.TypeOf(entity)

	// Check if entity is a struct or a pointer to a struct
	if entityType.Kind() != reflect.Struct && !(entityType.Kind() == reflect.Ptr && entityType.Elem().Kind() == reflect.Struct) {
		exception := fmt.Sprintf("entity must be a struct or a pointer to a struct, got %s", entityType.Kind())
		panic(exception)
	}

	if config.IdField == "" {
		config.IdField = "ID"
	}

	return &EntityReflection{
		entity: entity,
		config: config,
	}
}

// GetID retrieves the ObjectID from the entity's ID field specified in the configuration.
// It panics if the ID field is not found or is not of type primitive.ObjectID.
//
// Returns:
//   - The ObjectID from the entity's ID field.
func (er *EntityReflection) GetID() primitive.ObjectID {
	entityElem := reflect.ValueOf(er.entity).Elem()
	idField := entityElem.FieldByName(er.config.IdField)

	if !idField.IsValid() {
		exception := fmt.Sprintf("Error: Field %q not found in entity. Check if %q is the correct field name in the entity struct.", er.config.IdField, er.config.IdField)
		panic(exception)
	}

	if idField.Type() != reflect.TypeOf(primitive.ObjectID{}) {
		exception := fmt.Sprintf("Error: Field %q in entity is not of type primitive.ObjectID. Actual type: %s", er.config.IdField, idField.Type().String())
		panic(exception)
	}

	return idField.Interface().(primitive.ObjectID)
}

// SetNewID sets a new ObjectID to the entity's ID field specified in the configuration.
// It panics if the ID field is not found, cannot be set, or is not of type primitive.ObjectID.
func (er *EntityReflection) SetNewID() {
	entityElem := reflect.ValueOf(er.entity).Elem()
	idField := entityElem.FieldByName(er.config.IdField)

	if !idField.IsValid() || !idField.CanSet() || idField.Type() != reflect.TypeOf(primitive.ObjectID{}) {
		errorStr := fmt.Sprintf("Error: ID field %q is either not found or cannot be set. Ensure it is defined as primitive.ObjectID", er.config.IdField)
		panic(errorStr)
	}

	idField.Set(reflect.ValueOf(primitive.NewObjectID()))
}

// SetUpdateAt sets the current time to the entity's UpdatedAt field specified in the configuration.
func (er *EntityReflection) SetUpdateAt() {
	er.setTimeStampField(er.config.UpdatedAtField)
}

// SetCreatedAt sets the current time to the entity's CreatedAt field specified in the configuration.
func (er *EntityReflection) SetCreatedAt() {
	er.setTimeStampField(er.config.CreatedAtField)
}

// SetDeletedAt sets the current time to the entity's DeletedAt field specified in the configuration.
func (er *EntityReflection) SetDeletedAt() {
	er.setTimeStampField(er.config.DeletedAtField)
}

// setTimeStampField sets the current time to the specified field of type time.Time in the entity.
// It panics if the field is not found or is not of type time.Time.
//
// Parameters:
//   - field: The name of the field to set the timestamp on.
func (er *EntityReflection) setTimeStampField(field string) {
	entityElem := reflect.ValueOf(er.entity).Elem()
	timeField := entityElem.FieldByName(field)

	if !timeField.IsValid() {
		exception := fmt.Sprintf("Error: Field %q not found in entity. Ensure the field name is correct.", field)
		panic(exception)
	}

	if timeField.Type() != reflect.TypeOf(time.Time{}) {
		exception := fmt.Sprintf("Error: Field %q in entity is not of type time.Time. Actual type: %s", field, timeField.Type().String())
		panic(exception)
	}

	timeField.Set(reflect.ValueOf(time.Now()))
}
