package tests

import (
	"context"
	"github.com/wzynn/txndedup"
	"testing"
	"time"
)

func TestDetector_CheckDuplicate(t *testing.T) {
	config := txndedup.DefaultConfig()
	config.TimeWindow = 1 * time.Minute

	detector, err := txndedup.New(config)
	if err != nil {
		t.Fatal(err)
	}
	defer detector.Close()

	ctx := context.Background()

	request := &txndedup.TransactionRequest{
		FromAccount:  "test_001",
		ToAccount:    "test_002",
		Amount:       50.00,
		Currency:     "USD",
		BusinessType: "transfer",
		UserIP:       "127.0.0.1",
		DeviceID:     "test_device",
	}

	// 第一次检测，应该没有重复
	result, err := detector.CheckDuplicate(ctx, request)
	if err != nil {
		t.Fatal(err)
	}

	if result.IsDuplicate {
		t.Error("第一次检测不应该是重复")
	}

	// 记录交易
	record := &txndedup.TransactionRecord{
		FromAccount:  request.FromAccount,
		ToAccount:    request.ToAccount,
		Amount:       request.Amount,
		Currency:     request.Currency,
		BusinessType: request.BusinessType,
		Status:       txndedup.StatusSuccess,
		UserIP:       request.UserIP,
		DeviceID:     request.DeviceID,
	}

	err = detector.RecordTransaction(ctx, record)
	if err != nil {
		t.Fatal(err)
	}

	// 第二次检测，应该检测到重复
	result2, err := detector.CheckDuplicate(ctx, request)
	if err != nil {
		t.Fatal(err)
	}

	if !result2.IsDuplicate {
		t.Error("第二次检测应该是重复")
	}

	if len(result2.SimilarTransactions) != 1 {
		t.Errorf("应该有1个相似交易，实际有%d个", len(result2.SimilarTransactions))
	}
}
