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
	ID        	primitive.ObjectID 	`bson:"_id"`	// Even if you dont define this field will set _id = primitive.NewObjectID() with repo.Create(...)
	CreatedAt 	time.Time          	`bson:"created_at"`	// Even if you dont define this field will set created_at = time.Now().UTC() with repo.Create(...)
	UpdatedAt 	time.Time          	`bson:"updated_at"`	// Even if you dont define this field will set updated_at = time.Now().UTC() with repo.Update(...)
	DeletedAt 	time.Time          	`bson:"deleted_at"`	// dont forget deleted_at if you want to softDelete document repo.Delete(id, true)
	Name      	string             	`bson:"name" validate:"required"`
	Value		int 				`bson:"value"`
}

func main() {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic("failed to connect mongo db")
	}

	repo := mongorepo.New[EntityTest](&mongorepo.Config{
		Client:     client,
		Database:   "default_db",
		Collection: "entity_test",
	})

	entity := &EntityTest{
		Name: "Elias Noya", // name is required if isnt set on Create() or Update() will fail
	}

	createErr := repo.Create(entity)

	if createErr != nil {
		fmt.Println(createErr.Error())
	}

	entity.Value = 10
	updateErr := repo.Update("67da26f5ac7a82a238854216", entity)

	if updateErr != nil {
		fmt.Println(updateErr.Error())
	}

	fmt.Println(entity.Name) // Elias Noya
	fmt.Println(entity.Value) // 10

	deleteErr := repo.Delete("67da26f5ac7a82a238854216", true) // delete soft (update deleted_at field)

	if deleteErr != nil {
		fmt.Println(deleteErr.Error())
	}

	entityStored := repo.FindById("67da26f5ac7a82a238854216")

	if entityStored == nil {
		// not found
	}

	entityList := repo.Find(mongorepo.Find{"value": "10"})

	if entityList == nil {
		// not found
	}

	entityOne := repo.FindOne(mongorepo.Find{"value": "10"})

	if entityOne == nil {
		// not found
	}

}
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
			Client:     client,
			Database:   dbname,
			Collection: "entity_test",
		}),
	}
}

func (r *MyEntityRepository) MyFunc() {
	// my custom logic and/or queries
	cursor, err := r.Aggregate(&mongorepo.Pipe{/* .... */})
}

// Instantiate your custom repository by passing a generic Repository[T] implementation. 
// Here’s how you can set it up and use it:
myRepository := NewMyEntityRepository(client, "example_db")

// Use your functions
x := myRepository.MyFunc() // call your custom method
```
