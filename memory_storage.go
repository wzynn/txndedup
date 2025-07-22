package txndedup

import (
	"context"
	"sync"
	"time"
)

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	records map[string][]*TransactionRecord
	mu      sync.RWMutex
	config  *Config
}

// NewMemoryStorage 创建内存存储
func NewMemoryStorage(config *Config) *MemoryStorage {
	storage := &MemoryStorage{
		records: make(map[string][]*TransactionRecord),
		config:  config,
	}

	// 启动清理协程
	go storage.startCleanup()

	return storage
}

// Store 存储交易记录
func (ms *MemoryStorage) Store(ctx context.Context, fingerprint string, record *TransactionRecord) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.records[fingerprint]; !exists {
		ms.records[fingerprint] = make([]*TransactionRecord, 0)
	}

	ms.records[fingerprint] = append(ms.records[fingerprint], record)

	// 限制每个指纹的记录数量
	if len(ms.records[fingerprint]) > ms.config.MaxRecordsPerKey {
		// 保留最新的记录
		ms.records[fingerprint] = ms.records[fingerprint][1:]
	}

	return nil
}

// GetSimilar 获取相似交易
func (ms *MemoryStorage) GetSimilar(ctx context.Context, fingerprint string, timeWindow time.Duration) ([]*TransactionRecord, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	records, exists := ms.records[fingerprint]
	if !exists {
		return nil, nil
	}

	cutoffTime := time.Now().Add(-timeWindow)
	var similarTx []*TransactionRecord

	for _, record := range records {
		if record.CreatedAt.After(cutoffTime) {
			similarTx = append(similarTx, record)
		}
	}

	return similarTx, nil
}

// Cleanup 清理过期记录
func (ms *MemoryStorage) Cleanup(ctx context.Context, timeWindow time.Duration) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	cutoffTime := time.Now().Add(-timeWindow)

	for fingerprint, records := range ms.records {
		var validRecords []*TransactionRecord

		for _, record := range records {
			if record.CreatedAt.After(cutoffTime) {
				validRecords = append(validRecords, record)
			}
		}

		if len(validRecords) == 0 {
			delete(ms.records, fingerprint)
		} else {
			ms.records[fingerprint] = validRecords
		}
	}

	return nil
}

// Close 关闭存储
func (ms *MemoryStorage) Close() error {
	return nil
}

// startCleanup 启动清理协程
func (ms *MemoryStorage) startCleanup() {
	ticker := time.NewTicker(ms.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := ms.Cleanup(context.Background(), ms.config.TimeWindow); err != nil {
			ms.config.Logger.Errorf("cleanup failed: %v", err)
		}
	}
}
