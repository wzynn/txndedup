# TxnDedup - Transaction Deduplication Library

一个高性能的交易去重检测库，支持多种存储后端和灵活的风险规则配置。

## 特性

- 🚀 **高性能**：支持内存和Redis存储，毫秒级响应
- 🔧 **灵活配置**：可自定义指纹生成和风险规则
- 📊 **多级风险评估**：LOW/MEDIUM/HIGH三级风险
- 🔄 **实时检测**：支持实时重复交易检测
- 🛡️ **生产就绪**：完整的错误处理和日志记录
- 📈 **高并发**：支持大规模并发场景

## 安装

```bash
go get github.com/yourusername/txndedup
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/yourusername/txndedup"
)

func main() {
    // 创建检测器
    config := txndedup.DefaultConfig()
    detector, err := txndedup.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer detector.Close()
    
    // 检测重复交易
    request := &txndedup.TransactionRequest{
        FromAccount:  "account_001",
        ToAccount:    "account_002", 
        Amount:       100.00,
        Currency:     "USD",
        BusinessType: "transfer",
        UserIP:       "192.168.1.1",
        DeviceID:     "device_001",
    }
    
    result, err := detector.CheckDuplicate(context.Background(), request)
    if err != nil {
        log.Fatal(err)
    }
    
    switch result.SuggestionAction {
    case txndedup.ActionBlock:
        fmt.Println("❌ 交易被阻止:", result.Message)
    case txndedup.ActionWarn:
        fmt.Println("⚠️ 交易警告:", result.Message)
    case txndedup.ActionAllow:
        fmt.Println("✅ 交易允许")
    }
}
```

## 配置示例

### 使用Redis存储
```go
config := txndedup.DefaultConfig()
config.StorageType = "redis"
config.RedisConfig = &txndedup.RedisConfig{
    Address:  "localhost:6379",
    Password: "",
    DB:       0,
    KeyPrefix: "txndedup:",
}

detector, err := txndedup.New(config)
```

### 自定义风险规则
```go
config := txndedup.DefaultConfig()
config.RiskRules = []txndedup.RiskRule{
    {
        Name:       "custom_rule",
        TimeWindow: 1 * time.Minute,
        MaxCount:   0,
        RiskLevel:  txndedup.RiskLevelHigh,
        Action:     txndedup.ActionBlock,
    },
}
```

## API 文档

### 核心接口

#### CheckDuplicate
检测重复交易
```go
result, err := detector.CheckDuplicate(ctx, request)
```

#### RecordTransaction
记录交易
```go
err := detector.RecordTransaction(ctx, record)
```

### 响应结果
```go
type DuplicateCheckResult struct {
    IsDuplicate         bool                 // 是否重复
    SimilarTransactions []*TransactionRecord // 相似交易
    RiskLevel          RiskLevel            // 风险级别
    SuggestionAction   SuggestionAction     // 建议操作
    Message            string               // 提示消息
    Fingerprint        string               // 交易指纹
}
```

## 使用场景

- 💳 **支付系统**：防止重复支付
- 🏦 **银行转账**：避免重复转账
- 🛒 **电商订单**：重复下单检测
- 📱 **移动支付**：网络重试保护

## 性能基准
