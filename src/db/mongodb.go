package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoDB struct {
	Ctx    context.Context
	Client *mongo.Client
	Cancel context.CancelFunc
}

func GetMongoDB() MongoDB {
	mongoDB := MongoDB{}
	var err error
	mongoDB.Ctx, mongoDB.Cancel = context.WithTimeout(context.Background(), 10*time.Second)
	mongoDB.Client, err = mongo.Connect(mongoDB.Ctx, options.Client().ApplyURI("mongodb://192.168.118.13:27017"))

	if err != nil {
		log.Fatal(err)
	}
	return mongoDB
}

func (m *MongoDB) GetCollection(key string) *mongo.Collection {
	return m.Client.Database("a-matrix").Collection(key)
}
