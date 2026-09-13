package algorithm

import (
	"sync"
	"sync/atomic"
	"time"
)

type QueueOption func(opts *QueueOptions)

type QueueType interface {
	~string | ~int | ~uint32 | ~int64
}

type queueItem[T QueueType, K any] struct {
	pre  *queueItem[T, K]
	next *queueItem[T, K]
	key  T
	data K
}

type Queue[T QueueType, K any] struct {
	m        map[T]int
	first    *queueItem[T, K]
	end      *queueItem[T, K]
	size     int64
	duration time.Duration
	tick     *time.Ticker
	ticking  bool
	lock     sync.RWMutex
	gCounts  int32
}

type QueueOptions struct {
	duration time.Duration
	tick     *time.Ticker
}

// 定时移除队头
func WithRemoveTicker(duration time.Duration) QueueOption {
	return func(opts *QueueOptions) {
		opts.duration = duration
		opts.tick = time.NewTicker(duration)
	}
}

// Instance 初始化队列。修复：原实现把 tick/duration 写到局部新对象 qq 之外
// 的 receiver 上而返回 qq，WithRemoveTicker 完全失效；现直接在 receiver 上
// 初始化并返回自身，零值 Queue 亦可直接 Instance。
func (q *Queue[T, K]) Instance(opts ...QueueOption) *Queue[T, K] {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.m == nil {
		q.m = make(map[T]int)
	}
	qOpts := new(QueueOptions)
	for _, opt := range opts {
		opt(qOpts)
	}

	if qOpts.tick != nil {
		q.duration = qOpts.duration
		q.tick = qOpts.tick
	}
	return q
}

func (q *Queue[T, K]) Enqueue(key T, item K) {
	q.lock.Lock()
	q.m[key]++
	tmp := &queueItem[T, K]{
		data: item,
		key:  key,
	}
	if q.end == nil {
		q.end = tmp
		q.first = tmp
	} else {
		tmp.pre = q.end
		q.end.next = tmp
		q.end = q.end.next
	}
	q.size++

	// 定时清理 goroutine 仅启动一次（原实现经未初始化的 ctx.Deadline() 判断，
	// nil context 必 panic 且从未生效）
	if q.tick != nil && !q.ticking {
		q.ticking = true
		go q.dequeueByTick()
	}
	q.lock.Unlock()
}

func (q *Queue[T, K]) dequeueByTick() {
	atomic.AddInt32(&q.gCounts, 1)
	defer atomic.AddInt32(&q.gCounts, -1)
	q.tick.Reset(q.duration)
	for range q.tick.C {
		q.Dequeue()
		if q.Size() == 0 {
			return
		}
	}
}

func (q *Queue[T, K]) Dequeue() K {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.size == 0 {
		var zero K
		return zero
	}
	defer func() {
		q.m[q.first.key]--
		q.first = q.first.next
		q.size--
	}()
	return q.first.data
}

func (q *Queue[T, K]) Exists(key T) bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	count, exist := q.m[key]
	return exist && count > 0
}

func (q *Queue[T, K]) Size() int64 {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.size
}
