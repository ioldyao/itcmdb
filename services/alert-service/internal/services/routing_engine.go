package services

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/itcmdb/alert-service/internal/models"
	"gorm.io/gorm"
)

// RoutingEngine 路由引擎 - 基于标签的告警路由
type RoutingEngine struct {
	db    *gorm.DB
	cache *routingCache
	mu    sync.RWMutex
}

// routingCache 路由规则缓存
type routingCache struct {
	rules      []models.AlertRoutingRule
	lastUpdate time.Time
	ttl        time.Duration
	mu         sync.RWMutex
}

// RoutingResult 路由结果
type RoutingResult struct {
	ReceiverGroupIDs []int
	MatchedRules     []models.AlertRoutingRule
}

// NewRoutingEngine 创建路由引擎实例
func NewRoutingEngine(db *gorm.DB) *RoutingEngine {
	return &RoutingEngine{
		db: db,
		cache: &routingCache{
			rules:      []models.AlertRoutingRule{},
			lastUpdate: time.Time{},
			ttl:        5 * time.Minute, // 5分钟缓存TTL
		},
	}
}

// RouteAlert 根据标签路由告警到接收组
func (re *RoutingEngine) RouteAlert(labels map[string]string) (*RoutingResult, error) {
	// 获取路由规则（从缓存或数据库）
	rules, err := re.getRoutingRules()
	if err != nil {
		return nil, fmt.Errorf("failed to get routing rules: %w", err)
	}

	result := &RoutingResult{
		ReceiverGroupIDs: []int{},
		MatchedRules:     []models.AlertRoutingRule{},
	}

	// 按优先级排序规则（已在查询时排序）
	for _, rule := range rules {
		// 检查规则是否匹配
		matched, err := re.matchRule(&rule, labels)
		if err != nil {
			// 记录错误但继续处理其他规则
			fmt.Printf("Error matching rule %s: %v\n", rule.Name, err)
			continue
		}

		if matched {
			// 添加到匹配结果
			result.MatchedRules = append(result.MatchedRules, rule)

			// 添加接收组ID（如果存在）
			if rule.ReceiverGroupID != nil {
				result.ReceiverGroupIDs = append(result.ReceiverGroupIDs, *rule.ReceiverGroupID)
			}

			// 如果continue标志为false，停止匹配
			if !rule.Continue {
				break
			}
		}
	}

	// 如果没有匹配到任何规则，使用默认路由
	if len(result.ReceiverGroupIDs) == 0 {
		defaultGroupID, err := re.getDefaultReceiverGroup()
		if err == nil && defaultGroupID > 0 {
			result.ReceiverGroupIDs = append(result.ReceiverGroupIDs, defaultGroupID)
		}
	}

	return result, nil
}

// matchRule 检查规则是否匹配给定的标签
func (re *RoutingEngine) matchRule(rule *models.AlertRoutingRule, labels map[string]string) (bool, error) {
	if !rule.Enabled {
		return false, nil
	}

	// 如果matchers为空，根据match_type决定
	if len(rule.Matchers) == 0 {
		// 空matchers + match_type='all' 表示匹配所有
		if rule.MatchType == "all" {
			return true, nil
		}
		return false, nil
	}

	// 遍历所有matchers
	for key, expectedValue := range rule.Matchers {
		actualValue, exists := labels[key]

		// 如果标签不存在，不匹配
		if !exists {
			return false, nil
		}

		// 根据match_type进行匹配
		switch rule.MatchType {
		case "exact", "match":
			// 精确匹配
			expectedStr, ok := expectedValue.(string)
			if !ok {
				return false, fmt.Errorf("expected value for key %s is not a string", key)
			}
			if actualValue != expectedStr {
				return false, nil
			}

		case "regex", "match_re":
			// 正则匹配
			expectedStr, ok := expectedValue.(string)
			if !ok {
				return false, fmt.Errorf("expected value for key %s is not a string", key)
			}
			matched, err := regexp.MatchString(expectedStr, actualValue)
			if err != nil {
				return false, fmt.Errorf("invalid regex pattern %s: %w", expectedStr, err)
			}
			if !matched {
				return false, nil
			}

		case "all":
			// 匹配所有（已在上面处理空matchers的情况）
			return true, nil

		default:
			return false, fmt.Errorf("unknown match_type: %s", rule.MatchType)
		}
	}

	// 所有matchers都匹配
	return true, nil
}

// getRoutingRules 获取路由规则（带缓存）
func (re *RoutingEngine) getRoutingRules() ([]models.AlertRoutingRule, error) {
	re.cache.mu.RLock()

	// 检查缓存是否有效
	if time.Since(re.cache.lastUpdate) < re.cache.ttl && len(re.cache.rules) > 0 {
		rules := re.cache.rules
		re.cache.mu.RUnlock()
		return rules, nil
	}
	re.cache.mu.RUnlock()

	// 缓存过期或为空，从数据库加载
	re.cache.mu.Lock()
	defer re.cache.mu.Unlock()

	// 双重检查（防止并发加载）
	if time.Since(re.cache.lastUpdate) < re.cache.ttl && len(re.cache.rules) > 0 {
		return re.cache.rules, nil
	}

	var rules []models.AlertRoutingRule
	err := re.db.Where("enabled = ?", true).
		Order("priority DESC, id ASC").
		Preload("ReceiverGroup").
		Find(&rules).Error

	if err != nil {
		return nil, fmt.Errorf("failed to load routing rules: %w", err)
	}

	// 更新缓存
	re.cache.rules = rules
	re.cache.lastUpdate = time.Now()

	return rules, nil
}

// getDefaultReceiverGroup 获取默认接收组
func (re *RoutingEngine) getDefaultReceiverGroup() (int, error) {
	var rule models.AlertRoutingRule
	err := re.db.Where("enabled = ? AND match_type = ?", true, "all").
		Order("priority ASC").
		First(&rule).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}

	if rule.ReceiverGroupID != nil {
		return *rule.ReceiverGroupID, nil
	}

	return 0, nil
}

// InvalidateCache 使缓存失效
func (re *RoutingEngine) InvalidateCache() {
	re.cache.mu.Lock()
	defer re.cache.mu.Unlock()

	re.cache.rules = []models.AlertRoutingRule{}
	re.cache.lastUpdate = time.Time{}
}

// GetCacheStats 获取缓存统计信息
func (re *RoutingEngine) GetCacheStats() map[string]interface{} {
	re.cache.mu.RLock()
	defer re.cache.mu.RUnlock()

	return map[string]interface{}{
		"rule_count":  len(re.cache.rules),
		"last_update": re.cache.lastUpdate,
		"ttl_seconds": re.cache.ttl.Seconds(),
		"is_valid":    time.Since(re.cache.lastUpdate) < re.cache.ttl,
	}
}

// TestRoute 测试路由规则（用于调试）
func (re *RoutingEngine) TestRoute(labels map[string]string) (*RoutingResult, error) {
	return re.RouteAlert(labels)
}
