package dict

import "sync"

type SyncDict struct {
	m sync.Map
}

func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

// 将输入存入到map中
func (dict *SyncDict) Get(key string) (val interface{}, exists bool) {
	val, ok := dict.m.Load(key) //load方法
	return val, ok
}

// 获取长度
func (dict *SyncDict) Len() int {
	lenth := 0
	dict.m.Range(func(k, v interface{}) bool {
		lenth++
		return true
	})
	return lenth
}

// 插入值，当return为0说名是修改
func (dict *SyncDict) Put(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Store(key, val)
	if existed {
		return 0
	}
	return 1
}

// 不存在则插入
func (dict *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		return 0
	}
	dict.m.Store(key, val)
	return 1
}

// 存在则插入
func (dict *SyncDict) PutIfExists(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		dict.m.Store(key, val)
		return 1
	}
	return 0
}

// 删除
func (dict *SyncDict) Remove(key string) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Delete(key)
	if existed {
		return 1
	}
	return 0
}

// 查询特定的键
func (dict *SyncDict) Keys() []string {
	result := make([]string, dict.Len())
	i := 0
	dict.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string) //遍历，将所有符合的key传入result
		i++
		return true
	})
	return result
}

// 将cunsumer施加到每一个键，遍历
func (dict *SyncDict) ForEach(consumer Consumer) {
	dict.m.Range(func(key, value interface{}) bool {
		consumer(key.(string), value)
		return true
	})
}

// 返回限制数量key
func (dict *SyncDict) RandomKeys(limit int) []string {
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		dict.m.Range(func(key, value interface{}) bool {
			result[i] = key.(string)
			return false
		})
	}
	return result

}

// 返回不相同key
func (dict *SyncDict) RandomDistinctKeys(limit int) []string {
	result := make([]string, limit)
	i := 0
	dict.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string)
		i++
		return i != limit
	})
	return result
}

// 清空
func (dict *SyncDict) Clear() {
	*dict = *MakeSyncDict() //重建一个新的dict，旧的让系统GC
}
