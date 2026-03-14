package redis


import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"tablepro/internal/connection"
)

func New() *RedisDriver {
	return &RedisDriver{}
}

func (d *RedisDriver) Connect(ctx context.Context, config *connection.DatabaseConnection, password string) error {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	dbIndex := 0
	if config.Database != "" {
		if parsed, err := strconv.Atoi(config.Database); err == nil {
			dbIndex = parsed
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbIndex,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	d.client = client
	d.config = config
	d.Database = dbIndex

	return nil
}

func (d *RedisDriver) Execute(ctx context.Context, command string, args ...interface{}) (interface{}, error) {
	client := d.client.(*redis.Client)

	parts := parseCommand(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid command")
	}

	allArgs := []interface{}{parts[0]}
	allArgs = append(allArgs, args...)
	result, err := client.Do(ctx, allArgs...).Result()
	if err != nil {
		return nil, fmt.Errorf("command failed: %w", err)
	}

	return result, nil
}

func parseCommand(cmd string) []string {
	var parts []string
	var current []rune
	inQuote := false
	quoteChar := rune(0)

	for _, c := range cmd {
		if !inQuote && (c == '"' || c == '\'') {
			inQuote = true
			quoteChar = c
		} else if inQuote && c == quoteChar {
			inQuote = false
			quoteChar = 0
		} else if !inQuote && c == ' ' {
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = nil
			}
		} else {
			current = append(current, c)
		}
	}

	if len(current) > 0 {
		parts = append(parts, string(current))
	}

	return parts
}

func (d *RedisDriver) Get(ctx context.Context, key string) (string, error) {
	client := d.client.(*redis.Client)
	return client.Get(ctx, key).Result()
}

func (d *RedisDriver) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	client := d.client.(*redis.Client)
	return client.Set(ctx, key, value, ttl).Err()
}

func (d *RedisDriver) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	client := d.client.(*redis.Client)
	return client.HGetAll(ctx, key).Result()
}

func (d *RedisDriver) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	client := d.client.(*redis.Client)
	return client.LRange(ctx, key, start, stop).Result()
}

func (d *RedisDriver) SMembers(ctx context.Context, key string) ([]string, error) {
	client := d.client.(*redis.Client)
	return client.SMembers(ctx, key).Result()
}

func (d *RedisDriver) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	client := d.client.(*redis.Client)
	return client.ZRangeWithScores(ctx, key, start, stop).Result()
}

func (d *RedisDriver) Type(ctx context.Context, key string) (string, error) {
	client := d.client.(*redis.Client)
	return client.Type(ctx, key).Result()
}

func (d *RedisDriver) TTL(ctx context.Context, key string) (time.Duration, error) {
	client := d.client.(*redis.Client)
	return client.TTL(ctx, key).Result()
}

func (d *RedisDriver) Exists(ctx context.Context, keys ...string) (int64, error) {
	client := d.client.(*redis.Client)
	return client.Exists(ctx, keys...).Result()
}

func (d *RedisDriver) Del(ctx context.Context, keys ...string) (int64, error) {
	client := d.client.(*redis.Client)
	return client.Del(ctx, keys...).Result()
}

func (d *RedisDriver) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	client := d.client.(*redis.Client)
	return client.Expire(ctx, key, ttl).Result()
}

func (d *RedisDriver) Scan(ctx context.Context, cursor uint64, match string, count int64) (*scanResult, error) {
	client := d.client.(*redis.Client)
	keys, next, err := client.Scan(ctx, cursor, match, count).Result()
	if err != nil {
		return nil, err
	}

	return &scanResult{
		Cursor: int64(next),
		Keys:   keys,
	}, nil
}

func (d *RedisDriver) DBSize(ctx context.Context) (int64, error) {
	client := d.client.(*redis.Client)
	return client.DBSize(ctx).Result()
}

func (d *RedisDriver) Info(ctx context.Context, section string) (string, error) {
	client := d.client.(*redis.Client)
	return client.Info(ctx, section).Result()
}

func (d *RedisDriver) ConfigGet(ctx context.Context, parameter string) (map[string]string, error) {
	client := d.client.(*redis.Client)
	return client.ConfigGet(ctx, parameter).Result()
}

func (d *RedisDriver) Ping(ctx context.Context) error {
	client := d.client.(*redis.Client)
	return client.Ping(ctx).Err()
}

func (d *RedisDriver) Close() error {
	client := d.client.(*redis.Client)
	return client.Close()
}

func (d *RedisDriver) GetConfig() *connection.DatabaseConnection {
	return d.config
}

func (d *RedisDriver) IsConnected() bool {
	if d.client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client := d.client.(*redis.Client)
	err := client.Ping(ctx).Err()
	return err == nil
}

func (d *RedisDriver) GetDatabase() int {
	return d.Database
}

func (d *RedisDriver) GetClient() *redis.Client {
	if d.client == nil {
		return nil
	}
	return d.client.(*redis.Client)
}

func (d *RedisDriver) ExecuteQuery(ctx context.Context, query string) (*queryResult, error) {
	parts := parseCommand(query)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid command")
	}

	client := d.client.(*redis.Client)
	allArgs := make([]interface{}, len(parts))
	for i, p := range parts {
		allArgs[i] = p
	}
	result, err := client.Do(ctx, allArgs...).Result()
	if err != nil {
		return nil, fmt.Errorf("command failed: %w", err)
	}

	return d.formatResult(result), nil
}

func (d *RedisDriver) formatResult(result interface{}) *queryResult {
	r := &queryResult{
		Columns: []string{"result"},
		Rows:    [][]any{},
	}

	switch v := result.(type) {
	case string:
		r.Rows = append(r.Rows, []any{v})
	case int64:
		r.Rows = append(r.Rows, []any{v})
	case float64:
		r.Rows = append(r.Rows, []any{v})
	case bool:
		r.Rows = append(r.Rows, []any{v})
	case []string:
		for _, s := range v {
			r.Rows = append(r.Rows, []any{s})
		}
	case []interface{}:
		for _, item := range v {
			r.Rows = append(r.Rows, []any{item})
		}
	case map[string]string:
		r.Columns = []string{"key", "value"}
		for k, val := range v {
			r.Rows = append(r.Rows, []any{k, val})
		}
	case map[string]interface{}:
		r.Columns = []string{"key", "value"}
		for k, val := range v {
			r.Rows = append(r.Rows, []any{k, val})
		}
	case redis.Z:
		r.Columns = []string{"member", "score"}
		r.Rows = append(r.Rows, []any{v.Member, v.Score})
	case []redis.Z:
		r.Columns = []string{"member", "score"}
		for _, z := range v {
			r.Rows = append(r.Rows, []any{z.Member, z.Score})
		}
	default:
		r.Rows = append(r.Rows, []any{fmt.Sprintf("%v", v)})
	}

	return r
}

func parseVersion(info string) string {
	for _, line := range splitLines(info) {
		if contains(line, "redis_version:") {
			parts := splitAt(line, ":")
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}
	return ""
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitAt(s, sep string) []string {
	var parts []string
	current := ""
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			parts = append(parts, current)
			current = ""
			i += len(sep) - 1
		} else {
			current += string(s[i])
		}
	}
	parts = append(parts, current)
	return parts
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func parseInt(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

