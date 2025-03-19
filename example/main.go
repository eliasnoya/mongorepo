package main

import (
	"context"
	"fmt"
	"time"

	"github.com/eliasnoya/mongorepo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoEntity struct {
}

type EntityTest struct {
	ID        primitive.ObjectID `bson:"_id"`
	Sub       SubEntity          `bson:"sub_entity"`
	Name      string             `bson:"name" validate:"required"`
	Active    bool               `bson:"active"`
	Value     int                `bson:"value" validate:"required"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	DeletedAt time.Time          `bson:"deleted_at"`
}

type SubEntity struct {
	A   string     `bson:"a"`
	B   int        `bson:"b"`
	Sub SubEntity2 `bson:"sub"`
}

type SubEntity2 struct {
	A string `bson:"a"`
	B int    `bson:"b"`
}

type M = map[string]any

func main() {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		panic("failed to connect mongo db")
	}

	repo := mongorepo.New[EntityTest](&mongorepo.Config{
		Client:     client,
		Database:   "tenant_1",
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

	fmt.Println(entity.Name)  // Elias Noya
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

	for _, item := range entityList {
		fmt.Println(item.Active)
	}

	entityOne := repo.FindOne(mongorepo.Find{"value": "10"})

	if entityOne == nil {
		// not found
	}

	// result := repo.Find(mongorepo.Find{"name": "Florencia Noya"})

	// jr, _ := json.Marshal(result)

	// fmt.Println(string(jr))

	// all := repo.Find(bson.M{})

	// for _, v := range all {

	// 	// Marshal the struct into JSON
	// 	jsonData, err := json.Marshal(v)
	// 	if err != nil {
	// 		fmt.Println("Error marshaling JSON:", err)
	// 		return
	// 	}

	// 	// Print the JSON data
	// 	fmt.Println(string(jsonData))
	// }

	// if errC != nil {
	// 	log.Println(errC.Error())
	// }
}
