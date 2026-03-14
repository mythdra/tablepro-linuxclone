package mongodb

import (
	"tablepro/internal/connection"
)

// MongoDB-specific type mappings
const (
	TypeString    = "string"
	TypeInt       = "int"
	TypeInt32     = "int32"
	TypeInt64     = "int64"
	TypeDouble    = "double"
	TypeBool      = "bool"
	TypeObjectID  = "objectId"
	TypeDate      = "date"
	TypeTimestamp = "timestamp"
	TypeBinary    = "binary"
	TypeRegex     = "regex"
	TypeArray     = "array"
	TypeObject    = "object"
	TypeNull      = "null"
	TypeUndefined = "undefined"
	TypeDecimal   = "decimal"
	TypeMinKey    = "minKey"
	TypeMaxKey    = "maxKey"
)

// MongoDBDriver implements the DatabaseDriver interface for MongoDB
type MongoDBDriver struct {
	client       interface{}
	config       *connection.DatabaseConnection
	databaseName string
}

// MongoDBVersion represents MongoDB server version info
type MongoDBVersion struct {
	Version string
	Major   int
	Minor   int
	Patch   int
	IsAtlas bool
}

// queryResult represents the result of a query execution
type queryResult struct {
	Columns      []string
	Rows         [][]any
	AffectedRows int64
}

// collectionInfo represents MongoDB collection information
type collectionInfo struct {
	Name    string
	Type    string
	Options map[string]any
}

// indexInfo represents MongoDB index information
type indexInfo struct {
	Name                    string
	Namespace               string
	Keys                    map[string]int
	Unique                  bool
	Sparse                  bool
	TTL                     int
	Version                 int
	PartialFilterExpression map[string]any
}
