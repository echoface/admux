package dspbidder

import (
	"sync"
	"time"
)

// LRUCache LRU缓存实现
type LRUCache struct {
	capacity int
	items    map[string]*cacheItem
	head     *cacheItem
	tail     *cacheItem
	mu       sync.RWMutex
}

type cacheItem struct {
	key       string
	value     interface{}
	timestamp time.Time
	prev      *cacheItem
	next      *cacheItem
}

// NewLRUCache 创建LRU缓存
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 1000
	}

	cache := &LRUCache{
		capacity: capacity,
		items:    make(map[string]*cacheItem),
	}

	// 初始化双向链表
	cache.head = &cacheItem{}
	cache.tail = &cacheItem{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head

	return cache
}

// Get 获取缓存值
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// 移动到头部
	c.moveToHead(item)
	return item.value, true
}

// Set 设置缓存值
func (c *LRUCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[key]; exists {
		// 更新现有项
		item.value = value
		item.timestamp = time.Now()
		c.moveToHead(item)
		return
	}

	// 创建新项
	item := &cacheItem{
		key:       key,
		value:     value,
		timestamp: time.Now(),
	}

	// 添加到头部
	c.items[key] = item
	c.addToHead(item)

	// 检查容量
	if len(c.items) > c.capacity {
		// 移除尾部项
		c.removeTail()
	}
}

// Remove 移除缓存项
func (c *LRUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[key]; exists {
		c.removeItem(item)
		delete(c.items, key)
	}
}

// Clear 清空缓存
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	c.head.next = c.tail
	c.tail.prev = c.head
}

// GetKeys 获取所有键
func (c *LRUCache) GetKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// GetSize 获取当前大小
func (c *LRUCache) GetSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// addToHead 添加项到头部
func (c *LRUCache) addToHead(item *cacheItem) {
	item.next = c.head.next
	item.prev = c.head
	c.head.next.prev = item
	c.head.next = item
}

// moveToHead 移动项到头部
func (c *LRUCache) moveToHead(item *cacheItem) {
	c.removeItem(item)
	c.addToHead(item)
}

// removeItem 移除项
func (c *LRUCache) removeItem(item *cacheItem) {
	item.prev.next = item.next
	item.next.prev = item.prev
}

// removeTail 移除尾部项
func (c *LRUCache) removeTail() {
	if c.tail.prev == c.head {
		return
	}

	lastItem := c.tail.prev
	c.removeItem(lastItem)
	delete(c.items, lastItem.key)
}

// DSPDynamicCache DSP动态信息缓存
type DSPDynamicCache struct {
	lru     *LRUCache
	mu      sync.RWMutex
	metrics *DynamicCacheMetrics
}

// DynamicCacheMetrics 缓存指标
type DynamicCacheMetrics struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Evictions  int64 `json:"evictions"`
	TotalItems int64 `json:"total_items"`
}

// NewDSPDynamicCache 创建DSP动态信息缓存
func NewDSPDynamicCache(capacity int) *DSPDynamicCache {
	return &DSPDynamicCache{
		lru:     NewLRUCache(capacity),
		metrics: &DynamicCacheMetrics{},
	}
}

// GetDSPStatus 获取DSP状态
func (c *DSPDynamicCache) GetDSPStatus(dspID string) (*DSPStatus, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := "status:" + dspID
	value, found := c.lru.Get(key)
	if !found {
		c.metrics.Misses++
		return nil, false
	}

	c.metrics.Hits++
	return value.(*DSPStatus), true
}

// SetDSPStatus 设置DSP状态
func (c *DSPDynamicCache) SetDSPStatus(dspID string, status *DSPStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := "status:" + dspID
	c.lru.Set(key, status)
}

// GetDSPBudget 获取DSP预算信息
func (c *DSPDynamicCache) GetDSPBudget(dspID string) (*DSPBudget, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := "budget:" + dspID
	value, found := c.lru.Get(key)
	if !found {
		c.metrics.Misses++
		return nil, false
	}

	c.metrics.Hits++
	return value.(*DSPBudget), true
}

// SetDSPBudget 设置DSP预算信息
func (c *DSPDynamicCache) SetDSPBudget(dspID string, budget *DSPBudget) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := "budget:" + dspID
	c.lru.Set(key, budget)
}

// GetDSPQPS 获取DSP QPS信息
func (c *DSPDynamicCache) GetDSPQPS(dspID string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := "qps:" + dspID
	value, found := c.lru.Get(key)
	if !found {
		c.metrics.Misses++
		return 0, false
	}

	c.metrics.Hits++
	return value.(int), true
}

// SetDSPQPS 设置DSP QPS信息
func (c *DSPDynamicCache) SetDSPQPS(dspID string, qps int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := "qps:" + dspID
	c.lru.Set(key, qps)
}

// IncrementDSPQPS 增加DSP QPS计数
func (c *DSPDynamicCache) IncrementDSPQPS(dspID string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := "qps:" + dspID
	currentQPS, found := c.lru.Get(key)
	if !found {
		c.lru.Set(key, 1)
		return 1
	}

	newQPS := currentQPS.(int) + 1
	c.lru.Set(key, newQPS)
	return newQPS
}

// DecrementDSPQPS 减少DSP QPS计数
func (c *DSPDynamicCache) DecrementDSPQPS(dspID string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := "qps:" + dspID
	currentQPS, found := c.lru.Get(key)
	if !found {
		return 0
	}

	newQPS := currentQPS.(int) - 1
	if newQPS < 0 {
		newQPS = 0
	}
	c.lru.Set(key, newQPS)
	return newQPS
}

// ResetDSPQPS 重置DSP QPS计数
func (c *DSPDynamicCache) ResetDSPQPS(dspID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := "qps:" + dspID
	c.lru.Set(key, 0)
}

// GetMetrics 获取缓存指标
func (c *DSPDynamicCache) GetMetrics() *DynamicCacheMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &DynamicCacheMetrics{
		Hits:       c.metrics.Hits,
		Misses:     c.metrics.Misses,
		Evictions:  c.metrics.Evictions,
		TotalItems: int64(c.lru.GetSize()),
	}
}

// DSPStatus DSP状态信息
type DSPStatus struct {
	DSPID     string    `json:"dsp_id"`
	Status    string    `json:"status"` // active, inactive, blocked
	Reason    string    `json:"reason,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DSPBudget DSP预算信息
type DSPBudget struct {
	DSPID       string    `json:"dsp_id"`
	DailyBudget float64   `json:"daily_budget"`
	SpentBudget float64   `json:"spent_budget"`
	Remaining   float64   `json:"remaining"`
	UpdatedAt   time.Time `json:"updated_at"`
	ResetTime   time.Time `json:"reset_time"`
}
