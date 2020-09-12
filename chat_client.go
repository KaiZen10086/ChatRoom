package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	//获取服务端消息 you are xxx
	go ioCopy(os.Stdout, conn)

	//将用户输入的文本消息发送到到服务端
	ioCopy(conn, os.Stdin)
}

func ioCopy(dst io.Writer, src io.Reader) {
	//只要没出错就一直发送
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
		return
	}
}
