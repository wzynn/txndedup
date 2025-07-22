package txndedup

import (
	"github.com/sirupsen/logrus"
	"time"
)

// Config 检测器配置
type Config struct {
	// 基础配置
	TimeWindow       time.Duration `json:"time_window"`         // 时间窗口
	CleanupInterval  time.Duration `json:"cleanup_interval"`    // 清理间隔
	MaxRecordsPerKey int           `json:"max_records_per_key"` // 每个指纹最大记录数

	// 指纹配置
	FingerprintConfig FingerprintConfig `json:"fingerprint_config"`

	// 风险规则
	RiskRules []RiskRule `json:"risk_rules"`

	// 日志配置
	Logger   logrus.FieldLogger `json:"-"`
	LogLevel logrus.Level       `json:"log_level"`

	// 存储配置
	StorageType string       `json:"storage_type"` // "memory" | "redis"
	RedisConfig *RedisConfig `json:"redis_config,omitempty"`

	// 性能配置
	EnableAsync    bool `json:"enable_async"`     // 异步处理
	WorkerPoolSize int  `json:"worker_pool_size"` // 工作池大小
}

// RedisConfig Redis配置
type RedisConfig struct {
	Address      string        `json:"address"`
	Password     string        `json:"password"`
	DB           int           `json:"db"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	KeyPrefix    string        `json:"key_prefix"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		TimeWindow:       5 * time.Minute,
		CleanupInterval:  1 * time.Minute,
		MaxRecordsPerKey: 100,

		FingerprintConfig: FingerprintConfig{
			IncludeFromAccount:  true,
			IncludeToAccount:    true,
			IncludeAmount:       true,
			IncludeCurrency:     true,
			IncludeBusinessType: true,
			IncludeChannel:      false,
			AmountPrecision:     2, // 精确到分
		},

		RiskRules: []RiskRule{
			{
				Name:            "pending_duplicate",
				TimeWindow:      30 * time.Minute,
				MaxCount:        0, // 0表示不允许任何pending状态的重复
				RiskLevel:       RiskLevelHigh,
				Action:          ActionBlock,
				CheckSameIP:     false,
				CheckSameDevice: false,
				CheckStatus:     []TransactionStatus{StatusPending},
			},
			{
				Name:            "rapid_duplicate",
				TimeWindow:      30 * time.Second,
				MaxCount:        0,
				RiskLevel:       RiskLevelHigh,
				Action:          ActionWarn,
				CheckSameIP:     true,
				CheckSameDevice: true,
				CheckStatus:     []TransactionStatus{StatusSuccess, StatusPending},
			},
			{
				Name:            "frequent_duplicate",
				TimeWindow:      2 * time.Minute,
				MaxCount:        1,
				RiskLevel:       RiskLevelMedium,
				Action:          ActionWarn,
				CheckSameIP:     false,
				CheckSameDevice: false,
				CheckStatus:     []TransactionStatus{StatusSuccess},
			},
			{
				Name:            "recent_duplicate",
				TimeWindow:      5 * time.Minute,
				MaxCount:        2,
				RiskLevel:       RiskLevelLow,
				Action:          ActionAllow,
				CheckSameIP:     false,
				CheckSameDevice: false,
				CheckStatus:     []TransactionStatus{StatusSuccess},
			},
		},

		Logger:         logrus.StandardLogger(),
		LogLevel:       logrus.InfoLevel,
		StorageType:    "memory",
		EnableAsync:    false,
		WorkerPoolSize: 10,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.TimeWindow <= 0 {
		return ErrInvalidTimeWindow
	}

	if c.CleanupInterval <= 0 {
		return ErrInvalidCleanupInterval
	}

	if c.StorageType == "redis" && c.RedisConfig == nil {
		return ErrMissingRedisConfig
	}

	return nil
}
