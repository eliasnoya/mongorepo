# mongorepo

`mongorepo` is a lightweight, flexible wrapper designed to simplify MongoDB operations by applying the Repository Pattern. It provides an easy-to-use abstraction layer over MongoDB, allowing developers to interact with MongoDB collections using Go structs and standard CRUD operations.

## Key Features:
- **Repository Pattern**: Encapsulates the logic needed to interact with MongoDB, promoting separation of concerns and clean code architecture.
- **CRUD Operations**: Easily create, read, update, and delete documents using Go structs that follow MongoDB field conventions (e.g., using `bson` tags).
- **Custom Repository Logic**: Extend the default repository behavior by adding your own custom methods, enabling more complex or specialized operations.
- **Soft Deletes**: Optionally support soft deletes, allowing you to mark documents as deleted without physically removing them from the database.
- **Timestamps**: Automatically manage `createdAt` and `updatedAt` timestamps, making it easy to track document lifecycle changes.

## Install:

```bash
go get github.com/eliasnoya/mongorepo
```

## Simple and basic example using Generic Repo

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
repo := mongorepo.New[EntityTest](&mongorepo.Config{
	MongoClient:    client,
	DbName:         "test_db",
	IdField:        "MyID",      // Default: ID
	CreatedAtField: "CreatedAt", // Default: disabled
	UpdatedAtField: "UpdatedAt", // Default: disabled
	DeletedAtField: "DeletedAt", // Default: disabled
})

// if name: Jon exists will return an EntityTest with all the mongo document data or nil
// REMEMBER the repo will not exclude automatically when a document is softDeleted with DeletedAtField,
// if you want only non-deleted records add in the query for this example:
// bson.M{"name": "Jon", "deleted_at": bson.M{"$exists": false}}
entity := repo.FindOne(bson.M{"name": "Jon"}) 

if entity == nil {
	return "Document not found"
}
```

## Using Config:
```go
// Example entities with custom ID / CreateAt / UpdatedAt fields names and without softdeletes
repo := mongorepo.New[EntityTest](&mongorepo.Config{
	MongoClient:    client, 			// Mandatory config
	DbName:         "test_db", 			// Mandatory config
	CollectionName: "my_entity_test" 	// Default if not set: snake case, lower and plural struct name, in this case entity_tests
	IdField:        "MyID",      		// Default if not set: ID
	CreatedAtField: "CreatedAt", 		// Default if not set: disabled
	UpdatedAtField: "UpdatedAt", 		// Default if not set: disabled
	DeletedAtField: "DeletedAt", 		// Default if not set: disabled
})
```
```go
// All config properties
// Config holds the configuration necessary for connecting and interacting with a MongoDB collection.
type Config struct {
	MongoClient       *mongo.Client              // The MongoDB client instance used for database connections.
	DatabaseOptions   *options.DatabaseOptions   // The MongoDb Database options, default: nil
	CollectionOptions *options.CollectionOptions // The MongoDb Collection options, default: nil
	DbName            string                     // The name of the database where the collection resides.
	CollectionName    string                     // The name of the collection representing the entity.
	Context           context.Context            // The context to manage request lifecycle (e.g., timeouts, cancellations) during MongoDB operations.
	IdField           string                     // The field in the entity struct that represents the "_id" field in MongoDB, which must be a primitive.ObjectID.
	DeletedAtField    string                     // The field in the entity struct to track soft deletes, indicating when a document is marked as deleted.
	CreatedAtField    string                     // The field in the entity struct to store the timestamp of when the document was created; must be of type time.Time.
	UpdatedAtField    string                     // The field in the entity struct to store the timestamp of when the document was last updated; must be of type time.Time.
}
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

// FindByHexId
entity := repo.FindByHexId("66b70c0eb9bd318bec55d93d")
// or FindById(primitve.ObjectID)

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
// you can add parameters if you need it like collectionName
func NewMyEntityRepository(client *mongo.Client, dbname string) *MyEntityRepository {
	return &MyEntityRepository{
		IRepository: mongorepo.New[EntityTest](&mongorepo.Config{
			MongoClient: client, // Mandatory config
			DbName:      dbname, // Mandatory config
		}),
	}
}

func (r *MyEntityRepository) MyFunc() {
	// my custom logic and/or queries
	cursor, err := r.Aggregate(&mongo.Pipeline{/* .... */})
}

// Instantiate your custom repository by passing a generic Repository[T] implementation. 
// Here’s how you can set it up and use it:
myRepository := NewMyEntityRepository(client, "example_db")

// Use your functions
x := myRepository.MyFunc() // call your custom method
```
