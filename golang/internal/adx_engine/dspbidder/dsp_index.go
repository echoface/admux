package dspbidder

import "time"

// DSPIndex DSP索引结构，封装所有DSP信息用于快速检索
type DSPIndex struct {
	dspMap    map[string]*DSPInfo
	clauseMap map[string][]*DSPInfo
	updatedAt time.Time
}

// NewDSPIndex 创建新的DSP索引
func NewDSPIndex(dspMap map[string]*DSPInfo) *DSPIndex {
	index := &DSPIndex{
		dspMap:    make(map[string]*DSPInfo),
		clauseMap: make(map[string][]*DSPInfo),
	}

	for dspID, dspInfo := range dspMap {
		index.dspMap[dspID] = dspInfo

		// 构建条款索引
		if dspInfo.Targeting != nil {
			for _, clause := range dspInfo.Targeting.IndexingDoc {
				clauseDSPList := index.clauseMap[clause.ClauseID]
				clauseDSPList = append(clauseDSPList, dspInfo)
				index.clauseMap[clause.ClauseID] = clauseDSPList
			}
		}
	}

	index.updatedAt = time.Now()
	return index
}

// GetDSP 获取DSP信息
func (i *DSPIndex) GetDSP(dspID string) (*DSPInfo, bool) {
	dsp, exists := i.dspMap[dspID]
	return dsp, exists
}

// GetAllDSPs 获取所有DSP信息
func (i *DSPIndex) GetAllDSPs() map[string]*DSPInfo {
	result := make(map[string]*DSPInfo)
	for dspID, dspInfo := range i.dspMap {
		result[dspID] = dspInfo
	}
	return result
}

// GetDSPsByClause 根据条款ID获取DSP列表
func (i *DSPIndex) GetDSPsByClause(clauseID string) []*DSPInfo {
	return i.clauseMap[clauseID]
}

// GetActiveDSPs 获取活跃DSP列表
func (i *DSPIndex) GetActiveDSPs() []*DSPInfo {
	var activeDSPs []*DSPInfo
	for _, dspInfo := range i.dspMap {
		if dspInfo.Status == "active" {
			activeDSPs = append(activeDSPs, dspInfo)
		}
	}
	return activeDSPs
}

// GetDSPsByBudget 根据预算条件筛选DSP
func (i *DSPIndex) GetDSPsByBudget(minBudget float64, maxBudget float64) []*DSPInfo {
	var filteredDSPs []*DSPInfo
	for _, dspInfo := range i.dspMap {
		if dspInfo.Status != "active" {
			continue
		}
		if dspInfo.BudgetDaily >= minBudget && (maxBudget == 0 || dspInfo.BudgetDaily <= maxBudget) {
			filteredDSPs = append(filteredDSPs, dspInfo)
		}
	}
	return filteredDSPs
}

// GetUpdatedAt 获取最后更新时间
func (i *DSPIndex) GetUpdatedAt() time.Time {
	return i.updatedAt
}

// Size 返回DSP数量
func (i *DSPIndex) Size() int {
	return len(i.dspMap)
}

// MatchDSPs 匹配符合条件的DSP
func (i *DSPIndex) MatchDSPs(conditions map[string][]string) []*DSPInfo {
	if len(conditions) == 0 {
		return i.GetActiveDSPs()
	}

	var matchedDSPs []*DSPInfo
DSP_LOOP:
	for _, dspInfo := range i.dspMap {
		if dspInfo.Status != "active" {
			continue
		}

		if dspInfo.Targeting == nil {
			// 如果没有定向配置，默认匹配
			matchedDSPs = append(matchedDSPs, dspInfo)
			continue
		}

		// 检查是否匹配任一定向条款
		for _, clause := range dspInfo.Targeting.IndexingDoc {
			if i.matchClause(conditions, clause) {
				matchedDSPs = append(matchedDSPs, dspInfo)
				continue DSP_LOOP
			}
		}
	}

	return matchedDSPs
}

// matchClause 检查条件是否匹配条款
func (i *DSPIndex) matchClause(conditions map[string][]string, clause IndexingClause) bool {
	for _, condition := range clause.Conditions {
		// 从请求条件中获取值
		reqValues, exists := conditions[condition.Field]
		if !exists {
			return false
		}

		// 检查是否匹配操作符
		if !i.matchCondition(reqValues, condition) {
			return false
		}
	}
	return true
}

// matchCondition 检查单个条件是否匹配
func (i *DSPIndex) matchCondition(reqValues []string, condition Condition) bool {
	switch condition.Operator {
	case "EQ":
		// 等于操作
		for _, reqValue := range reqValues {
			for _, targetValue := range condition.Values {
				if reqValue == targetValue {
					return true
				}
			}
		}
		return false
	case "IN":
		// 在集合中
		for _, reqValue := range reqValues {
			for _, targetValue := range condition.Values {
				if reqValue == targetValue {
					return true
				}
			}
		}
		return false
	case "NOT_IN":
		// 不在集合中
		for _, reqValue := range reqValues {
			for _, targetValue := range condition.Values {
				if reqValue == targetValue {
					return false
				}
			}
		}
		return true
	case "GT":
		// 大于（用于数值比较）
		if len(reqValues) > 0 && len(condition.Values) > 0 {
			return reqValues[0] > condition.Values[0]
		}
		return false
	case "LT":
		// 小于（用于数值比较）
		if len(reqValues) > 0 && len(condition.Values) > 0 {
			return reqValues[0] < condition.Values[0]
		}
		return false
	default:
		return false
	}
}
