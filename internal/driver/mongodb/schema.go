package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d *MongoDBDriver) GetSchema(ctx context.Context) ([]string, error) {
	client := d.client.(*mongo.Client)
	db := client.Database(d.databaseName)

	cursor, err := db.ListCollections(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer cursor.Close(ctx)

	var databases []string
	seen := make(map[string]bool)

	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		if name, ok := result["name"].(string); ok {
			if !seen[name] {
				databases = append(databases, name)
				seen[name] = true
			}
		}
	}

	return databases, nil
}

func (d *MongoDBDriver) GetTables(ctx context.Context, databaseName string) ([]collectionInfo, error) {
	client := d.client.(*mongo.Client)
	db := client.Database(d.databaseName)

	filter := bson.M{}
	if databaseName != "" {
		filter = bson.M{"name": databaseName}
	}

	cursor, err := db.ListCollections(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer cursor.Close(ctx)

	var collections []collectionInfo

	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			continue
		}

		coll := collectionInfo{
			Name:    getStringSafe(result["name"]),
			Type:    getStringSafe(result["type"]),
			Options: getMapSafe(result["options"]),
		}
		collections = append(collections, coll)
	}

	return collections, nil
}

func (d *MongoDBDriver) GetCollections(ctx context.Context) ([]collectionInfo, error) {
	return d.GetTables(ctx, "")
}

func (d *MongoDBDriver) GetColumns(ctx context.Context, collectionName string) ([]map[string]any, error) {
	client := d.client.(*mongo.Client)
	collection := client.Database(d.databaseName).Collection(collectionName)

	opts := options.Find().SetLimit(1)
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get sample document: %w", err)
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return []map[string]any{}, nil
	}

	var doc bson.M
	if err := bson.Unmarshal(cursor.Current, &doc); err != nil {
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}

	var columns []map[string]any
	for key, value := range doc {
		col := map[string]any{
			"name":         key,
			"type":         inferBSONType(value),
			"nullable":     true,
			"is_primary":   key == "_id",
			"sample_value": value,
		}
		columns = append(columns, col)
	}

	return columns, nil
}

func (d *MongoDBDriver) GetIndexes(ctx context.Context, collectionName string) ([]indexInfo, error) {
	client := d.client.(*mongo.Client)
	collection := client.Database(d.databaseName).Collection(collectionName)

	indexes, err := collection.Indexes().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer indexes.Close(ctx)

	var result []indexInfo

	for indexes.Next(ctx) {
		var index bson.M
		if err := indexes.Decode(&index); err != nil {
			continue
		}

		idx := indexInfo{
			Name:      getStringSafe(index["name"]),
			Namespace: getStringSafe(index["ns"]),
			Unique:    getBoolSafe(index["unique"]),
			Sparse:    getBoolSafe(index["sparse"]),
		}

		if v, ok := index["key"].(bson.M); ok {
			idx.Keys = make(map[string]int)
			for k, val := range v {
				if num, ok := val.(int32); ok {
					idx.Keys[k] = int(num)
				} else if num, ok := val.(int64); ok {
					idx.Keys[k] = int(num)
				}
			}
		}

		result = append(result, idx)
	}

	return result, nil
}

func (d *MongoDBDriver) GetStats(ctx context.Context, collectionName string) (map[string]any, error) {
	client := d.client.(*mongo.Client)
	db := client.Database(d.databaseName)

	var result bson.M
	err := db.RunCommand(ctx, bson.D{
		{Key: "collStats", Value: collectionName},
	}).Decode(&result)

	if err != nil {
		return nil, fmt.Errorf("failed to get collection stats: %w", err)
	}

	return result, nil
}

func getStringSafe(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getBoolSafe(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func getMapSafe(v any) map[string]any {
	if m, ok := v.(bson.M); ok {
		result := make(map[string]any)
		for k, val := range m {
			result[k] = val
		}
		return result
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func inferBSONType(v any) string {
	switch v.(type) {
	case string:
		return "string"
	case int, int32, int64:
		return "int"
	case float32, float64:
		return "double"
	case bool:
		return "bool"
	case []byte:
		return "binData"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("%T", v)
	}
}
