package zdb

import "sync"

const (
	degree   = 128
	maxItems = degree*2 - 1
	minItems = maxItems / 2
)

type bItem interface {
  IID() int
  ToString() string
}

type Btree[T bItem] struct {
	mu    *sync.RWMutex
	cow   *cow
	root  *node[T]
	count int
	locks bool
	less  func(a, b T) bool
	empty T
}

type node[T bItem] struct {
	cow      *cow
	count    int
	items    []T
	children *[]*node[T]
}

type cow struct {
	_ int
}

type PathHint struct {
	used [8]bool
	path [8]uint8
}

type Options struct {
	NoLocks bool
}

func NewBtree[T bItem](less func(a, b T) bool) *Btree[T] {
	return NewBtreeOptions(less, Options{})
}

func NewBtreeOptions[T bItem](less func(a, b T) bool, opts Options) *Btree[T] {
	tr := new(Btree[T])
	tr.cow = new(cow)
	tr.mu = new(sync.RWMutex)
	tr.less = less
	tr.locks = !opts.NoLocks
	return tr
}

func (tr *Btree[T]) Less(a, b T) bool {
	return tr.less(a, b)
}

func (tr *Btree[T]) newNode(leaf bool) *node[T] {
	n := &node[T]{cow: tr.cow}
	if !leaf {
		n.children = new([]*node[T])
	}
	return n
}

func (n *node[T]) leaf() bool {
	return n.children == nil
}

func (tr *Btree[T]) find(n *node[T], key T, hint *PathHint, depth int,
) (index int, found bool) {
	if hint == nil {
		low := 0
		high := len(n.items)
		for low < high {
			mid := (low + high) / 2
			if !tr.Less(key, n.items[mid]) {
				low = mid + 1
			} else {
				high = mid
			}
		}
		if low > 0 && !tr.Less(n.items[low-1], key) {
			return low - 1, true
		}
		return low, false
	}

	low := 0
	high := len(n.items) - 1
	if depth < 8 && hint.used[depth] {
		index = int(hint.path[depth])
		if index >= len(n.items) {
			if tr.Less(n.items[len(n.items)-1], key) {
				index = len(n.items)
				goto path_match
			}
			index = len(n.items) - 1
		}
		if tr.Less(key, n.items[index]) {
			if index == 0 || tr.Less(n.items[index-1], key) {
				goto path_match
			}
			high = index - 1
		} else if tr.Less(n.items[index], key) {
			low = index + 1
		} else {
			found = true
			goto path_match
		}
	}

	for low <= high {
		mid := low + ((high+1)-low)/2
		if !tr.Less(key, n.items[mid]) {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	if low > 0 && !tr.Less(n.items[low-1], key) {
		index = low - 1
		found = true
	} else {
		index = low
		found = false
	}

path_match:
	if depth < 8 {
		hint.used[depth] = true
		var pathIndex uint8
		if n.leaf() && found {
			pathIndex = uint8(index + 1)
		} else {
			pathIndex = uint8(index)
		}
		if pathIndex != hint.path[depth] {
			hint.path[depth] = pathIndex
			for i := depth + 1; i < 8; i++ {
				hint.used[i] = false
			}
		}
	}
	return index, found
}

func (tr *Btree[T]) SetHint(item T, hint *PathHint) (prev T, replaced bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	return tr.setHint(item, hint)
}

func (tr *Btree[T]) setHint(item T, hint *PathHint) (prev T, replaced bool) {
	if tr.root == nil {
		tr.root = tr.newNode(true)
		tr.root.items = append([]T{}, item)
		tr.root.count = 1
		tr.count = 1
		return tr.empty, false
	}
	prev, replaced, split := tr.nodeSet(&tr.root, item, hint, 0)
	if split {
		left := tr.cowLoad(&tr.root)
		right, median := tr.nodeSplit(left)
		tr.root = tr.newNode(false)
		*tr.root.children = make([]*node[T], 0, maxItems+1)
		*tr.root.children = append([]*node[T]{}, left, right)
		tr.root.items = append([]T{}, median)
		tr.root.updateCount()
		return tr.setHint(item, hint)
	}
	if replaced {
		return prev, true
	}
	tr.count++
	return tr.empty, false
}

func (tr *Btree[T]) Set(item T) (T, bool) {
	return tr.SetHint(item, nil)
}

func (tr *Btree[T]) nodeSplit(n *node[T]) (right *node[T], median T) {
	i := maxItems / 2
	median = n.items[i]

	// left node
	left := tr.newNode(n.leaf())
	left.items = make([]T, len(n.items[:i]), maxItems/2)
	copy(left.items, n.items[:i])
	if !n.leaf() {
		*left.children = make([]*node[T], len((*n.children)[:i+1]), maxItems+1)
		copy(*left.children, (*n.children)[:i+1])
	}
	left.updateCount()

	// right node
	right = tr.newNode(n.leaf())
	right.items = make([]T, len(n.items[i+1:]), maxItems/2)
	copy(right.items, n.items[i+1:])
	if !n.leaf() {
		*right.children = make([]*node[T], len((*n.children)[i+1:]), maxItems+1)
		copy(*right.children, (*n.children)[i+1:])
	}
	right.updateCount()

	*n = *left
	return right, median
}

func (n *node[T]) updateCount() {
	n.count = len(n.items)
	if !n.leaf() {
		for i := 0; i < len(*n.children); i++ {
			n.count += (*n.children)[i].count
		}
	}
}

func (tr *Btree[T]) copy(n *node[T]) *node[T] {
	n2 := new(node[T])
	n2.cow = tr.cow
	n2.count = n.count
	n2.items = make([]T, len(n.items), cap(n.items))
	copy(n2.items, n.items)
	if !n.leaf() {
		n2.children = new([]*node[T])
		*n2.children = make([]*node[T], len(*n.children), maxItems+1)
		copy(*n2.children, *n.children)
	}
	return n2
}

func (tr *Btree[T]) cowLoad(cn **node[T]) *node[T] {
	if (*cn).cow != tr.cow {
		*cn = tr.copy(*cn)
	}
	return *cn
}

func (tr *Btree[T]) nodeSet(cn **node[T], item T,
	hint *PathHint, depth int,
) (prev T, replaced bool, split bool) {
	n := tr.cowLoad(cn)
	i, found := tr.find(n, item, hint, depth)
	if found {
		prev = n.items[i]
		n.items[i] = item
		return prev, true, false
	}
	if n.leaf() {
		if len(n.items) == maxItems {
			return tr.empty, false, true
		}
		n.items = append(n.items, tr.empty)
		copy(n.items[i+1:], n.items[i:])
		n.items[i] = item
		n.count++
		return tr.empty, false, false
	}
	prev, replaced, split = tr.nodeSet(&(*n.children)[i], item, hint, depth+1)
	if split {
		if len(n.items) == maxItems {
			return tr.empty, false, true
		}
		right, median := tr.nodeSplit((*n.children)[i])
		*n.children = append(*n.children, nil)
		copy((*n.children)[i+1:], (*n.children)[i:])
		(*n.children)[i+1] = right
		n.items = append(n.items, tr.empty)
		copy(n.items[i+1:], n.items[i:])
		n.items[i] = median
		return tr.nodeSet(&n, item, hint, depth)
	}
	if !replaced {
		n.count++
	}
	return prev, replaced, false
}

func (tr *Btree[T]) Scan(iter func(item T) bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return
	}
	tr.root.scan(iter)
}

func (n *node[T]) scan(iter func(item T) bool) bool {
	if n.leaf() {
		for i := 0; i < len(n.items); i++ {
			if !iter(n.items[i]) {
				return false
			}
		}
		return true
	}
	for i := 0; i < len(n.items); i++ {
		if !(*n.children)[i].scan(iter) {
			return false
		}
		if !iter(n.items[i]) {
			return false
		}
	}
	return (*n.children)[len(*n.children)-1].scan(iter)
}

func (tr *Btree[T]) Get(key T) (T, bool) {
	return tr.GetHint(key, nil)
}

func (tr *Btree[T]) GetHint(key T, hint *PathHint) (T, bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return tr.empty, false
	}
	n := tr.root
	depth := 0
	for {
		i, found := tr.find(n, key, hint, depth)
		if found {
			return n.items[i], true
		}
		if n.children == nil {
			return tr.empty, false
		}
		n = (*n.children)[i]
		depth++
	}
}

func (tr *Btree[T]) Len() int {
	return tr.count
}

func (tr *Btree[T]) Delete(key T) (T, bool) {
	return tr.DeleteHint(key, nil)
}

func (tr *Btree[T]) DeleteHint(key T, hint *PathHint) (T, bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	return tr.deleteHint(key, hint)
}

func (tr *Btree[T]) deleteHint(key T, hint *PathHint) (T, bool) {
	if tr.root == nil {
		return tr.empty, false
	}
	prev, deleted := tr.delete(&tr.root, false, key, hint, 0)
	if !deleted {
		return tr.empty, false
	}
	if len(tr.root.items) == 0 && !tr.root.leaf() {
		tr.root = (*tr.root.children)[0]
	}
	tr.count--
	if tr.count == 0 {
		tr.root = nil
	}
	return prev, true
}

func (tr *Btree[T]) delete(cn **node[T], max bool, key T,
	hint *PathHint, depth int,
) (T, bool) {
	n := tr.cowLoad(cn)
	var i int
	var found bool
	if max {
		i, found = len(n.items)-1, true
	} else {
		i, found = tr.find(n, key, hint, depth)
	}
	if n.leaf() {
		if found {
			prev := n.items[i]
			copy(n.items[i:], n.items[i+1:])
			n.items[len(n.items)-1] = tr.empty
			n.items = n.items[:len(n.items)-1]
			n.count--
			return prev, true
		}
		return tr.empty, false
	}

	var prev T
	var deleted bool
	if found {
		if max {
			i++
			prev, deleted = tr.delete(&(*n.children)[i], true, tr.empty, nil, 0)
		} else {
			prev = n.items[i]
			maxItem, _ := tr.delete(&(*n.children)[i], true, tr.empty, nil, 0)
			deleted = true
			n.items[i] = maxItem
		}
	} else {
		prev, deleted = tr.delete(&(*n.children)[i], max, key, hint, depth+1)
	}
	if !deleted {
		return tr.empty, false
	}
	n.count--
	if len((*n.children)[i].items) < minItems {
		tr.nodeRebalance(n, i)
	}
	return prev, true
}

func (tr *Btree[T]) nodeRebalance(n *node[T], i int) {
	if i == len(n.items) {
		i--
	}

	left := tr.cowLoad(&(*n.children)[i])
	right := tr.cowLoad(&(*n.children)[i+1])

	if len(left.items)+len(right.items) < maxItems {
		left.items = append(left.items, n.items[i])
		left.items = append(left.items, right.items...)
		if !left.leaf() {
			*left.children = append(*left.children, *right.children...)
		}
		left.count += right.count + 1

		// move the items over one slot
		copy(n.items[i:], n.items[i+1:])
		n.items[len(n.items)-1] = tr.empty
		n.items = n.items[:len(n.items)-1]

		// move the children over one slot
		copy((*n.children)[i+1:], (*n.children)[i+2:])
		(*n.children)[len(*n.children)-1] = nil
		(*n.children) = (*n.children)[:len(*n.children)-1]
	} else if len(left.items) > len(right.items) {
		// move left -> right over one slot
		right.items = append(right.items, tr.empty)
		copy(right.items[1:], right.items)
		right.items[0] = n.items[i]
		right.count++
		n.items[i] = left.items[len(left.items)-1]
		left.items[len(left.items)-1] = tr.empty
		left.items = left.items[:len(left.items)-1]
		left.count--

		if !left.leaf() {
			// move the left-node last child into the right-node first slot
			*right.children = append(*right.children, nil)
			copy((*right.children)[1:], *right.children)
			(*right.children)[0] = (*left.children)[len(*left.children)-1]
			(*left.children)[len(*left.children)-1] = nil
			(*left.children) = (*left.children)[:len(*left.children)-1]
			left.count -= (*right.children)[0].count
			right.count += (*right.children)[0].count
		}
	} else {
		// move left <- right over one slot

		// Same as above but the other direction
		left.items = append(left.items, n.items[i])
		left.count++
		n.items[i] = right.items[0]
		copy(right.items, right.items[1:])
		right.items[len(right.items)-1] = tr.empty
		right.items = right.items[:len(right.items)-1]
		right.count--

		if !left.leaf() {
			*left.children = append(*left.children, (*right.children)[0])
			copy(*right.children, (*right.children)[1:])
			(*right.children)[len(*right.children)-1] = nil
			*right.children = (*right.children)[:len(*right.children)-1]
			left.count += (*left.children)[len(*left.children)-1].count
			right.count -= (*left.children)[len(*left.children)-1].count
		}
	}
}

func (tr *Btree[T]) Ascend(pivot T, iter func(item T) bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return
	}
	tr.ascend(tr.root, pivot, nil, 0, iter)
}

func (tr *Btree[T]) ascend(n *node[T], pivot T,
	hint *PathHint, depth int, iter func(item T) bool,
) bool {
	i, found := tr.find(n, pivot, hint, depth)
	if !found {
		if !n.leaf() {
			if !tr.ascend((*n.children)[i], pivot, hint, depth+1, iter) {
				return false
			}
		}
	}

  for ; i < len(n.items); i++ {
		if !iter(n.items[i]) {
			return false
		}
		if !n.leaf() {
			if !(*n.children)[i+1].scan(iter) {
				return false
			}
		}
	}
	return true
}

func (tr *Btree[T]) Reverse(iter func(item T) bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return
	}
	tr.root.reverse(iter)
}

func (n *node[T]) reverse(iter func(item T) bool) bool {
	if n.leaf() {
		for i := len(n.items) - 1; i >= 0; i-- {
			if !iter(n.items[i]) {
				return false
			}
		}
		return true
	}
	if !(*n.children)[len(*n.children)-1].reverse(iter) {
		return false
	}
	for i := len(n.items) - 1; i >= 0; i-- {
		if !iter(n.items[i]) {
			return false
		}
		if !(*n.children)[i].reverse(iter) {
			return false
		}
	}
	return true
}

func (tr *Btree[T]) Descend(pivot T, iter func(item T) bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return
	}
	tr.descend(tr.root, pivot, nil, 0, iter)
}

func (tr *Btree[T]) descend(n *node[T], pivot T,
	hint *PathHint, depth int, iter func(item T) bool,
) bool {
	i, found := tr.find(n, pivot, hint, depth)
	if !found {
		if !n.leaf() {
			if !tr.descend((*n.children)[i], pivot, hint, depth+1, iter) {
				return false
			}
		}
		i--
	}
	for ; i >= 0; i-- {
		if !iter(n.items[i]) {
			return false
		}
		if !n.leaf() {
			if !(*n.children)[i].reverse(iter) {
				return false
			}
		}
	}
	return true
}

func (tr *Btree[T]) Load(item T) (T, bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	if tr.root == nil {
		return tr.setHint(item, nil)
	}
	n := tr.cowLoad(&tr.root)
	for {
		n.count++ // optimistically update counts
		if n.leaf() {
			if len(n.items) < maxItems {
				if tr.Less(n.items[len(n.items)-1], item) {
					n.items = append(n.items, item)
					tr.count++
					return tr.empty, false
				}
			}
			break
		}
		n = tr.cowLoad(&(*n.children)[len(*n.children)-1])
	}
	// revert the counts
	n = tr.root
	for {
		n.count--
		if n.leaf() {
			break
		}
		n = (*n.children)[len(*n.children)-1]
	}
	return tr.setHint(item, nil)
}

func (tr *Btree[T]) Min() (T, bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return tr.empty, false
	}
	n := tr.root
	for {
		if n.leaf() {
			return n.items[0], true
		}
		n = (*n.children)[0]
	}
}

func (tr *Btree[T]) Max() (T, bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil {
		return tr.empty, false
	}
	n := tr.root
	for {
		if n.leaf() {
			return n.items[len(n.items)-1], true
		}
		n = (*n.children)[len(*n.children)-1]
	}
}

func (tr *Btree[T]) PopMin() (T, bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	if tr.root == nil {
		return tr.empty, false
	}
	n := tr.cowLoad(&tr.root)
	var item T
	for {
		n.count-- // optimistically update counts
		if n.leaf() {
			item = n.items[0]
			if len(n.items) == minItems {
				break
			}
			copy(n.items[:], n.items[1:])
			n.items[len(n.items)-1] = tr.empty
			n.items = n.items[:len(n.items)-1]
			tr.count--
			if tr.count == 0 {
				tr.root = nil
			}
			return item, true
		}
		n = tr.cowLoad(&(*n.children)[0])
	}
	// revert the counts
	n = tr.root
	for {
		n.count++
		if n.leaf() {
			break
		}
		n = (*n.children)[0]
	}
	return tr.deleteHint(item, nil)
}

func (tr *Btree[T]) PopMax() (T, bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	if tr.root == nil {
		return tr.empty, false
	}
	n := tr.cowLoad(&tr.root)
	var item T
	for {
		n.count-- // optimistically update counts
		if n.leaf() {
			item = n.items[len(n.items)-1]
			if len(n.items) == minItems {
				break
			}
			n.items[len(n.items)-1] = tr.empty
			n.items = n.items[:len(n.items)-1]
			tr.count--
			if tr.count == 0 {
				tr.root = nil
			}
			return item, true
		}
		n = tr.cowLoad(&(*n.children)[len(*n.children)-1])
	}
	// revert the counts
	n = tr.root
	for {
		n.count++
		if n.leaf() {
			break
		}
		n = (*n.children)[len(*n.children)-1]
	}
	return tr.deleteHint(item, nil)
}

func (tr *Btree[T]) GetAt(index int) (T, bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root == nil || index < 0 || index >= tr.count {
		return tr.empty, false
	}
	n := tr.root
	for {
		if n.leaf() {
			return n.items[index], true
		}
		i := 0
		for ; i < len(n.items); i++ {
			if index < (*n.children)[i].count {
				break
			} else if index == (*n.children)[i].count {
				return n.items[i], true
			}
			index -= (*n.children)[i].count + 1
		}
		n = (*n.children)[i]
	}
}

func (tr *Btree[T]) DeleteAt(index int) (T, bool) {
	if tr.lock() {
		defer tr.unlock()
	}
	if tr.root == nil || index < 0 || index >= tr.count {
		return tr.empty, false
	}
	var pathbuf [8]uint8 // track the path
	path := pathbuf[:0]
	var item T
	n := tr.cowLoad(&tr.root)
outer:
	for {
		n.count-- // optimistically update counts
		if n.leaf() {
			// the index is the item position
			item = n.items[index]
			if len(n.items) == minItems {
				path = append(path, uint8(index))
				break outer
			}
			copy(n.items[index:], n.items[index+1:])
			n.items[len(n.items)-1] = tr.empty
			n.items = n.items[:len(n.items)-1]
			tr.count--
			if tr.count == 0 {
				tr.root = nil
			}
			return item, true
		}
		i := 0
		for ; i < len(n.items); i++ {
			if index < (*n.children)[i].count {
				break
			} else if index == (*n.children)[i].count {
				item = n.items[i]
				path = append(path, uint8(i))
				break outer
			}
			index -= (*n.children)[i].count + 1
		}
		path = append(path, uint8(i))
		n = tr.cowLoad(&(*n.children)[i])
	}
	// revert the counts
	var hint PathHint
	n = tr.root
	for i := 0; i < len(path); i++ {
		if i < len(hint.path) {
			hint.path[i] = uint8(path[i])
			hint.used[i] = true
		}
		n.count++
		if !n.leaf() {
			n = (*n.children)[uint8(path[i])]
		}
	}
	return tr.deleteHint(item, &hint)
}

func (tr *Btree[T]) Height() int {
	if tr.rlock() {
		defer tr.runlock()
	}
	var height int
	if tr.root != nil {
		n := tr.root
		for {
			height++
			if n.leaf() {
				break
			}
			n = (*n.children)[0]
		}
	}
	return height
}

func (tr *Btree[T]) Walk(iter func(item []T) bool) {
	if tr.rlock() {
		defer tr.runlock()
	}
	if tr.root != nil {
		tr.root.walk(iter)
	}
}

func (n *node[T]) walk(iter func(item []T) bool) bool {
	if n.leaf() {
		if !iter(n.items) {
			return false
		}
	} else {
		for i := 0; i < len(n.items); i++ {
			(*n.children)[i].walk(iter)
			if !iter(n.items[i : i+1]) {
				return false
			}
		}
		(*n.children)[len(n.items)].walk(iter)
	}
	return true
}

func (tr *Btree[T]) Copy() *Btree[T] {
	if tr.lock() {
		defer tr.unlock()
	}
	tr.cow = new(cow)
	tr2 := new(Btree[T])
	*tr2 = *tr
	tr2.mu = new(sync.RWMutex)
	tr2.cow = new(cow)
	return tr2
}

func (tr *Btree[T]) lock() bool {
	if tr.locks {
		tr.mu.Lock()
	}
	return tr.locks
}

func (tr *Btree[T]) unlock() {
	tr.mu.Unlock()
}

func (tr *Btree[T]) rlock() bool {
	if tr.locks {
		tr.mu.RLock()
	}
	return tr.locks
}

func (tr *Btree[T]) runlock() {
	tr.mu.RUnlock()
}

// Iter represents an iterator
type GenericIter[T bItem] struct {
	tr      *Btree[T]
	locked  bool
	seeked  bool
	atstart bool
	atend   bool
	stack   []genericIterStackItem[T]
	item    T
}

type genericIterStackItem[T bItem] struct {
	n *node[T]
	i int
}

func (tr *Btree[T]) Iter() GenericIter[T] {
	var iter GenericIter[T]
	iter.tr = tr
	iter.locked = tr.rlock()
	return iter
}

func (iter *GenericIter[T]) Seek(key T) bool {
	if iter.tr == nil {
		return false
	}
	iter.seeked = true
	iter.stack = iter.stack[:0]
	if iter.tr.root == nil {
		return false
	}
	n := iter.tr.root
	for {
		i, found := iter.tr.find(n, key, nil, 0)
		iter.stack = append(iter.stack, genericIterStackItem[T]{n, i})
		if found {
			iter.item = n.items[i]
			return true
		}
		if n.leaf() {
			iter.stack[len(iter.stack)-1].i--
			return iter.Next()
		}
		n = (*n.children)[i]
	}
}

func (iter *GenericIter[T]) First() bool {
	if iter.tr == nil {
		return false
	}
	iter.atend = false
	iter.atstart = false
	iter.seeked = true
	iter.stack = iter.stack[:0]
	if iter.tr.root == nil {
		return false
	}
	n := iter.tr.root
	for {
		iter.stack = append(iter.stack, genericIterStackItem[T]{n, 0})
		if n.leaf() {
			break
		}
		n = (*n.children)[0]
	}
	s := &iter.stack[len(iter.stack)-1]
	iter.item = s.n.items[s.i]
	return true
}

func (iter *GenericIter[T]) Last() bool {
	if iter.tr == nil {
		return false
	}
	iter.seeked = true
	iter.stack = iter.stack[:0]
	if iter.tr.root == nil {
		return false
	}
	n := iter.tr.root
	for {
		iter.stack = append(iter.stack, genericIterStackItem[T]{n, len(n.items)})
		if n.leaf() {
			iter.stack[len(iter.stack)-1].i--
			break
		}
		n = (*n.children)[len(n.items)]
	}
	s := &iter.stack[len(iter.stack)-1]
	iter.item = s.n.items[s.i]
	return true
}

// Release the iterator.
func (iter *GenericIter[T]) Release() {
	if iter.tr == nil {
		return
	}
	if iter.locked {
		iter.tr.runlock()
		iter.locked = false
	}
	iter.stack = nil
	iter.tr = nil
}

func (iter *GenericIter[T]) Next() bool {
	if iter.tr == nil {
		return false
	}
	if !iter.seeked {
		return iter.First()
	}
	if len(iter.stack) == 0 {
		if iter.atstart {
			return iter.First() && iter.Next()
		}
		return false
	}
	s := &iter.stack[len(iter.stack)-1]
	s.i++
	if s.n.leaf() {
		if s.i == len(s.n.items) {
			for {
				iter.stack = iter.stack[:len(iter.stack)-1]
				if len(iter.stack) == 0 {
					iter.atend = true
					return false
				}
				s = &iter.stack[len(iter.stack)-1]
				if s.i < len(s.n.items) {
					break
				}
			}
		}
	} else {
		n := (*s.n.children)[s.i]
		for {
			iter.stack = append(iter.stack, genericIterStackItem[T]{n, 0})
			if n.leaf() {
				break
			}
			n = (*n.children)[0]
		}
	}
	s = &iter.stack[len(iter.stack)-1]
	iter.item = s.n.items[s.i]
	return true
}

func (iter *GenericIter[T]) Prev() bool {
	if iter.tr == nil {
		return false
	}
	if !iter.seeked {
		return false
	}
	if len(iter.stack) == 0 {
		if iter.atend {
			return iter.Last() && iter.Prev()
		}
		return false
	}
	s := &iter.stack[len(iter.stack)-1]
	if s.n.leaf() {
		s.i--
		if s.i == -1 {
			for {
				iter.stack = iter.stack[:len(iter.stack)-1]
				if len(iter.stack) == 0 {
					iter.atstart = true
					return false
				}
				s = &iter.stack[len(iter.stack)-1]
				s.i--
				if s.i > -1 {
					break
				}
			}
		}
	} else {
		n := (*s.n.children)[s.i]
		for {
			iter.stack = append(iter.stack, genericIterStackItem[T]{n, len(n.items)})
			if n.leaf() {
				iter.stack[len(iter.stack)-1].i--
				break
			}
			n = (*n.children)[len(n.items)]
		}
	}
	s = &iter.stack[len(iter.stack)-1]
	iter.item = s.n.items[s.i]
	return true
}

// Item returns the current iterator item.
func (iter *GenericIter[T]) Item() T {
	return iter.item
}

// Items returns all the items in order.
func (tr *Btree[T]) Items() []T {
	items := make([]T, 0, tr.Len())
	if tr.root != nil {
		items = tr.root.aitems(items)
	}
	return items
}

func (n *node[T]) aitems(items []T) []T {
	if n.leaf() {
		return append(items, n.items...)
	}
	for i := 0; i < len(n.items); i++ {
		items = (*n.children)[i].aitems(items)
		items = append(items, n.items[i])
	}
	return (*n.children)[len(*n.children)-1].aitems(items)
}
