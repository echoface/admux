package concurrent

import (
	"context"
	"sync"
	"time"
)

// Result 泛型结果类型
type Result[T any] struct {
	Value T
	Error error
}

// Task 泛型任务类型
type Task[T any] func(ctx context.Context) (T, error)

// ConcurrencyController 并发控制器
type ConcurrencyController struct {
	semaphore chan struct{}  // 信号量控制并发数
	wg        sync.WaitGroup // 等待组
}

// NewConcurrencyController 创建并发控制器
func NewConcurrencyController(maxConcurrency int) *ConcurrencyController {
	return &ConcurrencyController{
		semaphore: make(chan struct{}, maxConcurrency),
	}
}

// ExecuteWithTimeout 执行带超时的并发任务
func ExecuteWithTimeout[T any](
	c *ConcurrencyController,
	ctx context.Context,
	tasks []Task[T],
	timeout time.Duration,
) ([]Result[T], error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 创建结果通道
	resultChan := make(chan Result[T], len(tasks))

	// 启动所有任务
	for i, task := range tasks {
		c.wg.Add(1)
		go executeTask(c, timeoutCtx, i, task, resultChan)
	}

	// 等待所有任务完成或超时
	go func() {
		c.wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var results []Result[T]
	for result := range resultChan {
		results = append(results, result)
	}

	// 检查是否超时
	if timeoutCtx.Err() == context.DeadlineExceeded {
		return results, timeoutCtx.Err()
	}

	return results, nil
}

// executeTask 执行单个任务
func executeTask[T any](
	c *ConcurrencyController,
	ctx context.Context,
	taskID int,
	task Task[T],
	resultChan chan<- Result[T],
) {
	defer c.wg.Done()

	// 获取信号量许可
	c.semaphore <- struct{}{}
	defer func() { <-c.semaphore }()

	// 执行任务
	value, err := task(ctx)
	resultChan <- Result[T]{
		Value: value,
		Error: err,
	}
}

// BatchProcessor 批量处理器
type BatchProcessor[T any] struct {
	batchSize   int
	processFunc func([]T)
	buffer      []T
	mu          sync.Mutex
}

// NewBatchProcessor 创建批量处理器
func NewBatchProcessor[T any](batchSize int, processFunc func([]T)) *BatchProcessor[T] {
	return &BatchProcessor[T]{
		batchSize:   batchSize,
		processFunc: processFunc,
		buffer:      make([]T, 0, batchSize),
	}
}

// Add 添加项目到批量处理器
func (bp *BatchProcessor[T]) Add(item T) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.buffer = append(bp.buffer, item)
	if len(bp.buffer) >= bp.batchSize {
		bp.flush()
	}
}

// Flush 强制刷新缓冲区
func (bp *BatchProcessor[T]) Flush() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if len(bp.buffer) > 0 {
		bp.flush()
	}
}

// flush 内部刷新方法
func (bp *BatchProcessor[T]) flush() {
	// 创建副本以避免在回调中修改
	batch := make([]T, len(bp.buffer))
	copy(batch, bp.buffer)

	// 异步处理批次
	go bp.processFunc(batch)

	// 清空缓冲区
	bp.buffer = bp.buffer[:0]
}