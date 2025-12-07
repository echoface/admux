package dspbidder

import (
	"testing"
	"time"

	"github.com/echoface/be_indexer"
	"github.com/stretchr/testify/assert"
)

func TestIndexBuilder_BEIndexerIntegration(t *testing.T) {
	// 创建索引构建器
	builder := NewIndexBuilder()
	assert.NotNil(t, builder)

	// 创建测试DSP数据
	dspMap := map[string]*DSPInfo{
		"dsp_001": {
			DSPID:       "dsp_001",
			DSPName:     "Test DSP 1",
			Status:      "active",
			QPSLimit:    1000,
			BudgetDaily: 50000.00,
			Endpoint:    "http://dsp1.example.com/bid",
			Timeout:     80 * time.Millisecond,
			RetryCount:  2,
			RetryDelay:  10 * time.Millisecond,
			Targeting: &DSPTargeting{
				IndexingDoc: []IndexingClause{
					{
						ClauseID:    "clause_1",
						Description: "iOS用户",
						Conditions: []Condition{
							{
								Field:    "USER_OS",
								Operator: "EQ",
								Values:   []string{"ios"},
							},
						},
					},
				},
			},
		},
		"dsp_002": {
			DSPID:       "dsp_002",
			DSPName:     "Test DSP 2",
			Status:      "active",
			QPSLimit:    2000,
			BudgetDaily: 100000.00,
			Endpoint:    "http://dsp2.example.com/bid",
			Timeout:     100 * time.Millisecond,
			RetryCount:  3,
			RetryDelay:  15 * time.Millisecond,
			Targeting: &DSPTargeting{
				IndexingDoc: []IndexingClause{
					{
						ClauseID:    "clause_2",
						Description: "Android用户",
						Conditions: []Condition{
							{
								Field:    "USER_OS",
								Operator: "EQ",
								Values:   []string{"android"},
							},
						},
					},
				},
			},
		},
	}

	// 构建索引
	err := builder.BuildDSPIndex(dspMap)
	assert.NoError(t, err)
	assert.Equal(t, 2, builder.GetDSPCount())

	// 验证 be_indexer 确实被初始化
	assert.True(t, builder.compiled)
	assert.NotNil(t, builder.indexer)

	// 验证 DSPMap 被保存
	assert.Equal(t, 2, len(builder.dspMap))

	// 测试搜索iOS用户
	conditions := map[string][]string{
		"USER_OS": []string{"ios"},
	}
	results, err := builder.SearchDSPs(conditions)
	assert.NoError(t, err)
	// 应该找到 dsp_001
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "dsp_001", results[0].DSPID)

	// 测试搜索Android用户
	conditions = map[string][]string{
		"USER_OS": []string{"android"},
	}
	results, err = builder.SearchDSPs(conditions)
	assert.NoError(t, err)
	// 应该找到 dsp_002
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "dsp_002", results[0].DSPID)

	// 测试搜索不存在的条件
	conditions = map[string][]string{
		"USER_OS": []string{"windows"},
	}
	results, err = builder.SearchDSPs(conditions)
	assert.NoError(t, err)
	// 应该没有匹配的结果
	assert.Equal(t, 0, len(results))

	// 测试空查询（应该返回所有DSP或有通配符匹配的DSP）
	conditions = map[string][]string{}
	results, err = builder.SearchDSPs(conditions)
	assert.NoError(t, err)
	// 可能返回0、1或2个结果，取决于通配符匹配
	t.Logf("Empty query returned %d results", len(results))

	t.Log("be_indexer integration test passed!")
}

func TestIndexBuilder_DocIDMapping(t *testing.T) {
	builder := NewIndexBuilder()
	assert.NotNil(t, builder)

	// 添加一个测试DSP
	dspID := "test_dsp_123"
	dspInfo := &DSPInfo{
		DSPID:     dspID,
		DSPName:   "Test DSP",
		Status:    "active",
		QPSLimit:  1000,
		Targeting: &DSPTargeting{
			IndexingDoc: []IndexingClause{
				{
					ClauseID:    "clause_1",
					Description: "Test clause",
					Conditions: []Condition{
						{
							Field:    "USER_OS",
							Operator: "EQ",
							Values:   []string{"ios"},
						},
					},
				},
			},
		},
	}

	// 手动调用 addDSPDocument
	err := builder.addDSPDocument(dspID, dspInfo)
	assert.NoError(t, err)

	// 验证 docID 到 dspID 的映射被建立
	assert.Equal(t, 1, len(builder.docIDMap))

	// 测试反向查找
	docID := be_indexer.DocID(hashStringToDocID(dspID))
	foundDSPID := builder.docIDMap[docID]
	assert.Equal(t, dspID, foundDSPID)

	t.Logf("DocID mapping test passed! DSP: %s -> DocID: %d", dspID, docID)
}
