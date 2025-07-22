package txndedup

import (
	"crypto/md5"
	"fmt"
	"math"
	"sort"
	"strings"
)

// FingerprintGenerator 指纹生成器
type FingerprintGenerator struct {
	config FingerprintConfig
}

// NewFingerprintGenerator 创建指纹生成器
func NewFingerprintGenerator(config FingerprintConfig) *FingerprintGenerator {
	return &FingerprintGenerator{
		config: config,
	}
}

// Generate 生成交易指纹
func (fg *FingerprintGenerator) Generate(request *TransactionRequest) string {
	var components []string

	if fg.config.IncludeFromAccount {
		components = append(components, "from:"+request.FromAccount)
	}

	if fg.config.IncludeToAccount {
		components = append(components, "to:"+request.ToAccount)
	}

	if fg.config.IncludeAmount {
		amount := fg.normalizeAmount(request.Amount)
		components = append(components, fmt.Sprintf("amount:%.2f", amount))
	}

	if fg.config.IncludeCurrency {
		components = append(components, "currency:"+strings.ToUpper(request.Currency))
	}

	if fg.config.IncludeBusinessType {
		components = append(components, "type:"+request.BusinessType)
	}

	if fg.config.IncludeChannel {
		components = append(components, "channel:"+request.Channel)
	}

	// 排序保证一致性
	sort.Strings(components)

	// 生成指纹
	data := strings.Join(components, "|")
	hash := md5.Sum([]byte(data))

	return fmt.Sprintf("%x", hash)
}

// normalizeAmount 标准化金额
func (fg *FingerprintGenerator) normalizeAmount(amount float64) float64 {
	if fg.config.AmountPrecision <= 0 {
		return amount
	}

	multiplier := math.Pow(10, float64(fg.config.AmountPrecision))
	return math.Round(amount*multiplier) / multiplier
}

// GenerateFromRecord 从记录生成指纹
func (fg *FingerprintGenerator) GenerateFromRecord(record *TransactionRecord) string {
	request := &TransactionRequest{
		FromAccount:  record.FromAccount,
		ToAccount:    record.ToAccount,
		Amount:       record.Amount,
		Currency:     record.Currency,
		BusinessType: record.BusinessType,
		Channel:      record.Channel,
	}

	return fg.Generate(request)
}
