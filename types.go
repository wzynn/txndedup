package txndedup

import (
	"time"
)

// TransactionRequest 交易请求
type TransactionRequest struct {
	FromAccount  string                 `json:"from_account"`
	ToAccount    string                 `json:"to_account"`
	Amount       float64                `json:"amount"`
	Currency     string                 `json:"currency"`
	BusinessType string                 `json:"business_type"`
	Channel      string                 `json:"channel"`
	UserIP       string                 `json:"user_ip"`
	DeviceID     string                 `json:"device_id"`
	UserAgent    string                 `json:"user_agent"`
	Extra        map[string]interface{} `json:"extra,omitempty"` // 扩展字段
}

// TransactionRecord 交易记录
type TransactionRecord struct {
	TransactionID string                 `json:"transaction_id"`
	Fingerprint   string                 `json:"fingerprint"`
	FromAccount   string                 `json:"from_account"`
	ToAccount     string                 `json:"to_account"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	BusinessType  string                 `json:"business_type"`
	Channel       string                 `json:"channel"`
	Status        TransactionStatus      `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	UserIP        string                 `json:"user_ip"`
	DeviceID      string                 `json:"device_id"`
	UserAgent     string                 `json:"user_agent"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

// TransactionStatus 交易状态
type TransactionStatus string

const (
	StatusPending   TransactionStatus = "PENDING"
	StatusSuccess   TransactionStatus = "SUCCESS"
	StatusFailed    TransactionStatus = "FAILED"
	StatusCancelled TransactionStatus = "CANCELLED"
)

// DuplicateCheckResult 重复检测结果
type DuplicateCheckResult struct {
	IsDuplicate         bool                 `json:"is_duplicate"`
	SimilarTransactions []*TransactionRecord `json:"similar_transactions"`
	RiskLevel           RiskLevel            `json:"risk_level"`
	SuggestionAction    SuggestionAction     `json:"suggestion_action"`
	Message             string               `json:"message"`
	Fingerprint         string               `json:"fingerprint"`
	CheckedAt           time.Time            `json:"checked_at"`
}

// RiskLevel 风险级别
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "LOW"
	RiskLevelMedium RiskLevel = "MEDIUM"
	RiskLevelHigh   RiskLevel = "HIGH"
)

// SuggestionAction 建议操作
type SuggestionAction string

const (
	ActionAllow SuggestionAction = "ALLOW"
	ActionWarn  SuggestionAction = "WARN"
	ActionBlock SuggestionAction = "BLOCK"
)

// FingerprintConfig 指纹配置
type FingerprintConfig struct {
	IncludeFromAccount  bool `json:"include_from_account"`
	IncludeToAccount    bool `json:"include_to_account"`
	IncludeAmount       bool `json:"include_amount"`
	IncludeCurrency     bool `json:"include_currency"`
	IncludeBusinessType bool `json:"include_business_type"`
	IncludeChannel      bool `json:"include_channel"`
	AmountPrecision     int  `json:"amount_precision"` // 金额精度，0表示精确匹配
}

// RiskRule 风险规则
type RiskRule struct {
	Name            string              `json:"name"`
	TimeWindow      time.Duration       `json:"time_window"`
	MaxCount        int                 `json:"max_count"`
	RiskLevel       RiskLevel           `json:"risk_level"`
	Action          SuggestionAction    `json:"action"`
	CheckSameIP     bool                `json:"check_same_ip"`
	CheckSameDevice bool                `json:"check_same_device"`
	CheckStatus     []TransactionStatus `json:"check_status"`
}
