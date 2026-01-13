package algorithm

//
//import "container/heap"
//
//type Priority1Queue[T any, K Comparable[T]] interface {
//	heap.Interface
//	Peek() any
//}
//
//type Comparable[T any] interface {
//	Compare(i, j T) bool
//}
//
//type Item struct {
//	value    string // The value of the item; arbitrary.
//	priority int    // The priority of the item in the queue.
//	// The index is needed by update and is maintained by the heap.Interface methods.
//	index int // The index of the item in the heap.
//}
//
//type priorityQueue[T any] struct {
//	items []T
//}
//
//func (pq *priorityQueue) Len() int { return len(pq.items) }
//
//func (pq *priorityQueue) Less(i, j int) bool {
//	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
//	return (*pq)[i].priority > (*pq)[j].priority
//}
//
//func (pq *priorityQueue) Swap(i, j int) {
//	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
//	(*pq)[i].index = i
//	(*pq)[j].index = j
//}
//
//func (pq *priorityQueue) Push(x any) {
//	n := len(*pq)
//	item := x.(*Item)
//	item.index = n
//	*pq = append(*pq, item)
//}
//
//func (pq *priorityQueue) Pop() any {
//	old := *pq
//	n := len(old)
//	item := old[n-1]
//	old[n-1] = nil  // don't stop the GC from reclaiming the item eventually
//	item.index = -1 // for safety
//	*pq = old[0 : n-1]
//	return item
//}
//
//func (pq *priorityQueue) Peek() any {
//	return nil
//}
//
//// update modifies the priority and value of an Item in the queue.
//func (pq *priorityQueue) update(item *Item, value string, priority int) {
//	item.value = value
//	item.priority = priority
//	heap.Fix(pq, item.index)
//}
