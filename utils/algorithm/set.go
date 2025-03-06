package algorithm

import (
	"reflect"
)

type SetType interface {
	~string | ~int | ~uint32 | ~int64
}

type CanHashUnique[T SetType] interface {
	Unique() T
}

type Set[T SetType] struct {
	set map[T]struct{}
}

func Instance[T SetType]() *Set[T] {
	ss := Set[T]{make(map[T]struct{})}
	return &ss
}

func InstanceSetAndMapFromStructSlice[T SetType, K CanHashUnique[T]](slice []K) (*Set[T], map[T]K) {
	s := Instance[T]()
	mapper := make(map[T]K)
	for i := range slice {
		uniqueValue := slice[i].Unique()
		mapper[uniqueValue] = slice[i]
		if !s.Has(uniqueValue) {
			s = s.Insert(uniqueValue)
		}
	}

	return s, mapper
}

func InstanceFromSlice[T SetType](slice *[]T) *Set[T] {
	s := Instance[T]()
	for i := range *slice {
		s = s.Insert((*slice)[i])
	}
	return s
}

func (s *Set[T]) Clear() {
	for k := range s.set {
		delete(s.set, k)
	}
}

func (s *Set[T]) Empty() bool {
	return len(s.set) == 0
}

func (s *Set[T]) Size() uint {
	if s == nil {
		return 0
	}
	return uint(len(s.set))
}

func (s *Set[T]) Insert(val T) *Set[T] {
	if s == nil || s.set == nil {
		set := Set[T]{set: make(map[T]struct{})}
		s = &set
	}

	s.set[val] = struct{}{}
	return s
}

func (s *Set[T]) TryInsert(val T) (*Set[T], bool) {
	// 是否不存在并插入成功
	var insertSuccess bool
	if s == nil || s.set == nil {
		set := Set[T]{make(map[T]struct{})}
		s = &set // 无效的，务必提前创建 Set
		insertSuccess = true
	} else {
		_, ok := s.set[val]
		insertSuccess = !ok
		s.set[val] = struct{}{}
	}

	return s, insertSuccess
}

func (s *Set[T]) InsertAll(val ...T) *Set[T] {
	if s == nil || s.set == nil {
		set := Set[T]{make(map[T]struct{})}
		s = &set
	}
	for i := range val {
		s.set[val[i]] = struct{}{}
	}
	return s
}

func (s *Set[T]) Remove(val T) {
	delete(s.set, val)
}

func (s *Set[T]) Has(val T) bool {
	_, exists := s.set[val]
	return exists
}

func (s *Set[T]) HasAnyItem(val ...T) bool {
	for i := range val {
		if _, exists := s.set[val[i]]; exists {
			return true
		}
	}
	return false
}

func (s *Set[T]) ToArray() []T {
	result := make([]T, s.Size())
	index := 0
	for k := range s.set {
		result[index] = k
		index++
	}
	return result
}

func (s *Set[T]) Clone() *Set[T] {
	result := Set[T]{make(map[T]struct{}, s.Size())}
	reflect.Copy(reflect.ValueOf(s), reflect.ValueOf(&result))
	return &result
}
