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

type SetV2[T SetType] struct {
	set map[T]struct{}
}

func (s *SetV2[T]) Instance() *SetV2[T] {
	ss := SetV2[T]{make(map[T]struct{})}
	return &ss
}

func InstanceSetAndMapFromStructSlice[T SetType, K CanHashUnique[T]](slice []K) (*SetV2[T], map[T]K) {
	s := (&SetV2[T]{}).Instance()
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

func (s *SetV2[T]) InstanceFromSlice(slice *[]T) *SetV2[T] {
	for i := range *slice {
		s = s.Insert((*slice)[i])
	}
	return s
}

func (s *SetV2[T]) Clear() {
	for k := range s.set {
		delete(s.set, k)
	}
}

func (s *SetV2[T]) Empty() bool {
	return len(s.set) == 0
}

func (s *SetV2[T]) Size() uint {
	if s == nil {
		return 0
	}
	return uint(len(s.set))
}

func (s *SetV2[T]) Insert(val T) *SetV2[T] {
	if s == nil || s.set == nil {
		set := SetV2[T]{set: make(map[T]struct{})}
		s = &set
	}

	s.set[val] = struct{}{}
	return s
}

func (s *SetV2[T]) TryInsert(val T) (*SetV2[T], bool) {
	// 是否不存在并插入成功
	var insertSuccess bool
	if s == nil || s.set == nil {
		set := SetV2[T]{make(map[T]struct{})}
		s = &set // 无效的，务必提前创建 SetV2
		insertSuccess = true
	} else {
		_, ok := s.set[val]
		insertSuccess = !ok
		s.set[val] = struct{}{}
	}

	return s, insertSuccess
}

func (s *SetV2[T]) InsertAll(val ...T) *SetV2[T] {
	if s == nil || s.set == nil {
		set := SetV2[T]{make(map[T]struct{})}
		s = &set
	}
	for i := range val {
		s.set[val[i]] = struct{}{}
	}
	return s
}

func (s *SetV2[T]) Remove(val T) {
	delete(s.set, val)
}

func (s *SetV2[T]) Has(val T) bool {
	_, exists := s.set[val]
	return exists
}

func (s *SetV2[T]) HasAnyItem(val ...T) bool {
	for i := range val {
		if _, exists := s.set[val[i]]; exists {
			return true
		}
	}
	return false
}

func (s *SetV2[T]) ToArray() []T {
	result := make([]T, s.Size())
	index := 0
	for k := range s.set {
		result[index] = k
		index++
	}
	return result
}

func (s *SetV2[T]) Clone() *SetV2[T] {
	result := SetV2[T]{make(map[T]struct{}, s.Size())}
	reflect.Copy(reflect.ValueOf(s), reflect.ValueOf(&result))
	return &result
}
