package db

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"imagine/utils"
)

func (db DB) Connect() (*mongo.Client, error) {
	host := fmt.Sprintf("%s:%d", db.Address, db.Port)
	fmt.Println("Connecting to MongoDB...")

	clientOpts := options.ClientOptions{
		AppName: &db.AppName,
		Auth: &options.Credential{
			Username:   db.User,
			Password:   db.Password,
			AuthSource: "admin",
		},
		Hosts: []string{host},
	}

	client, err := mongo.Connect(&clientOpts)
	if err != nil {
		return client, err
	}

	err = client.Ping(db.Context, nil)

	if err != nil {
		return client, fmt.Errorf("error pinging mongo: %w", err)
	}

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB at", db.Address, "on port", db.Port)
	fmt.Println("Using database", db.DatabaseName)

	return client, nil
}

// Disconnect closes the MongoDB client connection.
func (db DB) Disconnect(client *mongo.Client) error {
	err := client.Disconnect(db.Context)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a single document matching the filter.
func (db DB) Delete(document bson.D) (*mongo.DeleteResult, error) {
	result, err := db.Database.Collection(db.Collection).DeleteOne(db.Context, document)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Exists checks if a document matching the filter exists.
func (db DB) Exists(filter bson.D) (bool, error) {
	err := db.Database.Collection(db.Collection).FindOne(db.Context, filter).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Find retrieves documents matching the filter.
func (db DB) Find(filter bson.D, result any) (*mongo.Cursor, error) {
	cursor, err := db.Database.Collection(db.Collection).Find(db.Context, filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(db.Context)
	cursorErr := cursor.All(db.Context, result)
	if cursorErr != nil {
		return nil, cursorErr
	}

	return cursor, nil
}

// FindOne retrieves a single document matching the filter and decodes it into result.
func (db DB) FindOne(filter bson.D, result any) error {
	err := db.Database.Collection(db.Collection).FindOne(db.Context, filter).Decode(result)
	if err != nil {
		return err
	}
	return nil
}

// Insert adds a new document to the collection.
func (db DB) Insert(document bson.D) (*mongo.InsertOneResult, error) {
	result, err := db.Database.Collection(db.Collection).InsertOne(db.Context, document)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update modifies a single document matching the filter.
func (db DB) Update(filter bson.D, document bson.D) (*mongo.UpdateResult, error) {
	result, err := db.Database.Collection(db.Collection).UpdateOne(db.Context, filter, document)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db DB) UpdateMany(filter bson.D, documents []bson.D) (*mongo.UpdateResult, error) {
	result, err := db.Database.Collection(db.Collection).UpdateMany(db.Context, filter, documents)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db DB) DeleteMany(documents []bson.D) (*mongo.DeleteResult, error) {
	result, err := db.Database.Collection(db.Collection).DeleteMany(db.Context, documents)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db DB) InsertMany(documents []bson.D) (*mongo.InsertManyResult, error) {
	result, err := db.Database.Collection(db.Collection).InsertMany(db.Context, documents)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ReplaceOne replaces a single document matching the filter with the replacement document.
func (db DB) ReplaceOne(filter bson.D, replacement bson.D) (*mongo.UpdateResult, error) {
	result, err := db.Database.Collection(db.Collection).ReplaceOne(db.Context, filter, replacement)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Initis initializes the database connection.
func Initis() error {
	mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelMongo()

	var db DBClient = DB{
		Address:      "localhost",
		Port:         27017,
		User:         os.Getenv("MONGO_USER"),
		Password:     os.Getenv("MONGO_PASSWORD"),
		AppName:      utils.AppName,
		DatabaseName: "imagine-dev",
		Collection:   "images",
		Context:      mongoCtx,
	}

	client, err := db.Connect()
	if err != nil {
		panic(err)
	}

	defer func() {
		if client != nil {
			if disconnectErr := db.Disconnect(client); disconnectErr != nil {
				panic("error disconnecting from MongoDB: "+  disconnectErr.Error())
			}

			fmt.Println("Disconnected from MongoDB")
		}
	}()
	return err
}
