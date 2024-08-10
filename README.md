# go-mongo-repository

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

// using defaults ID, DeletedAt, CreatedAt, UpdatedAt and Context
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
repo := mongorepo.New[EntityTest](&mongorepo.Config{
    Collection:     client.Database("example_db").Collection("entity_test"),
	Context: 		context.TODO(), // default context.Background()
    IdField:        "MyID", 		// default ID
    CreatedAtField: "MyCreatedAt", 	// default CreatedAt
    UpdatedAtField: "MyUpdatedAt", 	// default UpdatedAt
    DeletedAtField: "MyDeletedAt" 	// default "" (if is empty will hard-delete the documents, if is set to a time.Time field will update that)
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

// You can add custom methods to your repository to extend its functionality. For example, you might want to retrieve all entities sorted by a specific field:
func (m *MyEntityRepository) All() []*EntityTest {
	return m.Find(bson.M{}, &options.FindOptions{Sort: bson.M{"created_at": -1}})
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
