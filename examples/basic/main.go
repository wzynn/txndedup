package main

import (
	"context"
	"fmt"
	"log"
)

func main() {
	// 创建配置
	config := txndedup.DefaultConfig()

	// 创建检测器
	detector, err := txndedup.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer detector.Close()

	ctx := context.Background()

	// 模拟交易请求
	request := &txndedup.TransactionRequest{
		FromAccount:  "account_001",
		ToAccount:    "account_002",
		Amount:       100.00,
		Currency:     "USD",
		BusinessType: "transfer",
		Channel:      "web",
		UserIP:       "192.168.1.1",
		DeviceID:     "device_001",
	}

	// 第一次检测
	result, err := detector.CheckDuplicate(ctx, request)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("第一次检测结果: %+v\n", result)

	// 记录交易
	record := &txndedup.TransactionRecord{
		TransactionID: "tx_001",
		FromAccount:   request.FromAccount,
		ToAccount:     request.ToAccount,
		Amount:        request.Amount,
		Currency:      request.Currency,
		BusinessType:  request.BusinessType,
		Channel:       request.Channel,
		Status:        txndedup.StatusSuccess,
		UserIP:        request.UserIP,
		DeviceID:      request.DeviceID,
	}

	if err := detector.RecordTransaction(ctx, record); err != nil {
		log.Fatal(err)
	}

	// 第二次检测（应该检测到重复）
	result2, err := detector.CheckDuplicate(ctx, request)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("第二次检测结果: %+v\n", result2)
}
