package txndedup

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStorage Redis存储实现
type RedisStorage struct {
	client    *redis.Client
	keyPrefix string
}

// NewRedisStorage 创建Redis存储
func NewRedisStorage(config *RedisConfig) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.Address,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisStorage{
		client:    rdb,
		keyPrefix: config.KeyPrefix,
	}, nil
}

// Store 存储交易记录
func (rs *RedisStorage) Store(ctx context.Context, fingerprint string, record *TransactionRecord) error {
	key := rs.buildKey(fingerprint)

	// 序列化记录
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal record failed: %w", err)
	}

	// 使用有序集合存储，score为时间戳
	score := float64(record.CreatedAt.Unix())

	pipe := rs.client.Pipeline()
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: data,
	})

	// 设置过期时间
	pipe.Expire(ctx, key, 30*time.Minute)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("store record failed: %w", err)
	}

	return nil
}

// GetSimilar 获取相似交易
func (rs *RedisStorage) GetSimilar(ctx context.Context, fingerprint string, timeWindow time.Duration) ([]*TransactionRecord, error) {
	key := rs.buildKey(fingerprint)
	cutoffTime := time.Now().Add(-timeWindow)

	// 从有序集合中获取指定时间范围内的记录
	result, err := rs.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", cutoffTime.Unix()),
		Max: "+inf",
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("get similar records failed: %w", err)
	}

	var records []*TransactionRecord
	for _, data := range result {
		var record TransactionRecord
		if err := json.Unmarshal([]byte(data), &record); err != nil {
			continue // 跳过无法解析的记录
		}
		records = append(records, &record)
	}

	return records, nil
}

// Cleanup 清理过期记录
func (rs *RedisStorage) Cleanup(ctx context.Context, timeWindow time.Duration) error {
	// Redis会自动过期，这里可以做额外的清理
	cutoffTime := time.Now().Add(-timeWindow)

	// 扫描所有相关的key
	iter := rs.client.Scan(ctx, 0, rs.keyPrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// 删除过期的记录
		rs.client.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", cutoffTime.Unix()))
	}

	return iter.Err()
}

// Close 关闭存储
func (rs *RedisStorage) Close() error {
	return rs.client.Close()
}

// buildKey 构建Redis key
func (rs *RedisStorage) buildKey(fingerprint string) string {
	return rs.keyPrefix + "tx:" + fingerprint
}
