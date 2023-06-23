package consistenthash

import (
	"hash/crc32"
	"sort"
)

type HashFunc func(data []byte) uint32

// 一致性hash结点
type NodeMap struct {
	hashFunc    HashFunc       //存储hash函数
	nodeHashs   []int          //存储结点hash值
	nodehashMap map[int]string //存储结点位置
}

// 新建
func NewNodeMap(fn HashFunc) *NodeMap {
	m := &NodeMap{
		hashFunc:    fn,
		nodehashMap: make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE //如果未传入方法则默认crc32方法
	}
	return m
}

// 判断是否空一致性hash
func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

// 增加结点
func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(key)))
		m.nodeHashs = append(m.nodeHashs, hash)
		m.nodehashMap[hash] = key
	}
	sort.Ints(m.nodeHashs)
}

// 选择结点
func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}
	hash := int(m.hashFunc([]byte(key)))
	//根据hash值去nodeHashs中选择
	idx := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})
	if idx == len(m.nodeHashs) {
		idx = 0
	}
	return m.nodehashMap[m.nodeHashs[idx]]
}
