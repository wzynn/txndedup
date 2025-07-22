package txndedup

import (
	"fmt"
	"time"
)

// RiskAssessor 风险评估器
type RiskAssessor struct {
	rules []RiskRule
}

// NewRiskAssessor 创建风险评估器
func NewRiskAssessor(rules []RiskRule) *RiskAssessor {
	return &RiskAssessor{
		rules: rules,
	}
}

// Assess 评估风险
func (ra *RiskAssessor) Assess(request *TransactionRequest, similarTx []*TransactionRecord) (RiskLevel, SuggestionAction, string) {
	if len(similarTx) == 0 {
		return RiskLevelLow, ActionAllow, ""
	}

	// 按规则优先级评估
	for _, rule := range ra.rules {
		if ra.matchRule(rule, request, similarTx) {
			message := ra.generateMessage(rule, request, similarTx)
			return rule.RiskLevel, rule.Action, message
		}
	}

	return RiskLevelLow, ActionAllow, ""
}

// matchRule 匹配规则
func (ra *RiskAssessor) matchRule(rule RiskRule, request *TransactionRequest, similarTx []*TransactionRecord) bool {
	// 过滤时间窗口内的交易
	cutoffTime := time.Now().Add(-rule.TimeWindow)
	var matchingTx []*TransactionRecord

	for _, tx := range similarTx {
		if tx.CreatedAt.After(cutoffTime) {
			// 检查状态
			if len(rule.CheckStatus) > 0 {
				statusMatch := false
				for _, status := range rule.CheckStatus {
					if tx.Status == status {
						statusMatch = true
						break
					}
				}
				if !statusMatch {
					continue
				}
			}

			// 检查IP
			if rule.CheckSameIP && tx.UserIP != request.UserIP {
				continue
			}

			// 检查设备
			if rule.CheckSameDevice && tx.DeviceID != request.DeviceID {
				continue
			}

			matchingTx = append(matchingTx, tx)
		}
	}

	return len(matchingTx) > rule.MaxCount
}

// generateMessage 生成提示消息
func (ra *RiskAssessor) generateMessage(rule RiskRule, request *TransactionRequest, similarTx []*TransactionRecord) string {
	switch rule.Name {
	case "pending_duplicate":
		return "检测到您有一笔相同的交易正在处理中，请勿重复提交"

	case "rapid_duplicate":
		if len(similarTx) > 0 {
			timeDiff := time.Since(similarTx[len(similarTx)-1].CreatedAt)
			return fmt.Sprintf("检测到您在%.0f秒前刚完成了一笔相同的交易，请确认是否要继续", timeDiff.Seconds())
		}

	case "frequent_duplicate":
		return "检测到您最近有多笔相似交易，请确认交易信息"

	case "recent_duplicate":
		if len(similarTx) > 0 {
			timeDiff := time.Since(similarTx[len(similarTx)-1].CreatedAt)
			return fmt.Sprintf("温馨提示：您在%.0f分钟前有类似交易记录", timeDiff.Minutes())
		}
	}

	return "检测到相似交易"
}
