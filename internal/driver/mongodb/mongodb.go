package mongodb

import "tablepro/internal/driver"

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tablepro/internal/connection"
)

func New() *MongoDBDriver {
	return &MongoDBDriver{}
}

func (d *MongoDBDriver) Connect(ctx context.Context, config *connection.DatabaseConnection, password string) error {
	uri := buildURI(config, password)

	clientOpts := options.Client().ApplyURI(uri)

	if config.SSL.Enabled {
		clientOpts.TLSConfig = &tls.Config{
			InsecureSkipVerify: config.SSL.Mode == "skip-verify",
		}
	}

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB server: %w", err)
	}

	d.client = client
	d.config = config
	d.databaseName = config.Database

	return nil
}

func buildURI(config *connection.DatabaseConnection, password string) string {
	uri := fmt.Sprintf("mongodb://")

	if config.Username != "" {
		uri += fmt.Sprintf("%s:%s@", config.Username, password)
	}

	uri += fmt.Sprintf("%s:%d", config.Host, config.Port)

	if config.Database != "" {
		uri += "/" + config.Database
	}

	queryParams := []string{}
	if config.SSL.Enabled && config.SSL.Mode != "disable" {
		queryParams = append(queryParams, fmt.Sprintf("ssl=%v", config.SSL.Mode != "disable"))
	}

	if len(queryParams) > 0 {
		uri += "?" + strings.Join(queryParams, "&")
	}

	return uri
}

func (d *MongoDBDriver) Execute(ctx context.Context, query string) (*queryResult, error) {
	client := d.client.(*mongo.Client)
	collection := client.Database(d.databaseName).Collection(query)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("find execution failed: %w", err)
	}
	defer cursor.Close(ctx)

	return d.decodeCursor(ctx, cursor)
}

func (d *MongoDBDriver) ExecuteWithFilter(ctx context.Context, collectionName string, filter bson.M) (*queryResult, error) {
	client := d.client.(*mongo.Client)
	collection := client.Database(d.databaseName).Collection(collectionName)

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find execution failed: %w", err)
	}
	defer cursor.Close(ctx)

	return d.decodeCursor(ctx, cursor)
}

func (d *MongoDBDriver) ExecuteAggregate(ctx context.Context, collectionName string, pipeline interface{}) (*queryResult, error) {
	client := d.client.(*mongo.Client)
	collection := client.Database(d.databaseName).Collection(collectionName)

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate execution failed: %w", err)
	}
	defer cursor.Close(ctx)

	return d.decodeCursor(ctx, cursor)
}

func (d *MongoDBDriver) decodeCursor(ctx context.Context, cursor *mongo.Cursor) (*queryResult, error) {
	result := &queryResult{
		Columns: []string{},
		Rows:    make([][]any, 0),
	}

	if !cursor.Next(ctx) {
		return result, nil
	}

	firstDoc := cursor.Current
	var firstMap bson.M
	if err := bson.Unmarshal(firstDoc, &firstMap); err != nil {
		return nil, fmt.Errorf("failed to decode first document: %w", err)
	}

	for key := range firstMap {
		result.Columns = append(result.Columns, key)
	}

	rows := [][]any{docToRow(firstMap, result.Columns)}
	result.Rows = rows

	for cursor.Next(ctx) {
		var doc bson.M
		if err := bson.Unmarshal(cursor.Current, &doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		result.Rows = append(result.Rows, docToRow(doc, result.Columns))
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return result, nil
}

func docToRow(doc bson.M, columns []string) []any {
	row := make([]any, len(columns))
	for i, col := range columns {
		row[i] = doc[col]
	}
	return row
}

func (d *MongoDBDriver) ExecuteNonQuery(ctx context.Context, collectionName string, filter bson.M, update interface{}) (int64, error) {
	client := d.client.(*mongo.Client)
	collection := client.Database(d.databaseName).Collection(collectionName)

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("update execution failed: %w", err)
	}

	return result.ModifiedCount, nil
}

func (d *MongoDBDriver) InsertOne(ctx context.Context, collectionName string, document interface{}) (interface{}, int64, error) {
	client := d.client.(*mongo.Client)
	collection := client.Database(d.databaseName).Collection(collectionName)

	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return nil, 0, fmt.Errorf("insert failed: %w", err)
	}

	return result.InsertedID, 1, nil
}

func (d *MongoDBDriver) DeleteMany(ctx context.Context, collectionName string, filter bson.M) (int64, error) {
	client := d.client.(*mongo.Client)
	collection := client.Database(d.databaseName).Collection(collectionName)

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("delete failed: %w", err)
	}

	return result.DeletedCount, nil
}

func (d *MongoDBDriver) Ping(ctx context.Context) error {
	client := d.client.(*mongo.Client)
	return client.Ping(ctx, nil)
}

func (d *MongoDBDriver) Close() error {
	client := d.client.(*mongo.Client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return client.Disconnect(ctx)
}

func (d *MongoDBDriver) GetConfig() *connection.DatabaseConnection {
	return d.config
}

func (d *MongoDBDriver) IsConnected() bool {
	if d.client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client := d.client.(*mongo.Client)
	err := client.Ping(ctx, nil)
	return err == nil
}

func (d *MongoDBDriver) GetDatabase() string {
	return d.databaseName
}

func (d *MongoDBDriver) GetClient() *mongo.Client {
	if d.client == nil {
		return nil
	}
	return d.client.(*mongo.Client)
}

// Type returns the DatabaseType for this driver.
func (d *MongoDBDriver) Type() driver.DatabaseType {
	return driver.DatabaseTypeMongoDB
}
