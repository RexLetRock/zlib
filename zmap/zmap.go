// Map Superfast With : Int Array
package zmap

func New[T Item](size int) *ZMap[T] {
  tr := new(ZMap[T])
  tr.items = make([]T, size)
  tr.size = size
  for i := range tr.items {
    tr.items[i] = *new(T)
  }
  return tr
}

func (tr *ZMap[T]) Add(item T) (T, bool) {
  tr.items[tr.count.inc() - 1] = item
  return item, true
}

func (tr *ZMap[T]) SetAt(index int, item T) (T, bool) {
  if index >= tr.size {
    return *new(T), false
  }
  tr.items[index] = item
  return item, true
}

func (tr *ZMap[T]) GetAt(index int) (T, bool) {
  if index >= tr.size {
    return *new(T), false
  }
  return tr.items[index], true
}

func (tr *ZMap[T]) GetAll() []T {
  result := tr.items[0:tr.count.get()]
  tr.mutex.Lock()
  tr.items = make([]T, tr.size)
  tr.count.reset()
  tr.mutex.Unlock()
  return result
}

func (tr *ZMap[T]) Clean() {
  tr.mutex.Lock()
  tr.items = make([]T, tr.size)
  tr.count.reset()
  tr.mutex.Unlock()
}

// Fast function, not check success
func (tr *ZMap[T]) ZSetAt(index int, item T) T {
  if index >= tr.size {
    return *new(T)
  }
  tr.items[index] = item
  return item
}

func (tr *ZMap[T]) ZGetAt(index int) T {
  if index >= tr.size {
    return *new(T)
  }
  return tr.items[index]
}
