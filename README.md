# mongorepo

Is a library for use generic Repository based on Entities (structs) with bson definition

## Install:

```bash
go get github.com/eliasnoya/mongorepo
```

mongorepo includes go.mongodb.org/mongo-driver an all his dependencies

## main.go Example

```go
package main

import (
	"context"
	"time"

	"github.com/eliasnoya/mongorepo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EntityTest struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
	DeletedAt time.Time          `bson:"deleted_at,omitempty"`
	Name      string             `bson:"name"`
}

func main() {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic("failed to connect mongo db")
	}

	repo := mongorepo.New[EntityTest](&mongorepo.Config{
		Collection: client.Database("sarasa").Collection("entity_test"),
	})

	repo.Create(&EntityTest{
		Name: "Elias Noya",
	})
}
```


## Example Entity:

```go
// EntityTest represents an example MongoDB entity with fields for ID, creation, update, deletion timestamps, and a name.
// 
// This entity is designed to work with the mongorepo package, which provides generic repository functions for MongoDB.
//
// Important Notes:
// 1. **Panic Conditions**: The repository functions will panic under the following circumstances:
//    - If the `ID` field is not present or is not of type `primitive.ObjectID`.
//    - If `CreatedAt`, `DeletedAt`, or `UpdatedAt` fields are set in the repository configuration (see `mongorepo.Config`),
//      but are missing in the entity or are not of type `time.Time`. You can always disable timestamp fields by setting 
//      the respective fields in `mongorepo.Config` to empty values.
// 2. **Soft Deletes**: The `DeletedAt` field must include the `omitempty` tag to allow for proper handling of soft deletes,
//    meaning it will be omitted from the BSON document if it has a zero value (i.e., the field has not been set).
type EntityTest struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
	DeletedAt time.Time          `bson:"deleted_at,omitempty"` // Important!! dont forget omitempty for entities with softdeletes
	Name      string             `bson:"name"`
}
```

## Using the go-generic repository:

```go
// Instance your mongo client
client, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))

// using defaults ID, CreatedAt, UpdatedAt, DeletedAt and Context
repo := mongorepo.NewDefault[EntityTest](client.Database("example_db").Collection("entity_test"))

// if name: Jon exists will return an EntityTest with all the mongo document data or nil
// REMEMBER the repo will not exclude automatically when a document is softDeleted with DeletedAtField,
// if you want only non-deleted records add in the query for this example:
// bson.M{"name": "Jon", "deleted_at": bson.M{"$exists": false}}
entity := repo.FindOne(bson.M{"name": "Jon"}) 

if entity == nil {
	return "Document not found"
}
```

## Using custom Config:
```go
// Example entities with custom ID / CreateAt / UpdatedAt fields names and without softdeletes
repo := mongorepo.New[EntityTest](&mongorepo.Config{
    Collection:     client.Database("example_db").Collection("entity_test"),
	Context: 		context.Background(), // default context.Background()
    IdField:        "MyID", 		// default ID
    CreatedAtField: "MyCreatedAt",
    UpdatedAtField: "MyUpdatedAt",
    DeletedAtField: ""
})

// or
// if you want a repository without timestamps and softdeletes
// "ID" (mayus) is the default property name for id fields
repo := mongorepo.New[EntityTest](&mongorepo.Config{
    Collection:     client.Database("example_db").Collection("entity_test"),
})
```

## Crud Operations

```go
// Create
err := repo.Create(&EntityTest{
    Name: "Elías",
})

// Update
err := repo.Update(&EntityTest{
    Id:     primitive.ObjectIDFromHex("66b70c0eb9bd318bec55d93d")
    Name:   "Jorge",
})

// Delete
err := repo.Delete(&EntityTest{
    Id:     primitive.ObjectIDFromHex("66b70c0eb9bd318bec55d93d")
})

// FindById
entity := repo.FindById(primitive.ObjectIDFromHex("66b70c0eb9bd318bec55d93d"))

// FindOne (by name in this case)
entity := repo.FindOne(bson.M{
    "name": "Elías",
}, &options.FindOneOptions{
    Sort: bson.M{"created_at": -1}
})

// Find (no filters, sorted by created_at)
entities := repo.Find(bson.M{}, &options.FindOptions{Sort: bson.M{"created_at": -1}})
```

## Using your own implementations

```go
// To define a repository for your entity type, extend the generic IRepository interface provided by the base mongorepo. 
// Here’s how to set up and use your custom repository:
type MyEntityRepository struct {
	mongorepo.IRepository[EntityTest]
}

// Create a constructor function for your repository that accepts an IRepository[T] instance. 
// This allows you to initialize your custom repository with the generic repository functionality.
func NewMyEntityRepository(repository mongorepo.IRepository[EntityTest]) *MyEntityRepository {
	return &MyEntityRepository{
		IRepository: repository,
	}
}

// You can add custom methods to your repository to extend its functionality. 
// For example, you might want to make your custom query with *mongo.Collection
func (m *MyEntityRepository) MyAggregate() []*EntityTest {
	// or m.collection().... access mongo collection and do staff
	m.Collection().Aggregate(/* your aggregate logic */)
}

// Instantiate your custom repository by passing a generic Repository[T] implementation. 
// Here’s how you can set it up and use it:
myRepository := NewMyEntityRepository(mongorepo.New[EntityTest](&mongorepo.Config{
    Collection: client.Database("example_db").Collection("entity_test"),
}))

// Use your functions
records := mymongorepo.All() // loop your EntityTest slice

// Use generic functions
entityOne := mymongorepo.FindOne(bson.M{"name": "Elías"})
```
