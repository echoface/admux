package dspbidder

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/echoface/admux/internal/adx_engine/adxcore"
	"github.com/echoface/admux/internal/adx_engine/config"
)

// BidderIndexManager DSP索引管理器
type BidderIndexManager struct {
	config       *config.ServerConfig
	configLoader *ConfigLoader
	indexBuilder *IndexBuilder
	dynamicCache *DSPDynamicCache
	currentIndex *DSPIndex
	factory      *adxcore.BidderFactory

	// 运行时状态
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup

	// 监控
	lastScanTime time.Time
	scanCount    int64
	errorCount   int64
}

// NewBidderIndexManager 创建DSP索引管理器
func NewBidderIndexManager(cfg *config.ServerConfig) (*BidderIndexManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建DSP配置加载器
	s3Config := &S3Config{
		Endpoint:        cfg.S3.Endpoint,
		AccessKeyID:     cfg.S3.AccessKeyID,
		SecretAccessKey: cfg.S3.SecretAccessKey,
		BucketName:      cfg.S3.BucketName,
		Prefix:          cfg.S3.Prefix,
		UseSSL:          cfg.S3.UseSSL,
		ScanInterval:    cfg.S3.ScanInterval,
		Timeout:         10 * time.Second,
	}

	configLoader, err := NewConfigLoader(s3Config)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create ConfigLoader: %w", err)
	}

	// 创建索引构建器
	indexBuilder := NewIndexBuilder()

	// 创建动态缓存
	dynamicCache := NewDSPDynamicCache(10000)

	// 获取全局Bidder工厂
	factory := adxcore.GetGlobalBidderFactory()

	mgr := &BidderIndexManager{
		config:       cfg,
		configLoader: configLoader,
		indexBuilder: indexBuilder,
		dynamicCache: dynamicCache,
		factory:      factory,
		ctx:          ctx,
		cancel:       cancel,
	}

	return mgr, nil
}

// Start 启动索引管理器
func (m *BidderIndexManager) Start() error {
	log.Println("Starting BidderIndexManager...")

	// 初始加载DSP索引
	if err := m.initialLoad(); err != nil {
		return fmt.Errorf("failed to initial load DSPs: %w", err)
	}

	// 启动定时扫描任务
	m.wg.Add(1)
	go m.scanLoop()

	// 启动QPS重置任务
	m.wg.Add(1)
	go m.qpsResetLoop()

	log.Println("BidderIndexManager started successfully")
	return nil
}

// Stop 停止索引管理器
func (m *BidderIndexManager) Stop() error {
	log.Println("Stopping BidderIndexManager...")

	m.cancel()

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("BidderIndexManager stopped successfully")
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for BidderIndexManager to stop")
	}
}

// initialLoad 初始加载DSP索引
func (m *BidderIndexManager) initialLoad() error {
	log.Println("Loading DSPs from S3...")

	dspMap, err := m.configLoader.ReadAllDSPs()
	if err != nil {
		return fmt.Errorf("failed to read DSPs: %w", err)
	}

	if len(dspMap) == 0 {
		log.Println("No DSPs found in S3")
		return nil
	}

	log.Printf("Found %d DSPs from S3", len(dspMap))

	// 构建索引
	index, err := m.indexBuilder.BuildDSPIndex(dspMap)
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	// 注册Bidder
	if err := m.registerBidders(dspMap); err != nil {
		return fmt.Errorf("failed to register bidders: %w", err)
	}

	// 更新当前索引
	m.mu.Lock()
	m.currentIndex = index
	m.mu.Unlock()

	// 切换索引
	if err := m.indexBuilder.SwitchIndex(index); err != nil {
		log.Printf("Failed to switch index: %v", err)
	}

	m.lastScanTime = time.Now()
	m.scanCount++

	log.Println("Initial DSP load completed")
	return nil
}

// registerBidders 注册DSP Bidder到工厂
func (m *BidderIndexManager) registerBidders(dspMap map[string]*DSPInfo) error {
	for dspID, dspInfo := range dspMap {
		if dspInfo.Status != "active" {
			continue
		}

		// 创建Bidder实例
		bidder, err := m.createBidder(dspInfo)
		if err != nil {
			log.Printf("Failed to create bidder for %s: %v", dspID, err)
			continue
		}

		// 注册到工厂
		if err := m.factory.RegisterBidder(bidder); err != nil {
			log.Printf("Failed to register bidder %s: %v", dspID, err)
		} else {
			log.Printf("Registered bidder: %s (%s)", dspID, dspInfo.DSPName)
		}
	}

	return nil
}

// createBidder 创建Bidder实例
func (m *BidderIndexManager) createBidder(dspInfo *DSPInfo) (adxcore.Bidder, error) {
	// 创建BaseBidder（使用现有实现）
	bidder := NewBaseBidder(
		dspInfo.DSPID,
		dspInfo.Endpoint,
		dspInfo.QPSLimit,
		dspInfo.Timeout,
	)

	return bidder, nil
}

// scanLoop 定时扫描循环
func (m *BidderIndexManager) scanLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.S3.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.scanAndUpdate()
		case <-m.ctx.Done():
			return
		}
	}
}

// scanAndUpdate 扫描并更新DSP索引
func (m *BidderIndexManager) scanAndUpdate() {
	log.Println("Scanning DSPs from S3...")

	dspMap, err := m.configLoader.ReadAllDSPs()
	if err != nil {
		log.Printf("Failed to scan DSPs: %v", err)
		m.errorCount++
		return
	}

	m.mu.RLock()
	oldIndex := m.currentIndex
	m.mu.RUnlock()

	// 检查是否有变化
	if m.hasChanges(oldIndex, dspMap) {
		log.Println("DSP changes detected, updating index...")

		// 构建新索引
		newIndex, err := m.indexBuilder.BuildDSPIndex(dspMap)
		if err != nil {
			log.Printf("Failed to build new index: %v", err)
			m.errorCount++
			return
		}

		// 切换索引
		m.mu.Lock()
		m.currentIndex = newIndex
		m.mu.Unlock()

		// 使用be_indexer切换索引
		if err := m.indexBuilder.SwitchIndex(newIndex); err != nil {
			log.Printf("Failed to switch index: %v", err)
		}

		// 更新Bidder注册
		m.updateBidderRegistrations(dspMap)

		log.Printf("DSP index updated successfully, total DSPs: %d", newIndex.Size())
	}

	m.lastScanTime = time.Now()
	m.scanCount++
}

// hasChanges 检查索引是否有变化
func (m *BidderIndexManager) hasChanges(oldIndex *DSPIndex, newDSPMap map[string]*DSPInfo) bool {
	if oldIndex == nil {
		return true
	}

	if oldIndex.Size() != len(newDSPMap) {
		return true
	}

	// 详细检查每个DSP的变化
	for dspID, newDSP := range newDSPMap {
		oldDSP, exists := oldIndex.GetDSP(dspID)
		if !exists {
			return true // 新增DSP
		}

		// 检查关键字段是否变化
		if oldDSP.Status != newDSP.Status ||
			oldDSP.QPSLimit != newDSP.QPSLimit ||
			oldDSP.BudgetDaily != newDSP.BudgetDaily {
			return true
		}
	}

	return false
}

// updateBidderRegistrations 更新Bidder注册
func (m *BidderIndexManager) updateBidderRegistrations(dspMap map[string]*DSPInfo) {
	// 记录已注册的DSP
	registeredDSPs := make(map[string]bool)

	// 添加/更新DSP
	for dspID, dspInfo := range dspMap {
		if dspInfo.Status != "active" {
			// 如果状态变为非活跃，取消注册
			if m.factory.HasBidder(dspID) {
				if err := m.factory.UnregisterBidder(dspID); err != nil {
					log.Printf("Failed to unregister bidder %s: %v", dspID, err)
				} else {
					log.Printf("Unregistered bidder: %s", dspID)
				}
			}
			continue
		}

		// 创建或更新Bidder
		bidder, err := m.createBidder(dspInfo)
		if err != nil {
			log.Printf("Failed to create bidder for %s: %v", dspID, err)
			continue
		}

		// 注册Bidder
		if err := m.factory.RegisterBidder(bidder); err != nil {
			// 如果注册失败，可能已存在，尝试更新
			log.Printf("Failed to register bidder %s: %v", dspID, err)
		} else {
			log.Printf("Registered bidder: %s (%s)", dspID, dspInfo.DSPName)
			registeredDSPs[dspID] = true
		}
	}
}

// qpsResetLoop QPS重置循环
func (m *BidderIndexManager) qpsResetLoop() {
	defer m.wg.Done()

	// 每分钟重置QPS计数
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Resetting QPS counters...")

			// 重置所有活跃DSP的QPS计数
			dspMap := m.GetAllActiveDSPs()
			for dspID := range dspMap {
				m.dynamicCache.ResetDSPQPS(dspID)
			}

			log.Printf("QPS counters reset for %d DSPs", len(dspMap))

		case <-m.ctx.Done():
			return
		}
	}
}

// GetDSP 获取DSP信息
func (m *BidderIndexManager) GetDSP(dspID string) (*DSPInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentIndex == nil {
		return nil, false
	}

	return m.currentIndex.GetDSP(dspID)
}

// GetAllDSPs 获取所有DSP信息
func (m *BidderIndexManager) GetAllDSPs() map[string]*DSPInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentIndex == nil {
		return nil
	}

	return m.currentIndex.GetAllDSPs()
}

// GetAllActiveDSPs 获取所有活跃DSP信息
func (m *BidderIndexManager) GetAllActiveDSPs() map[string]*DSPInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentIndex == nil {
		return nil
	}

	activeDSPs := make(map[string]*DSPInfo)
	for dspID, dspInfo := range m.currentIndex.GetAllDSPs() {
		if dspInfo.Status == "active" {
			activeDSPs[dspID] = dspInfo
		}
	}

	return activeDSPs
}

// MatchDSPs 匹配DSP（使用be_indexer进行查询）
func (m *BidderIndexManager) MatchDSPs(conditions map[string][]string) []*DSPInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentIndex == nil {
		return nil
	}

	// 优先使用 be_indexer 进行查询
	if m.indexBuilder != nil {
		results, err := m.indexBuilder.SearchDSPs(conditions)
		if err == nil && len(results) > 0 {
			return results
		}
		// 如果 be_indexer 查询失败，回退到内部索引
	}

	// 回退到内部DSPIndex查询
	return m.currentIndex.MatchDSPs(conditions)
}

// GetDSPQPS 获取DSP当前QPS
func (m *BidderIndexManager) GetDSPQPS(dspID string) (int, bool) {
	return m.dynamicCache.GetDSPQPS(dspID)
}

// IncrementDSPQPS 增加DSP QPS计数
func (m *BidderIndexManager) IncrementDSPQPS(dspID string) int {
	return m.dynamicCache.IncrementDSPQPS(dspID)
}

// GetDSPStatus 获取DSP状态
func (m *BidderIndexManager) GetDSPStatus(dspID string) (*DSPStatus, bool) {
	return m.dynamicCache.GetDSPStatus(dspID)
}

// GetDSPBudget 获取DSP预算
func (m *BidderIndexManager) GetDSPBudget(dspID string) (*DSPBudget, bool) {
	return m.dynamicCache.GetDSPBudget(dspID)
}

// GetMetrics 获取指标
func (m *BidderIndexManager) GetMetrics() *BidderIndexManagerMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &BidderIndexManagerMetrics{
		LastScanTime:   m.lastScanTime,
		ScanCount:      m.scanCount,
		ErrorCount:     m.errorCount,
		CurrentDSPs:    m.currentIndex.Size(),
		ActiveDSPs:     len(m.GetAllActiveDSPs()),
		RegisteredDSPs: m.factory.BidderCount(),
		CacheMetrics:   m.dynamicCache.GetMetrics(),
	}
}

// BidderIndexManagerMetrics 索引管理器指标
type BidderIndexManagerMetrics struct {
	LastScanTime   time.Time            `json:"last_scan_time"`
	ScanCount      int64                `json:"scan_count"`
	ErrorCount     int64                `json:"error_count"`
	CurrentDSPs    int                  `json:"current_dsp_count"`
	ActiveDSPs     int                  `json:"active_dsp_count"`
	RegisteredDSPs int                  `json:"registered_dsp_count"`
	CacheMetrics   *DynamicCacheMetrics `json:"cache_metrics"`
}

// HTTPClientImpl HTTP客户端实现
type HTTPClientImpl struct {
	Endpoint  string
	AuthToken string
	Timeout   time.Duration
}

// Do 执行HTTP请求
func (c *HTTPClientImpl) Do(req *Request) (*Response, error) {
	// 实际实现中，这里会执行真实的HTTP请求
	// 当前为模拟实现
	return &Response{
		StatusCode: 200,
		Body:       []byte(`{"id":"test","bid":100}`),
	}, nil
}

// Request HTTP请求结构
type Request struct {
	Method  string
	URL     string
	Header  map[string]string
	Body    []byte
	Timeout time.Duration
}

// Response HTTP响应结构
type Response struct {
	StatusCode int
	Header     map[string]string
	Body       []byte
}
