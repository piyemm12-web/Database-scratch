package memtable

import (
	"bytes"
	"math/rand"
	"time"
)

const (
	maxLevel    = 16
	probability = 0.5
)

type Node struct {
	key   []byte
	value []byte
	forward []*Node
}

type SkipList struct {
	header *Node
	level  int
	length int
	rnd    *rand.Rand
}

func NewSkipList() *SkipList {
	return &SkipList{
		header: &Node{forward: make([]*Node, maxLevel)},
		level:  1,
		rnd:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *SkipList) randomLevel() int {
	lvl := 1
	for s.rnd.Float64() < probability && lvl < maxLevel {
		lvl++
	}
	return lvl
}

func (s *SkipList) Put(key, value []byte) {
	update := make([]*Node, maxLevel)
	curr := s.header

	for i := s.level - 1; i >= 0; i-- {
		for curr.forward[i] != nil && bytes.Compare(curr.forward[i].key, key) < 0 {
			curr = curr.forward[i]
		}
		update[i] = curr
	}

	curr = curr.forward[0]

	if curr != nil && bytes.Equal(curr.key, key) {
		curr.value = value
		return
	}

	newLvl := s.randomLevel()
	if newLvl > s.level {
		for i := s.level; i < newLvl; i++ {
			update[i] = s.header
		}
		s.level = newLvl
	}

	newNode := &Node{
		key:     key,
		value:   value,
		forward: make([]*Node, newLvl),
	}

	for i := 0; i < newLvl; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	s.length++
}

func (s *SkipList) Get(key []byte) ([]byte, bool) {
	curr := s.header

	for i := s.level - 1; i >= 0; i-- {
		for curr.forward[i] != nil && bytes.Compare(curr.forward[i].key, key) < 0 {
			curr = curr.forward[i]
		}
	}

	curr = curr.forward[0]
	if curr != nil && bytes.Equal(curr.key, key) {
		return curr.value, true
	}

	return nil, false
}

func (s *SkipList) Entries() []struct{ Key, Value []byte } {
	var entries []struct{ Key, Value []byte }
	curr := s.header.forward[0]

	for curr != nil {
		entries = append(entries, struct{ Key, Value []byte }{Key: curr.key, Value: curr.value})
		curr = curr.forward[0]
	}

	return entries
}