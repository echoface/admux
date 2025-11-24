package dspbidder

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBidderIndexManager(t *testing.T) {
	// 测试创建（注意：这会失败，因为我们没有真实的S3）
	// cfg := &config.ServerConfig{
	//	S3: config.S3Config{
	//		Endpoint:     "localhost:9000",
	//		BucketName:   "test-bucket",
	//		Prefix:       "test/",
	//		UseSSL:       false,
	//		ScanInterval: time.Minute,
	//	},
	// }
	// mgr, err := NewBidderIndexManager(cfg)
	// assert.NoError(t, err)
	// assert.NotNil(t, mgr)

	t.Log("Test case prepared - requires real S3 endpoint to fully test")
}

func TestDSPInfoMarshaling(t *testing.T) {
	dspJSON := `{
		"dsp_id": "dsp_test_001",
		"dsp_name": "Test DSP",
		"status": "active",
		"qps_limit": 1000,
		"budget_daily": 50000.00,
		"endpoint": "http://test-dsp.com/bid",
		"timeout": 80000000000,
		"retry_count": 2,
		"retry_delay": 10000000
	}`

	var dspInfo DSPInfo
	err := json.Unmarshal([]byte(dspJSON), &dspInfo)
	assert.NoError(t, err)
	assert.Equal(t, "dsp_test_001", dspInfo.DSPID)
	assert.Equal(t, "Test DSP", dspInfo.DSPName)
	assert.Equal(t, "active", dspInfo.Status)
	assert.Equal(t, 1000, dspInfo.QPSLimit)
	assert.Equal(t, 50000.00, dspInfo.BudgetDaily)
	assert.Equal(t, int64(80000000000), dspInfo.Timeout.Nanoseconds())
}

func TestDSPTargetingMarshaling(t *testing.T) {
	targetingJSON := `{
		"targeting": {
			"indexingdoc": [
				{
					"clause_id": "clause_1",
					"description": "北京或上海的iOS用户",
					"conditions": [
						{
							"field": "USER_GEO",
							"operator": "IN",
							"values": ["BJ", "SH"]
						},
						{
							"field": "USER_OS",
							"operator": "EQ",
							"values": ["ios"]
						}
					]
				}
			]
		}
	}`

	var dspInfo DSPInfo
	err := json.Unmarshal([]byte(targetingJSON), &dspInfo)
	assert.NoError(t, err)
	assert.NotNil(t, dspInfo.Targeting)
	assert.Equal(t, 1, len(dspInfo.Targeting.IndexingDoc))
	assert.Equal(t, "clause_1", dspInfo.Targeting.IndexingDoc[0].ClauseID)
	assert.Equal(t, 2, len(dspInfo.Targeting.IndexingDoc[0].Conditions))
}

func TestNewDSPIndex(t *testing.T) {
	dspMap := map[string]*DSPInfo{
		"dsp_1": {
			DSPID:     "dsp_1",
			DSPName:   "DSP 1",
			Status:    "active",
			QPSLimit:  1000,
			Targeting: &DSPTargeting{},
		},
		"dsp_2": {
			DSPID:     "dsp_2",
			DSPName:   "DSP 2",
			Status:    "active",
			QPSLimit:  2000,
			Targeting: &DSPTargeting{},
		},
	}

	index := NewDSPIndex(dspMap)
	assert.NotNil(t, index)
	assert.Equal(t, 2, index.Size())

	// 测试获取DSP
	dsp, exists := index.GetDSP("dsp_1")
	assert.True(t, exists)
	assert.Equal(t, "DSP 1", dsp.DSPName)

	// 测试获取所有DSP
	allDSPs := index.GetAllDSPs()
	assert.Equal(t, 2, len(allDSPs))
}

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(3)

	// 测试设置和获取
	cache.Set("key1", "value1")
	value, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// 测试更新
	cache.Set("key1", "value2")
	value, _ = cache.Get("key1")
	assert.Equal(t, "value2", value)

	// 测试容量限制
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	cache.Set("key4", "value4") // 这应该淘汰key1

	_, exists = cache.Get("key1")
	assert.False(t, exists)

	_, exists = cache.Get("key4")
	assert.True(t, exists)
}

func TestDSPDynamicCache(t *testing.T) {
	cache := NewDSPDynamicCache(100)

	// 测试设置和获取DSP状态
	status := &DSPStatus{
		DSPID:     "dsp_1",
		Status:    "active",
		UpdatedAt: time.Now(),
	}

	cache.SetDSPStatus("dsp_1", status)
	retrievedStatus, exists := cache.GetDSPStatus("dsp_1")
	assert.True(t, exists)
	assert.Equal(t, "active", retrievedStatus.Status)

	// 测试QPS计数
	qps := cache.IncrementDSPQPS("dsp_1")
	assert.Equal(t, 1, qps)

	qps = cache.IncrementDSPQPS("dsp_1")
	assert.Equal(t, 2, qps)

	qps = cache.DecrementDSPQPS("dsp_1")
	assert.Equal(t, 1, qps)

	cache.ResetDSPQPS("dsp_1")
	qps, _ = cache.GetDSPQPS("dsp_1")
	assert.Equal(t, 0, qps)
}

func TestDSPIndexMatchDSPs(t *testing.T) {
	dspMap := map[string]*DSPInfo{
		"dsp_1": {
			DSPID:  "dsp_1",
			Status: "active",
			Targeting: &DSPTargeting{
				IndexingDoc: []IndexingClause{
					{
						ClauseID: "ios_users",
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
		"dsp_2": {
			DSPID:  "dsp_2",
			Status: "active",
			Targeting: &DSPTargeting{
				IndexingDoc: []IndexingClause{
					{
						ClauseID: "android_users",
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

	index := NewDSPIndex(dspMap)

	// 测试匹配iOS用户
	conditions := map[string][]string{
		"USER_OS": []string{"ios"},
	}
	matchedDSPs := index.MatchDSPs(conditions)
	assert.Equal(t, 1, len(matchedDSPs))
	assert.Equal(t, "dsp_1", matchedDSPs[0].DSPID)

	// 测试匹配Android用户
	conditions = map[string][]string{
		"USER_OS": []string{"android"},
	}
	matchedDSPs = index.MatchDSPs(conditions)
	assert.Equal(t, 1, len(matchedDSPs))
	assert.Equal(t, "dsp_2", matchedDSPs[0].DSPID)

	// 测试匹配不存在的条件
	conditions = map[string][]string{
		"USER_OS": []string{"windows"},
	}
	matchedDSPs = index.MatchDSPs(conditions)
	assert.Equal(t, 0, len(matchedDSPs))

	// 测试空条件匹配所有活跃DSP
	conditions = map[string][]string{}
	matchedDSPs = index.MatchDSPs(conditions)
	assert.Equal(t, 2, len(matchedDSPs))
}
