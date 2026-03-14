package mysql

import (
	"database/sql"

	"tablepro/internal/connection"
)

// MySQL-specific type mappings
const (
	TypeBool               = "bool"
	TypeInt                = "int"
	TypeBigInt             = "bigint"
	TypeFloat              = "float"
	TypeDouble             = "double"
	TypeDecimal            = "decimal"
	TypeVarchar            = "varchar"
	TypeChar               = "char"
	TypeText               = "text"
	TypeTinyText           = "tinytext"
	TypeMediumText         = "mediumtext"
	TypeLongText           = "longtext"
	TypeBlob               = "blob"
	TypeTinyBlob           = "tinyblob"
	TypeMediumBlob         = "mediumblob"
	TypeLongBlob           = "longblob"
	TypeEnum               = "enum"
	TypeSet                = "set"
	TypeDate               = "date"
	TypeDateTime           = "datetime"
	TypeTimestamp          = "timestamp"
	TypeTime               = "time"
	TypeYear               = "year"
	TypeJSON               = "json"
	TypeGeometry           = "geometry"
	TypePoint              = "point"
	TypeLineString         = "linestring"
	TypePolygon            = "polygon"
	TypeMultiPoint         = "multipoint"
	TypeMultiLineString    = "multilinestring"
	TypeMultiPolygon       = "multipolygon"
	TypeGeometryCollection = "geometrycollection"
)

// MySQLVersion represents MySQL/MariaDB version info
type MySQLVersion struct {
	Version   string
	Major     int
	Minor     int
	Patch     int
	IsMariaDB bool
}

// queryResult represents the result of a query execution
type queryResult struct {
	Columns      []string
	Rows         [][]any
	AffectedRows int64
}

// MySQLDriver implements the DatabaseDriver interface for MySQL/MariaDB
type MySQLDriver struct {
	db          *sql.DB
	config      *connection.DatabaseConnection
	version     *MySQLVersion
	transaction *sql.Tx
}
