## Go Rpc Demo

使用 Go 语言开发的 Rpc Demo 项目，使用 jsonrpc 来开启及调用 rpc 服务。

整体功能：

客户端读取本目录下的 `id.txt` ，然后调用服务端的接口逐行上报，服务端收到后再写入服务端目录下的 `id.txt`

## 服务端

服务端提供三个接口：`记录日志`、`删除日志文件`、`统计日志文件行数`

1. 记录日志

收到客户端调用后，将客户端发送的文件名称和消息以 `:` 号连接，再将数据写入通道(容量10W)，`main` 方法中会开启一个协程一直读取这个通道中的内容，解析后写入文件系统

2. 删除日志文件

收到客户端调用后，将客户端指定的日志文件删掉，主要是方便客户端测试，但这个接口没有对文件做限制，有安全风险

3. 统计文件行数

收到客户端调用后，统计客户端指定文件的行数，主要是方便客户端针对写入行数做统计，判断调用是否完全成功

## 客户端

客户端定义了多个变量，如下：

```go
var (
	// rpc 连接客户端
	client = initRpc()
	// ID 通道
	ch = make(chan string, 1000)
	// WaitGroup
	wg = sync.WaitGroup{}
	// ID文件名称
	file = "id.txt"
	// 定义任务数
	idsNum = 100000
	// 处理数量
	dealNum = 0
	// 工人数
	workers = 100
	// 单个工人任务数
	size = idsNum / workers
	// 工人
	worker = func(no int) {
		fmt.Printf("[worker - %d] 工人启动\n", no)
		count := 0
		defer func() {
			fmt.Printf("[worker - %d] 工人退出，共处理任务[%d]个\n", no, count)
		}()
		for {
			id, _ := <-ch
			if id != "" {
				// 业务处理
				req := LogRequest{
					File:    file,
					Message: id + "\n",
				}
				res := LogResponse{}
				err := client.Call("Log.Record", req, &res)
				if err != nil {
					continue
				}
				// 统计
				count += 1
				dealNum += 1
				// 全部任务处理完毕
				if dealNum == idsNum {
					// 关闭 ID 通道
					close(ch)
				}
				if count == size {
					wg.Done()
					break
				}
			}
		}
	}
)
```

客户端流程：

1. WaitGroup 添加工人
2. 循环启动工人协程
3. 从客户端目录下的 `id.txt` 读取内容，并逐行写入 ID 通道
4. 主进程等待工人工作
5. 等待10秒，让服务器处理完所有数据
6. 统计服务器文件行数

## 测试方法

1. 开启服务端

   ```shell
   ➜  go-rpc-demo cd server 
   ➜  server go run server.go 
   new client in coming, network is tcp, addr is 127.0.0.1:9900
   
   ```

2. 开启客户端

   ```shell
   ➜  go-rpc-demo cd client 
   ➜  client go run client.go 
   [worker - 5] 工人启动
   [worker - 15] 工人启动
   [worker - 1] 工人启动
   [worker - 6] 工人启动
   [worker - 2] 工人启动
   [worker - 3] 工人启动
   ...
   [worker - 83] 工人启动
   [worker - 55] 工人退出，共处理任务[1000]个
   [worker - 11] 工人退出，共处理任务[1000]个
   [worker - 62] 工人退出，共处理任务[1000]个
   [worker - 91] 工人退出，共处理任务[1000]个
   [worker - 32] 工人退出，共处理任务[1000]个
   [worker - 95] 工人退出，共处理任务[1000]个
   [worker - 37] 工人退出，共处理任务[1000]个
   [main] 从文件读取ID数量：100000
   [worker - 44] 工人退出，共处理任务[1000]个
   [worker - 1] 工人退出，共处理任务[1000]个
   ...
   [worker - 20] 工人退出，共处理任务[1000]个
   [worker - 88] 工人退出，共处理任务[1000]个
   [worker - 19] 工人退出，共处理任务[1000]个
   [worker - 22] 工人退出，共处理任务[1000]个
   [main] 等待10秒，让服务器处理完所有数据
   [main] 服务端日志文件行数：100000
   ```

3. 客户端其它操作

   - 统计服务器文件行数

     ```go
     CountServerFileLine()
     ```

   - 删除服务器文件

     ```go
     RemoveServerFile()
     ```