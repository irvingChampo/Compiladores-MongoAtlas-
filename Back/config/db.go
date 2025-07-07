package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv" // Importa godotenv
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

func ConnectMongo() {
	// Carga las variables del archivo .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No se pudo cargar archivo .env, se usarán variables del entorno")
	}

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI no definido en el entorno")
	}

	clientOptions := options.Client().ApplyURI(uri)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Conectado a MongoDB Atlas")
	MongoClient = client
}

func GetCollection(collectionName string) *mongo.Collection {
	return MongoClient.Database("BaseChampo").Collection(collectionName)
}
