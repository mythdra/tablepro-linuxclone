package redis

import (
	"tablepro/internal/connection"
)

const (
	TypeString = "string"
	TypeList   = "list"
	TypeSet    = "set"
	TypeZSet   = "zset"
	TypeHash   = "hash"
	TypeStream = "stream"
	TypeNone   = "none"
)

const (
	FeatureKeyspace     = "KEYSPACE"
	FeaturePubSub       = "PUBSUB"
	FeatureTransactions = "TRANSACTIONS"
	FeatureCluster      = "CLUSTER"
	FeatureScripts      = "SCRIPTS"
	FeatureAOF          = "AOF"
	FeatureRDB          = "RDB"
)

type RedisDriver struct {
	client   interface{}
	config   *connection.DatabaseConnection
	Database int
	cluster  bool
}

type RedisVersion struct {
	Major   int
	Minor   int
	Patch   int
	Mode    string
	Modules []string
}

type queryResult struct {
	Columns      []string
	Rows         [][]any
	AffectedRows int64
}

type keyInfo struct {
	Key       string
	Type      string
	TTL       int64
	Encoding  string
	MemoryUse int64
	Frequency float64
	IdleTime  int64
}

type serverInfo struct {
	Version     string
	Mode        string
	Connected   bool
	DBSize      int64
	MemoryUsed  int64
	CPUUsed     float64
	Clients     int64
	Uptime      int64
	Role        string
	Replication map[string]any
	Cluster     bool
}

type scanResult struct {
	Cursor int64
	Keys   []string
}
