package dict

//redis数据结构
//用空接口存各种类型的val
type Consumer func(key string, val interface{}) bool

type Dict interface {
	Get(key string) (val interface{}, exists bool)
	Len() int
	Put(key string, val interface{}) (result int)         //存入k-v
	PutIfAbsent(key string, val interface{}) (result int) //如果没有则存入
	PutIfExists(key string, val interface{}) (result int) //如果有则存入
	Remove(key string) (result int)                       //删除
	ForEach(consumer Consumer)                            //遍历
	Keys() []string                                       //把所有的Key列出
	RandomKeys(limit int) []string                        //返回限制数量key
	RandomDistinctKeys(limit int) []string                //返回不重复key
	Clear()                                               //清空

}
