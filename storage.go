package txndedup

import (
	"context"
	"time"
)

// Storage 存储接口
type Storage interface {
	// 存储交易记录
	Store(ctx context.Context, fingerprint string, record *TransactionRecord) error

	// 获取相似交易
	GetSimilar(ctx context.Context, fingerprint string, timeWindow time.Duration) ([]*TransactionRecord, error)

	// 清理过期记录
	Cleanup(ctx context.Context, timeWindow time.Duration) error

	// 关闭存储
	Close() error
}

// StorageFactory 存储工厂
type StorageFactory struct{}

// NewStorage 创建存储实例
func (sf *StorageFactory) NewStorage(config *Config) (Storage, error) {
	switch config.StorageType {
	case "memory":
		return NewMemoryStorage(config), nil
	case "redis":
		return NewRedisStorage(config.RedisConfig)
	default:
		return nil, ErrUnsupportedStorageType
	}
}
