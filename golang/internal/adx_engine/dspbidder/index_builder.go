package dspbidder

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/echoface/be_indexer"
)

// IndexBuilder 索引构建器，使用be_indexer构建DSP索引
type IndexBuilder struct {
	builder      *be_indexer.IndexerBuilder
	indexer      be_indexer.BEIndex
	dspMap       map[string]*DSPInfo
	docIDToDSPID map[be_indexer.DocID]string
	compiled     bool
	stats        map[string]interface{}
	totalDocs    int64
	mu           sync.RWMutex
}

// NewIndexBuilder 创建索引构建器
func NewIndexBuilder() *IndexBuilder {
	builder := be_indexer.NewIndexerBuilder()

	return &IndexBuilder{
		builder:      builder,
		stats:        make(map[string]interface{}),
		compiled:     false,
		dspMap:       make(map[string]*DSPInfo),
		docIDToDSPID: make(map[be_indexer.DocID]string),
	}
}

// BuildDSPIndex 构建DSP索引
func (b *IndexBuilder) BuildDSPIndex(dspMap map[string]*DSPInfo) (*DSPIndex, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 重置构建器
	b.builder.Reset()

	// 保存DSPMap以便后续查询
	b.dspMap = dspMap

	// 创建内部索引
	dspIndex := NewDSPIndex(dspMap)

	// 将DSP转换为be_indexer Document并构建索引
	for dspID, dspInfo := range dspMap {
		if err := b.addDSPDocument(dspID, dspInfo); err != nil {
			return nil, fmt.Errorf("failed to add DSP %s: %w", dspID, err)
		}
	}

	// 构建索引
	b.indexer = b.builder.BuildIndex()

	// 编译索引器以优化性能（通过调用BuildIndex已隐式编译）
	b.compiled = true
	b.totalDocs = int64(len(dspMap))
	b.stats["total_docs"] = b.totalDocs

	return dspIndex, nil
}

// addDSPDocument 将DSP信息转换为Document并添加到索引
func (b *IndexBuilder) addDSPDocument(dspID string, dspInfo *DSPInfo) error {
	// 为每个DSP创建唯一ID
	docID := be_indexer.DocID(hashStringToDocID(dspID))

	// 建立docID到dspID的映射
	b.docIDToDSPID[docID] = dspID

	// 将DSP信息转换为be_indexer Document格式的JSON
	docJSON, err := convertDSPToDocumentJSON(docID, dspInfo)
	if err != nil {
		return fmt.Errorf("convert DSP to document JSON failed: %w", err)
	}

	// 解析JSON为Document
	doc := &be_indexer.Document{}
	if err := json.Unmarshal(docJSON, doc); err != nil {
		return fmt.Errorf("unmarshal document failed: %w", err)
	}

	// 添加文档到构建器
	return b.builder.AddDocument(doc)
}

// SearchDSPs 使用be_indexer搜索DSP，返回匹配DSP的详细信息
func (b *IndexBuilder) SearchDSPs(query map[string][]string) ([]*DSPInfo, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.compiled || b.indexer == nil {
		return []*DSPInfo{}, nil
	}

	// 将查询条件转换为Assignments
	assignments := make(be_indexer.Assignments)

	for field, values := range query {
		assignments[be_indexer.BEField(field)] = values
	}

	// 执行查询
	docIDs, err := b.indexer.Retrieve(assignments)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// 将DocID转换回DSPID，并查找对应的DSPInfo
	var results []*DSPInfo
	dspMap := b.getDSPMap()
	for _, docID := range docIDs {
		dspID := b.docIDToDSPID(docID)
		if dspInfo, exists := dspMap[dspID]; exists {
			results = append(results, dspInfo)
		}
	}

	return results, nil
}

// getDSPMap 获取内部DSPMap（用于从DocID查找DSPInfo）
func (b *IndexBuilder) getDSPMap() map[string]*DSPInfo {
	return b.dspMap
}

// docIDToDSPID 将DocID转换回DSPID
func (b *IndexBuilder) docIDToDSPID(docID be_indexer.DocID) string {
	if dspID, exists := b.docIDToDSPID[docID]; exists {
		return dspID
	}
	// 如果没找到，返回默认值
	return fmt.Sprintf("dsp_%d", docID)
}

// SwitchIndex 切换索引（be_indexer暂不支持原子切换，使用重建替代）
func (b *IndexBuilder) SwitchIndex(newIndex *DSPIndex) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	fmt.Printf("Index switch requested for %d DSPs (rebuilding index)\n", newIndex.Size())
	// be_indexer当前不支持原子索引切换
	// 在真实场景中，可以通过版本号管理实现零停机切换

	return nil
}

// GetIndexStats 获取索引统计信息
func (b *IndexBuilder) GetIndexStats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range b.stats {
		stats[k] = v
	}

	stats["total_docs"] = b.totalDocs
	stats["compiled"] = b.compiled

	return stats
}

// RebuildIndex 重建索引
func (b *IndexBuilder) RebuildIndex(dspMap map[string]*DSPInfo) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 重置构建器
	b.builder.Reset()
	b.compiled = false

	// 重建索引
	_, err := b.BuildDSPIndex(dspMap)
	return err
}

// Close 关闭索引构建器
func (b *IndexBuilder) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// be_indexer当前无显式Close方法
	// 清理资源
	b.stats = nil
	b.builder = nil
	b.indexer = nil
	b.compiled = false

	return nil
}

// convertDSPToDocumentJSON 将DSP信息转换为be_indexer Document的JSON格式
func convertDSPToDocumentJSON(docID be_indexer.DocID, dspInfo *DSPInfo) ([]byte, error) {
	// 构建Document结构
	doc := map[string]interface{}{
		"id":   int64(docID),
		"cons": []map[string]interface{}{},
	}

	// 处理定向条件
	if dspInfo.Targeting != nil && len(dspInfo.Targeting.IndexingDoc) > 0 {
		// 为每个条款创建一个Conjunction
		cons := []map[string]interface{}{}

		for _, clause := range dspInfo.Targeting.IndexingDoc {
			// 构建exprs字段
			exprs := map[string]interface{}{}

			for _, condition := range clause.Conditions {
				// 根据操作符确定包含/排除
				var include bool
				var values interface{}

				switch condition.Operator {
				case "EQ", "IN":
					include = true
					values = condition.Values
				case "NOT_IN":
					include = false
					values = condition.Values
				default:
					// 默认当作IN处理
					include = true
					values = condition.Values
				}

				// 添加到exprs
				exprs[condition.Field] = []map[string]interface{}{
					{
						"inc":   include,
						"value": values,
					},
				}
			}

			// 只有非空exprs才添加
			if len(exprs) > 0 {
				cons = append(cons, map[string]interface{}{
					"exprs": exprs,
				})
			}
		}

		// 如果有Conjunction，添加到文档
		if len(cons) > 0 {
			doc["cons"] = cons
		}
	}

	// 如果没有定向条件，创建一个通配符Conjunction匹配所有请求
	if doc["cons"] == nil {
		doc["cons"] = []map[string]interface{}{
			{
				"exprs": map[string]interface{}{
					"_Z_": []map[string]interface{}{
						{
							"inc":   true,
							"value": []string{},
						},
					},
				},
			},
		}
	}

	// 序列化为JSON
	return json.Marshal(doc)
}

// hashStringToDocID 将字符串哈希为DocID
func hashStringToDocID(s string) be_indexer.DocID {
	// 简单的哈希实现，实际应使用更好的哈希算法
	hash := int64(0)
	for i, c := range s {
		hash = hash*31 + int64(c)
		if i > 10 { // 限制长度
			break
		}
	}
	if hash < 0 {
		hash = -hash
	}
	return be_indexer.DocID(hash % 2147483647) // 限制在int32范围内
}

// DSPIndexerDoc DSP索引文档结构（保留以兼容现有代码）
type DSPIndexerDoc struct {
	DSPID           string           `json:"dsp_id"`
	DSPName         string           `json:"dsp_name"`
	Status          string           `json:"status"`
	QPSLimit        int              `json:"qps_limit"`
	BudgetDaily     float64          `json:"budget_daily"`
	Endpoint        string           `json:"endpoint"`
	AuthToken       string           `json:"auth_token"`
	Timeout         time.Duration    `json:"timeout"`
	RetryCount      int              `json:"retry_count"`
	RetryDelay      time.Duration    `json:"retry_delay"`
	UpdatedAt       time.Time        `json:"updated_at"`
	TargetingFields []TargetingField `json:"targeting_fields,omitempty"`
}

// TargetingField 定向字段（保留以兼容现有代码）
type TargetingField struct {
	Field       string   `json:"field"`
	Operator    string   `json:"operator"`
	Values      []string `json:"values"`
	ClauseID    string   `json:"clause_id,omitempty"`
	Description string   `json:"description,omitempty"`
}
