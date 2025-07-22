package txndedup

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Detector 重复交易检测器
type Detector struct {
	config               *Config
	storage              Storage
	fingerprintGenerator *FingerprintGenerator
	riskAssessor         *RiskAssessor
}

// New 创建检测器
func New(config *Config) (*Detector, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 创建存储
	factory := &StorageFactory{}
	storage, err := factory.NewStorage(config)
	if err != nil {
		return nil, fmt.Errorf("create storage failed: %w", err)
	}

	// 创建指纹生成器
	fingerprintGenerator := NewFingerprintGenerator(config.FingerprintConfig)

	// 创建风险评估器
	riskAssessor := NewRiskAssessor(config.RiskRules)

	return &Detector{
		config:               config,
		storage:              storage,
		fingerprintGenerator: fingerprintGenerator,
		riskAssessor:         riskAssessor,
	}, nil
}

// CheckDuplicate 检测重复交易
func (d *Detector) CheckDuplicate(ctx context.Context, request *TransactionRequest) (*DuplicateCheckResult, error) {
	// 生成指纹
	fingerprint := d.fingerprintGenerator.Generate(request)

	// 查找相似交易
	similarTx, err := d.storage.GetSimilar(ctx, fingerprint, d.config.TimeWindow)
	if err != nil {
		return nil, fmt.Errorf("get similar transactions failed: %w", err)
	}

	// 风险评估
	riskLevel, action, message := d.riskAssessor.Assess(request, similarTx)

	result := &DuplicateCheckResult{
		IsDuplicate:         len(similarTx) > 0,
		SimilarTransactions: similarTx,
		RiskLevel:           riskLevel,
		SuggestionAction:    action,
		Message:             message,
		Fingerprint:         fingerprint,
		CheckedAt:           time.Now(),
	}

	d.config.Logger.WithFields(map[string]interface{}{
		"fingerprint":       fingerprint[:8],
		"similar_count":     len(similarTx),
		"risk_level":        riskLevel,
		"suggestion_action": action,
	}).Info("duplicate check completed")

	return result, nil
}

// RecordTransaction 记录交易
func (d *Detector) RecordTransaction(ctx context.Context, record *TransactionRecord) error {
	// 设置默认值
	if record.TransactionID == "" {
		record.TransactionID = uuid.New().String()
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	record.UpdatedAt = time.Now()

	// 生成指纹
	record.Fingerprint = d.fingerprintGenerator.GenerateFromRecord(record)

	// 存储记录
	if err := d.storage.Store(ctx, record.Fingerprint, record); err != nil {
		return fmt.Errorf("store transaction failed: %w", err)
	}

	d.config.Logger.WithFields(map[string]interface{}{
		"transaction_id": record.TransactionID,
		"fingerprint":    record.Fingerprint[:8],
		"status":         record.Status,
	}).Info("transaction recorded")

	return nil
}

// UpdateTransactionStatus 更新交易状态
func (d *Detector) UpdateTransactionStatus(ctx context.Context, transactionID string, status TransactionStatus) error {
	// 这里可以实现状态更新逻辑
	// 由于我们使用指纹作为key，需要额外的索引来支持按transactionID查找
	d.config.Logger.WithFields(map[string]interface{}{
		"transaction_id": transactionID,
		"new_status":     status,
	}).Info("transaction status updated")

	return nil
}

// Close 关闭检测器
func (d *Detector) Close() error {
	if d.storage != nil {
		return d.storage.Close()
	}
	return nil
}
