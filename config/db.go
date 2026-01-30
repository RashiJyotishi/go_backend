package config

import (
    "os"
    "context"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "log"
)

// We create a global variable so other files can use this "pipe"
var UserCollection *mongo.Collection

func ConnectDB() {
    // 1. The Blueprint: Where is the DB?
    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        log.Fatal("MONGO_URI not set in environment")
    }
    clientOptions := options.Client().ApplyURI(mongoURI)

    // 2. The Connection: Open the pipe
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    UserCollection = client.Database("go_backend").Collection("users")
}