package database

import(
"fmt"
"log"
"time"
"os"
"context"
"github.com/joho/godotenv"
"go.mongodb.org/mongo-driver/mongo"
"go.mongodb.org/mongo-driver/mongo/options"
)

// this is the database instance that we used in our project to store the data.

func DBinstance() *mongo.Client{

	// load the env for functioning.
	err := godotenv.Load(".env")
	if err!=nil{
		log.Fatal("Error loading .env file")
	}

	//mongo url for passing the instance.
	MongoDb := os.Getenv("MONGODB_URL")

	//making the client to act as the interface.
	client, err:= mongo.NewClient(options.Client().ApplyURI(MongoDb))
	if err != nil {
		log.Fatal(err)
	}

	//context is used to connet eith the data copy, for a particular time being.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	return client
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection{
	var collection *mongo.Collection = client.Database("cluster0").Collection(collectionName)
	// cluster0 is the database name, which we made in the database.
	return collection
}