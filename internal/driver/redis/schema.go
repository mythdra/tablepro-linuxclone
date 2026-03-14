package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func (d *RedisDriver) GetSchema(ctx context.Context) (map[string]any, error) {
	client := d.client.(*redis.Client)

	info, err := client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	result := make(map[string]any)
	result["server_info"] = parseInfo(info)
	result["database"] = d.Database

	dbSize, err := client.DBSize(ctx).Result()
	if err == nil {
		result["db_size"] = dbSize
	}

	return result, nil
}

func (d *RedisDriver) GetTables(ctx context.Context) ([]keyInfo, error) {
	client := d.client.(*redis.Client)
	var keys []string
	var cursor uint64

	for {
		result, next, err := client.Scan(ctx, cursor, "*", 1000).Result()
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		keys = append(keys, result...)
		cursor = next
		if cursor == 0 {
			break
		}
	}

	var keyInfos []keyInfo

	for _, key := range keys {
		ktype, _ := client.Type(ctx, key).Result()
		ttl, _ := client.TTL(ctx, key).Result()

		info := keyInfo{
			Key:  key,
			Type: ktype,
			TTL:  int64(ttl.Seconds()),
		}
		keyInfos = append(keyInfos, info)
	}

	return keyInfos, nil
}

func (d *RedisDriver) GetKeys(ctx context.Context, pattern string, count int64) ([]string, error) {
	client := d.client.(*redis.Client)
	var keys []string
	var cursor uint64

	if count <= 0 {
		count = 100
	}

	for {
		result, next, err := client.Scan(ctx, cursor, pattern, count).Result()
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		keys = append(keys, result...)
		cursor = next
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

func (d *RedisDriver) GetKeyInfo(ctx context.Context, key string) (*keyInfo, error) {
	client := d.client.(*redis.Client)

	ktype, err := client.Type(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get key type: %w", err)
	}

	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get key TTL: %w", err)
	}

	info := &keyInfo{
		Key:  key,
		Type: ktype,
		TTL:  int64(ttl.Seconds()),
	}

	switch ktype {
	case "string":
		val, err := client.Get(ctx, key).Result()
		if err == nil {
			info.Encoding = "string"
			info.MemoryUse = int64(len(val))
		}
	case "hash":
		val, err := client.HGetAll(ctx, key).Result()
		if err == nil {
			info.Encoding = "hash"
			info.MemoryUse = int64(len(val) * 50)
		}
	case "list":
		val, err := client.LRange(ctx, key, 0, -1).Result()
		if err == nil {
			info.Encoding = "list"
			info.MemoryUse = int64(len(val) * 50)
		}
	case "set":
		val, err := client.SMembers(ctx, key).Result()
		if err == nil {
			info.Encoding = "set"
			info.MemoryUse = int64(len(val) * 50)
		}
	case "zset":
		val, err := client.ZRangeWithScores(ctx, key, 0, -1).Result()
		if err == nil {
			info.Encoding = "zset"
			info.MemoryUse = int64(len(val) * 50)
		}
	}

	return info, nil
}

func (d *RedisDriver) GetServerInfo(ctx context.Context) (*serverInfo, error) {
	client := d.client.(*redis.Client)

	info, err := client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	infoMap := parseInfo(info)
	dbSize, _ := client.DBSize(ctx).Result()

	result := &serverInfo{
		Version:    getInfoValue(infoMap, "redis_version"),
		Mode:       getInfoValue(infoMap, "redis_mode"),
		DBSize:     dbSize,
		MemoryUsed: parseInt(getInfoValue(infoMap, "used_memory")),
		CPUUsed:    parseFloat(getInfoValue(infoMap, "used_cpu_user_children")),
		Clients:    parseInt(getInfoValue(infoMap, "connected_clients")),
		Uptime:     parseInt(getInfoValue(infoMap, "uptime_in_seconds")),
		Role:       getInfoValue(infoMap, "role"),
		Cluster:    getInfoValue(infoMap, "cluster_enabled") == "1",
	}

	return result, nil
}

func (d *RedisDriver) GetACL(ctx context.Context) (map[string]any, error) {
	client := d.client.(*redis.Client)

	result, err := client.Do(ctx, "ACL", "LIST").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get ACL: %w", err)
	}

	acl := make(map[string]any)
	if strSlice, ok := result.([]interface{}); ok {
		for i := 0; i < len(strSlice)-1; i += 2 {
			if key, ok := strSlice[i].(string); ok {
				if val, ok := strSlice[i+1].(string); ok {
					acl[key] = val
				}
			}
		}
	}

	return acl, nil
}

func (d *RedisDriver) GetMemoryStats(ctx context.Context) (map[string]any, error) {
	client := d.client.(*redis.Client)

	result, err := client.Do(ctx, "MEMORY", "STATS").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	stats := make(map[string]any)
	if pairs, ok := result.([]interface{}); ok {
		for i := 0; i < len(pairs)-1; i += 2 {
			if key, ok := pairs[i].(string); ok {
				stats[key] = pairs[i+1]
			}
		}
	}

	return stats, nil
}

func (d *RedisDriver) GetClients(ctx context.Context) ([]map[string]any, error) {
	client := d.client.(*redis.Client)

	result, err := client.Do(ctx, "CLIENT", "LIST").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %w", err)
	}

	var clients []map[string]any
	if str, ok := result.(string); ok {
		for _, line := range splitLines(str) {
			if line == "" {
				continue
			}
			client := parseClientLine(line)
			if client != nil {
				clients = append(clients, client)
			}
		}
	}

	return clients, nil
}

func parseInfo(info string) map[string]string {
	result := make(map[string]string)

	for _, line := range splitLines(info) {
		line = trimSpace(line)
		if line == "" || startsWith(line, "#") {
			continue
		}

		colonIdx := -1
		for i, c := range line {
			if c == ':' {
				colonIdx = i
				break
			}
		}

		if colonIdx > 0 {
			key := line[:colonIdx]
			value := line[colonIdx+1:]
			result[key] = value
		}
	}

	return result
}

func getInfoValue(info map[string]string, key string) string {
	return info[key]
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func trimSpace(s string) string {
	start := 0
	end := len(s) - 1

	for start <= end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end >= start && (s[end] == ' ' || s[end] == '\t' || s[end] == '\r' || s[end] == '\n') {
		end--
	}

	return s[start : end+1]
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func parseClientLine(line string) map[string]any {
	client := make(map[string]any)
	parts := splitAt(line, " ")
	for _, part := range parts {
		kv := splitAt(part, "=")
		if len(kv) == 2 {
			client[kv[0]] = kv[1]
		}
	}
	if len(client) > 0 {
		return client
	}
	return nil
}

func getKeyType(client *redis.Client, ctx context.Context, key string) (string, error) {
	return client.Type(ctx, key).Result()
}

func getKeyTTL(client *redis.Client, ctx context.Context, key string) (time.Duration, error) {
	return client.TTL(ctx, key).Result()
}
