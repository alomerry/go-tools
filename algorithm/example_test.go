package algorithm

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime/debug"
	"sync/atomic"
	"testing"
	"time"
)

type stats struct {
	count int
}

func (s stats) Less(i, j interface{}) bool {
	return i.(*stats).count < j.(*stats).count
}

func (s stats) Equal(i, j interface{}) bool {
	return i.(*stats).count == j.(*stats).count
}

func (s stats) Empty() bool {
	return s.count == 0
}

func newStats(count int) *stats {
	return &stats{count: count}
}

func genBst1() *BSTree[*stats] {
	var (
		bst *BSTree[*stats]
	)
	bst = bst.Insert(NewBSTNode(newStats(9)))
	bst.Insert(NewBSTNode(newStats(1)))
	bst.Insert(NewBSTNode(newStats(6)))
	bst.Insert(NewBSTNode(newStats(4)))
	bst.Insert(NewBSTNode(newStats(2)))
	bst.Insert(NewBSTNode(newStats(7)))
	return bst
}

func genBst2() *BSTree[*stats] {
	var (
		bst *BSTree[*stats]
	)
	bst = bst.Insert(NewBSTNode(newStats(5)))
	bst.Insert(NewBSTNode(newStats(3)))
	bst.Insert(NewBSTNode(newStats(8)))
	bst.Insert(NewBSTNode(newStats(20)))
	bst.Insert(NewBSTNode(newStats(10)))
	return bst
}

func TestBSTToArray(t *testing.T) {
	var (
		bst = genBst1()
	)
	valid := []int{1, 2, 4, 6, 7, 9}
	list := bst.ToArray()
	assert.Equal(t, len(list), len(valid))
	for i := range list {
		assert.True(t, list[i].count == valid[i])
	}
}

func TestBSTMerge1(t *testing.T) {
	var (
		bst1 = genBst1()
		bst2 = genBst2()
	)
	bst1 = bst1.Merge(bst2)
	list := bst1.ToArray()
	valid := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 20}
	assert.Equal(t, len(list), len(valid))
	for i := range list {
		assert.True(t, list[i].count == valid[i])
	}
}

func TestBSTMerge2(t *testing.T) {
	var (
		bst1 = genBst1()
		bst2 *BSTree[*stats]
	)
	bst1 = bst1.Merge(bst2)
	list := bst1.ToArray()
	valid := []int{1, 2, 4, 6, 7, 9}
	assert.Equal(t, len(list), len(valid))
	for i := range list {
		assert.True(t, list[i].count == valid[i])
	}
}

func TestBSTMerge3(t *testing.T) {
	var (
		bst1 = genBst1()
		bst2 *BSTree[*stats]
	)
	bst2 = bst2.Merge(bst1)
	list := bst2.ToArray()
	valid := []int{1, 2, 4, 6, 7, 9}
	assert.Equal(t, len(list), len(valid))
	for i := range list {
		assert.True(t, list[i].count == valid[i])
	}
}

type Refund struct {
	Id int
}

func TestQueueRWConcurrent(t *testing.T) {
	var (
		queue = (&Queue[int, Refund]{}).Instance()
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 2)
	ctx, f := context.WithCancel(context.Background())
	go func() {
		for {
			assert.True(t, atomic.LoadInt32(&queue.gCounts) <= 1)
			select {
			case <-ticker.C:
				queue.Dequeue()
			case <-ctx.Done():
				return
			}
		}
	}()
	for i := 0; i < 99999; i++ {
		order := Refund{i}
		queue.Enqueue(order.Id, order)
		assert.True(t, atomic.LoadInt32(&queue.gCounts) <= 1)
	}
	f()
	time.Sleep(time.Second)
}

func TestQueueRWConcurrentWithRemoveTicker(t *testing.T) {
	var (
		queue = (&Queue[int, Refund]{}).Instance(WithRemoveTicker(time.Nanosecond * 11))
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 2)
	ctx, f := context.WithCancel(context.Background())
	go func() {
		for {
			assert.True(t, atomic.LoadInt32(&queue.gCounts) <= 1)
			select {
			case <-ticker.C:
				queue.Dequeue()
			case <-ctx.Done():
				return
			}
		}
	}()
	for i := 0; i < 99999; i++ {
		order := Refund{i}
		queue.Enqueue(order.Id, order)
		assert.True(t, atomic.LoadInt32(&queue.gCounts) <= 1)
	}
	f()
	time.Sleep(time.Second)
}

func TestQueueExists(t *testing.T) {
	var (
		queue = (&Queue[int, Refund]{}).Instance()
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
		}
	}()

	max := 99999
	for i := 0; i < max; i++ {
		order := Refund{1}
		queue.Enqueue(order.Id, order)
	}

	assert.Equal(t, max, queue.m[1])
	i := 0
	for queue.Size() > 0 {
		assert.True(t, queue.Exists(1))
		assert.Equal(t, max-i, queue.m[1])
		queue.Dequeue()
		i++
	}
	assert.Equal(t, 0, queue.m[1])
}

func TestQueueSequence(t *testing.T) {
	var (
		queue = (&Queue[int, Refund]{}).Instance()
	)
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			debug.PrintStack()
		}
	}()

	for i := 0; i < 99999; i++ {
		order := Refund{i}
		queue.Enqueue(order.Id, order)
	}
	i := 0
	for queue.Size() > 0 {
		assert.Equal(t, i, queue.Dequeue().Id)
		i++
	}
}
