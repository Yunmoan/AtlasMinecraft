package tickqueue

// The native nextTickForContainer is fastutil's Long2LongOpenHashMap with
// default capacity/load factor. Its iteration and iterator-removal behavior
// are observable when two chunks' head ticks have equal priority and order
// (in particular the negative orders unpacked from saved NBT). A Go map or a
// sorted chunk list changes those ties. This small private table implements
// only the operations needed by LevelTicks, including wrapped iterator keys.
type deadlineMap struct {
	keys   []uint64
	values []int64
	size   int
	zero   bool
}

func newDeadlineMap() deadlineMap {
	return deadlineMap{keys: make([]uint64, 33), values: make([]int64, 33)}
}
func deadlineHash(k uint64) uint64 { h := k * 0x9e3779b97f4a7c15; h ^= h >> 32; return h ^ (h >> 16) }
func (m *deadlineMap) n() int      { return len(m.keys) - 1 }
func (m *deadlineMap) find(k uint64) (int, bool) {
	if k == 0 {
		return m.n(), m.zero
	}
	mask := m.n() - 1
	p := int(deadlineHash(k) & uint64(mask))
	for m.keys[p] != 0 {
		if m.keys[p] == k {
			return p, true
		}
		p = (p + 1) & mask
	}
	return p, false
}
func (m *deadlineMap) put(k uint64, v int64) {
	p, found := m.find(k)
	m.keys[p] = k
	m.values[p] = v
	if found {
		return
	}
	if k == 0 {
		m.zero = true
	}
	m.size++
	if m.size > m.n()*3/4 {
		m.rehash(m.n() * 2)
	}
}
func (m *deadlineMap) rehash(n int) {
	oldKeys, oldValues := m.keys, m.values
	m.keys = make([]uint64, n+1)
	m.values = make([]int64, n+1)
	for i := len(oldKeys) - 2; i >= 0; i-- {
		if k := oldKeys[i]; k != 0 {
			p, _ := m.find(k)
			m.keys[p] = k
			m.values[p] = oldValues[i]
		}
	}
	m.values[n] = oldValues[len(oldValues)-1]
}
func (m *deadlineMap) shift(pos int, wrapped *[]uint64) {
	mask := m.n() - 1
	for {
		last := pos
		pos = (last + 1) & mask
		for {
			k := m.keys[pos]
			if k == 0 {
				m.keys[last] = 0
				return
			}
			slot := int(deadlineHash(k) & uint64(mask))
			if last <= pos && (last >= slot || slot > pos) || last > pos && last >= slot && slot > pos {
				break
			}
			pos = (pos + 1) & mask
		}
		if wrapped != nil && pos < last {
			*wrapped = append(*wrapped, m.keys[pos])
		}
		m.keys[last] = m.keys[pos]
		m.values[last] = m.values[pos]
	}
}
func (m *deadlineMap) remove(k uint64) {
	p, found := m.find(k)
	if !found {
		return
	}
	if k == 0 {
		m.zero = false
	} else {
		m.shift(p, nil)
	}
	m.size--
	if m.n() > 32 && m.size < (m.n()*3/4)/4 {
		m.rehash(m.n() / 2)
	}
}

type deadlineIterator struct {
	m                           *deadlineMap
	pos, last, remaining, index int
	zero                        bool
	wrapped                     []uint64
}

func (m *deadlineMap) iterator() deadlineIterator {
	return deadlineIterator{m: m, pos: m.n(), last: -1, remaining: m.size, zero: m.zero}
}
func (i *deadlineIterator) next() bool {
	if i.remaining == 0 {
		return false
	}
	i.remaining--
	if i.zero {
		i.zero = false
		i.last = i.m.n()
		i.index = i.last
		return true
	}
	for {
		i.pos--
		if i.pos < 0 {
			i.last = -2
			i.index, _ = i.m.find(i.wrapped[-i.pos-1])
			return true
		}
		if i.m.keys[i.pos] != 0 {
			i.last = i.pos
			i.index = i.pos
			return true
		}
	}
}
func (i *deadlineIterator) entry() (uint64, int64) { return i.m.keys[i.index], i.m.values[i.index] }
func (i *deadlineIterator) setValue(v int64)       { i.m.values[i.index] = v }
func (i *deadlineIterator) remove() {
	if i.last == i.m.n() {
		i.m.zero = false
	} else if i.pos >= 0 {
		i.m.shift(i.last, &i.wrapped)
	} else {
		i.m.remove(i.wrapped[-i.pos-1])
		i.last = -1
		return
	}
	i.m.size--
	i.last = -1
}
