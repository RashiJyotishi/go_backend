package config

import (
    "context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "log"
)

// We create a global variable so other files can use this "pipe"
var UserCollection *mongo.Collection

func ConnectDB() {
    // 1. The Blueprint: Where is the DB?
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

    // 2. The Connection: Open the pipe
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    UserCollection = client.Database("go_backend").Collection("users")
}