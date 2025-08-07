package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	*mongo.Client
	Database *mongo.Database
}

type MongoConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

func NewMongoClient(config MongoConfig) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	return &MongoClient{
		Client:   client,
		Database: database,
	}, nil
}

func (m *MongoClient) GetTenantDatabase(tenantID string) *mongo.Database {
	dbName := fmt.Sprintf("tenant_%s", tenantID)
	return m.Client.Database(dbName)
}

func (m *MongoClient) GetTenantCollection(tenantID, collectionName string) *mongo.Collection {
	db := m.GetTenantDatabase(tenantID)
	return db.Collection(collectionName)
}

func (m *MongoClient) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}
