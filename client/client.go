package main

import (
	"bufio"
	"fmt"
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strings"
	"sync"
	"time"
)

type LogRequest struct {
	File    string
	Message string
}

type LogResponse struct {
	Message string
	Line    int
}

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

func main() {

	// 删除服务器文件
	//RemoveServerFile()
	//return

	// WaitGroup 添加工人
	wg.Add(workers)

	// 协程启动工人
	for i := 1; i <= workers; i++ {
		go worker(i)
	}

	// 启动读取文件并写通道
	readId()

	// 等待工人工作
	wg.Wait()

	// 由于服务端是通道依次写入文件，这里等待2秒让服务器写完
	fmt.Println("[main] 等待10秒，让服务器处理完所有数据")
	time.Sleep(time.Second * 10)

	// 统计服务器文件行数
	CountServerFileLine()

}

// 初始化 Rpc 连接
func initRpc() *rpc.Client {
	client, err := jsonrpc.Dial("tcp", "0.0.0.0:9900")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return client
}

// 读取ID
func readId() {
	// 打开文件
	file, err := os.Open(file)
	if err != nil {
		fmt.Println("打开文件异常：" + err.Error())
		return
	}
	// 创建buf
	buf := bufio.NewReader(file)
	// 从文件读取ID数量
	readNum := 0
	// 循环读取数据，以换行分割
	for {
		b, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		idStr := strings.Replace(string(b), "\n", "", -1)
		ch <- idStr
		readNum += 1
	}
	fmt.Printf("[main] 从文件读取ID数量：%d\n", readNum)
}

// 删除服务端测试日志文件
func RemoveServerFile() {
	// 业务处理
	req := LogRequest{
		File: file,
	}
	res := LogResponse{}
	_ = client.Call("Log.Remove", req, &res)
}

// 统计服务端测试日志文件行数
func CountServerFileLine() {
	// 业务处理
	req := LogRequest{
		File: file,
	}
	res := LogResponse{}
	_ = client.Call("Log.CountLine", req, &res)
	fmt.Printf("[main] 服务端日志文件行数：%d\n", res.Line)
}
