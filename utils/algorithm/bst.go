package algorithm

type Priority interface {
	Less(i, j interface{}) bool
	Equal(i, j interface{}) bool
	Empty() bool
}

type BSTree[T Priority] struct {
	data  T
	Left  *BSTree[T]
	Right *BSTree[T]
}

// NewBSTNode Attention: data need implement Priority
func NewBSTNode[T Priority](data T) *BSTree[T] {
	return &BSTree[T]{
		data: data,
	}
}

func (p *BSTree[T]) Insert(newNode *BSTree[T]) *BSTree[T] {
	if newNode == nil {
		return p
	}
	if p == nil {
		return NewBSTNode(newNode.data)
	}

	if p.data.Empty() {
		panic("can't less nil data")
	}

	if p.data.Less(p.data, newNode.data) {
		if p.Right == nil {
			p.Right = newNode
		} else {
			p.Right.Insert(newNode)
		}
	} else {
		if p.Left == nil {
			p.Left = newNode
		} else {
			p.Left.Insert(newNode)
		}
	}
	return p
}

func (p *BSTree[T]) Remove(newNode *BSTree[T]) *BSTree[T] {
	if p.Left != nil {
		p.Left = p.Left.Remove(newNode)
	}

	if p.Right != nil {
		p.Right = p.Right.Remove(newNode)
	}

	// ATTENTION golang 无法修改 receiver 本身，需要 return
	if p.data.Equal(p.data, newNode.data) {
		if p.Left == nil && p.Right == nil {
			return nil
		}

		if p.Left == nil || p.Right == nil {
			if p.Left != nil {
				return p.Left
			} else {
				return p.Right
			}
		}

		if p.Left != nil && p.Right != nil {
			p.Right.Insert(p.Left)
			return p.Right
		}
	}

	return p
}

func (p *BSTree[T]) Has(newNode *BSTree[T]) bool {
	if p == nil {
		return false
	}

	if p.data.Empty() {
		panic("can't less nil data")
	}

	if p.Left != nil {
		found := p.Left.Has(newNode)
		if found {
			return true
		}
	}

	if p.Right != nil {
		found := p.Right.Has(newNode)
		if found {
			return true
		}
	}

	if p.data.Equal(p.data, newNode.data) {
		return true
	}
	return false
}

func (p *BSTree[T]) Minimum() *BSTree[T] {
	if p.Left == nil {
		return p
	} else {
		return p.Left.Minimum()
	}
}

func (p *BSTree[T]) Maximum() *BSTree[T] {
	if p.Right == nil {
		return p
	} else {
		return p.Right.Maximum()
	}
}

func (p *BSTree[T]) Data() T {
	return p.data
}

func (p *BSTree[T]) ToArray() []T {
	result := make([]T, 0)
	p.PreTravel(&result)
	return result
}

func (p *BSTree[T]) PreTravel(list *[]T) {
	if p == nil {
		return
	}

	if p.Left != nil {
		p.Left.PreTravel(list)
	}

	*list = append(*list, p.data)

	if p.Right != nil {
		p.Right.PreTravel(list)
	}
}

func (p *BSTree[T]) Merge(b *BSTree[T]) *BSTree[T] {
	if p == nil {
		return b
	} else {
		list := b.ToArray()
		for i := range list {
			p.Insert(NewBSTNode(list[i]))
		}
	}
	return p
}
