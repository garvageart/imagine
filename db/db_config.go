package db

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type DB struct {
	Address      string
	Protocol     string
	Port         int
	User         string
	Password     string
	DatabaseName string
	Database     *mongo.Database
	AppName      string
	Collection   string
	Context      context.Context
}

// DBClient defines the interface for database operations.
type DBClient interface {
	Connect() (*mongo.Client, error)
	Disconnect(*mongo.Client) (error)
	Insert(document bson.D) (*mongo.InsertOneResult, error)
	InsertMany(documents []bson.D) (*mongo.InsertManyResult, error)
	Update(filter bson.D, document bson.D) (*mongo.UpdateResult, error)
	UpdateMany(filter bson.D, documents []bson.D) (*mongo.UpdateResult, error)
	ReplaceOne(filter bson.D, replacement bson.D) (*mongo.UpdateResult, error)
	Exists(filter bson.D) (bool, error)
	Delete(document bson.D) (*mongo.DeleteResult, error)
	DeleteMany(documents []bson.D) (*mongo.DeleteResult, error)
	Find(filter bson.D, result any) (*mongo.Cursor, error)
	FindOne(filter bson.D, result any) (error)
}