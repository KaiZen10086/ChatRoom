package main

import (
	"./chatroom2"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

//先建一个默认房间
var room = chatroom2.NewChatRoom()
var defaultRoom = chatroom2.RoomInfo{Name: "default", CreateId: "0.0.0.0"}

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	room.AddRoom(defaultRoom)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}

type MessageStruct struct {
	User *chatroom2.ClientInfo //消息的发送者
	Msg  string
}

var message = make(chan *MessageStruct)

func broadcaster() {
	for {
		select {
		case msgStruct := <-message:
			msg := msgStruct.Msg
			currentRoom := msgStruct.User.CurrentRoom
			//获得当前消息发送者所在房间的所有用户
			users := room.RoomUsers(currentRoom)
			for _, user := range users {
				user.ClientResource <- msg
			}
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)

	go writeToClient(conn, ch)

	who := conn.RemoteAddr().String()
	cli := &chatroom2.ClientInfo{}
	cli.Id = who

	//当客户端连接过来时，给客户端一条消息
	//注意，这时的ch会立马被writeToCLient goroutine读取，并发送到当前客户端
	//所以已连接的其他客户端不会接受到该条消息
	ch <- "your nickname:"
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	nickname := strings.TrimSpace(line)
	cli.NickName = nickname
	cli.CurrentRoom = defaultRoom //进入默认的房间
	cli.ClientResource = ch

	msgStruct := &MessageStruct{
		User: cli,
	}

	msgStruct.Msg = nickname + " are arrived"
	message <- msgStruct
	room.AddClient(defaultRoom, cli)

	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg := input.Text()
		msg = strings.TrimSpace(msg)
		paras := strings.Split(msg, " ")
		switch paras[0] {
		case "list":
			r := room.RoomList()
			var roomList string
			for _, room := range r {
				roomList += room.Name + " " + room.CreateId + "\t"
			}
			ch <- roomList
		case "enter":
			roomName, ok := room.FindRoom(paras[1])
			if ok {
				currentRoom := cli.CurrentRoom
				room.RemoveClient(currentRoom, cli)
				cli.CurrentRoom = roomName
				room.AddClient(roomName, cli)
				ch <- cli.NickName + " enter " + cli.CurrentRoom.Name
			}
		case "exit":
			//如果不在默认房间，则进入默认房间
			//如果已经在默认房间了，则跳出循环
			if cli.CurrentRoom == defaultRoom {
				fmt.Println(111)
				room.RemoveClient(defaultRoom, cli)
				goto exit //跳转到链接关闭
			}
			room.RemoveClient(cli.CurrentRoom, cli)
			cli.CurrentRoom = defaultRoom
			room.AddClient(defaultRoom, cli)
			ch <- cli.NickName + " enter " + cli.CurrentRoom.Name
		case "creat":
			newRoom := chatroom2.RoomInfo{Name: paras[1], CreateId: who}
			fmt.Println(newRoom.Name, newRoom.CreateId)
			room.AddRoom(newRoom)
		case "delete":
			r := room.RoomList()
			for _, deleteRoom := range r {
				if deleteRoom.Name == paras[1] && deleteRoom.CreateId == cli.Id {
					users := room.RoomUsers(deleteRoom)
					for _, user := range users {
						room.RemoveClient(deleteRoom, user)
						room.AddClient(defaultRoom, user)
						user.CurrentRoom = defaultRoom
						user.ClientResource <- user.NickName + " enter " + "defaultRoom"
					}
					room.RemoveRoom(deleteRoom)
					break
				}
			}
		default:
			msgStruct.Msg = nickname + ":" + msg
			message <- msgStruct
		}
	}
exit:
	msgStruct.Msg = nickname + " are left"
	message <- msgStruct
	conn.Close()
	close(cli.ClientResource)
}

func writeToClient(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
