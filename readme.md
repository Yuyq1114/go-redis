# Redis协议
REdis Serialization Protocol(RESP)  
## 通信格式
### 正常回复
以"+"开头，以"\r\n"结尾的字符串形式|+OK\r\n    

### 错误回复
以"-"开头，以"\r\n"结尾的字符串形式|-false\r\n    

### 整数
以":"开头，以"\r\n"结尾的字符串形式|:12345\r\n  

### 多行字符串
以"$"开头，后跟实际发送字节数，以"\r\n"结尾的字符串形式    
$4\r\nmooc\r\n
 
### 数组
以"*"开头，后跟成员个数。  
SET key value  
*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n  
*2\r\n$6\r\nselect\r\n$1\r\n1\r\n


# 流程
1. tcp层服务TCP连接，交给resp->handler进行处理
2. handler监听连接，交给resp->parser对发送的协议进行解析，通过管道输出
3. 将指令转给database->database（exec方法）,如果是select则执行select，其他命令则交给database->db(用户的exec方法)
4. database->db将输入的指令从对应的指令表里寻找对应的方法执行，database->keys|ping|string
5. aof->aof实现持久化
6. 实现集群redis,lib->consistenthash存储hash结构，cluster中存储实现


# redis持久化策略
## AOF
1. 将存储在内存中的数据以文件的形式存储在硬盘上的。这个文件我们称之为AOF文件.
2. 存储的数据是客户端连接提交给Redis执行的写命令
## RDB
1. 保存了 Redis 在某个时间点上的数据集



# 集群redis
## 一致性哈希
1. 传统的hash会导致当redis结点不够需要扩充结点时，hash中mod会出问题，需要重新分布  
2. 一致性hash|一个环状，算出n个结点的位置，当有key过来时算出hash不取mod，根据hash放在环上位置在不同的结点

## 集群
1. standalone_database//单机版database做业务
2. cluster_database //集群database做转发
3. 需要一个客户端client，实现每个结点都是客户端也是服务端
4. 需要一个连接池维持多个连接，保证并发

## 执行模式
1. 本地执行，如ping
2. 转发执行，如get/set
3. 群发模式，如flush


