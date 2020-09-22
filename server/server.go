package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strings"
)

type Log struct {
}

type LogRequest struct {
	File    string
	Message string
}

type LogResponse struct {
	Message string
	Line    int
}

// ID 通道
var ch = make(chan string, 100000)

// 记录日志
func (log *Log) Record(req LogRequest, res *LogResponse) error {
	// 先记录数据到通道
	ch <- req.File + ":" + req.Message
	res.Message = req.Message
	return nil
}

// 删除日志文件
func (log *Log) Remove(req LogRequest, res *LogResponse) error {
	err := os.Remove(req.File)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// 统计行数
func (log *Log) CountLine(req LogRequest, res *LogResponse) error {
	f, err := os.OpenFile(req.File, os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Println(err)
		return err
	}
	line := 0
	buf := bufio.NewReader(f)
	for {
		b, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("------" + err.Error())
			continue
		}
		idStr := strings.Replace(string(b), "\n", "", -1)
		if idStr == "" {
			fmt.Println("=====" + idStr + "********")
		}
		line += 1
	}
	res.Line = line
	return nil
}

func main() {
	_ = rpc.Register(new(Log))
	lis, err := net.Listen("tcp", "0.0.0.0:9900")
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		// 从通道读取内容写入文件
		for str := range ch {
			recordFile(str)
		}
	}()

	for {
		con, err := lis.Accept()
		if err != nil {
			continue
		}
		go func(con net.Conn) {
			str := fmt.Sprintf("new client in coming, network is %s, addr is %s\n", con.LocalAddr().Network(), con.LocalAddr().String())
			_, _ = fmt.Fprintf(os.Stdout, "%s", str)
			jsonrpc.ServeConn(con)
		}(con)
	}
}

// 记录日志文件
func recordFile(str string) {
	arr := strings.Split(str, ":")
	f, err := os.OpenFile(arr[0], os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Println(err)
	} else {
		_, err = f.Write([]byte(arr[1]))
		if err != nil {
			fmt.Println(err)
		}
	}
	err = f.Close()
	if err != nil {
		fmt.Println(err)
	}
}


