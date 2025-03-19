package main

import (
	"context"
	"encoding/json"
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

	createEntity := &EntityTest{
		Name:  "Elias",
		Value: 10,
	}

	createErr := repo.Create(createEntity)

	if createErr != nil {
		fmt.Println(createErr.Error())
	}

	jr0, _ := json.Marshal(createEntity)

	fmt.Println(string(jr0))

	repo.SetDatabase("tenant_2").Create(createEntity)

	repo.Delete(createEntity.ID.Hex(), true)

	createEntity.Active = false
	createEntity.Value = 0

	updateErr := repo.SetDatabase("tenant_2").Update("67da1de1c48a6b86134e3c98", createEntity)

	if updateErr != nil {
		fmt.Println(updateErr.Error())
	}

	find := repo.FindOne(mongorepo.Find{})

	jr, _ := json.Marshal(find)
	fmt.Println("Entity find:", string(jr))

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
