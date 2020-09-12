/*package main

import (
	"fmt"
	"net"
	"os"
)

var clients []net.Conn

func main() {
	var (
		host   = "localhost"
		port   = "8000"
		remote = host + ":" + port
		data   = make([]byte, 1024)
	)
	fmt.Println("Initiating server...")

	lis, err := net.Listen("tcp", remote)
	defer lis.Close()

	if err != nil {
		fmt.Printf("Error when listen: %s, Err: %s\n", remote, err)
		os.Exit(-1)
	}

	for {
		var res string
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("Error accepting client: ", err.Error())
			os.Exit(0)
		}
		clients = append(clients, conn)

		go func(con net.Conn) {
			fmt.Println("New connection: ", con.RemoteAddr())

			// Get client's name
			length, err := con.Read(data)
			if err != nil {
				fmt.Printf("Client %v quit.\n", con.RemoteAddr())
				con.Close()
				disconnect(con, con.RemoteAddr().String())
				return
			}
			name := string(data[:length])
			comeStr := name + " entered the room."
			notify(con, comeStr)

			// Begin recieve message from client
			for {
				length, err := con.Read(data)
				if err != nil {
					fmt.Printf("Client %s quit.\n", name)
					con.Close()
					disconnect(con, name)
					return
				}
				res = string(data[:length])
				sprdMsg := name + " said: " + res
				fmt.Println(sprdMsg)
				res = "You said:" + res
				con.Write([]byte(res))
				notify(con, sprdMsg)
			}
		}(conn)
	}
}

func notify(conn net.Conn, msg string) {
	for _, con := range clients {
		if con.RemoteAddr() != conn.RemoteAddr() {
			con.Write([]byte(msg))
		}
	}
}

func disconnect(conn net.Conn, name string) {
	for index, con := range clients {
		if con.RemoteAddr() == conn.RemoteAddr() {
			disMsg := name + " has left the room."
			fmt.Println(disMsg)
			clients = append(clients[:index], clients[index+1:]...)
			notify(conn, disMsg)
		}
	}
}
*/

/*package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Client struct {
	C    chan string // pipeline to send data
	Name string
	Addr string
}

//save online users.
var onlineMap map[string]Client

var message = make(chan string)

func WriteMsgToClient(cli Client, conn net.Conn) {
	for msg := range cli.C {
		_, _ = conn.Write([]byte(msg + "\n"))
	}
}

func MakeMsg(cli Client, msg string) (buf string) {
	buf = "[" + cli.Addr + "]" + cli.Name + ": " + msg
	return
}

func HandleConn(conn net.Conn) {
	defer conn.Close()
	cliAddr := conn.RemoteAddr().String()
	cli := Client{make(chan string), cliAddr, cliAddr}
	onlineMap[cliAddr] = cli

	// Start a new association to send information to the current client
	go WriteMsgToClient(cli, conn)
	message <- MakeMsg(cli, "login")

	cli.C <- MakeMsg(cli, "i am here")

	// Record whether the other party voluntarily exits
	hasQuit := make(chan bool)
	// Whether the data of the opposite party is sent
	hasData := make(chan bool)
	// Open a new association to accept the data sent by users
	go func() {
		buf := make([]byte, 2048)
		for {
			n, err := conn.Read(buf)
			fmt.Println(n)
			if n == 0 {
				fmt.Println("此时n为0")
				hasQuit <- true
				fmt.Println("conn.Read err= ", err)
				return
			}
			msg := string(buf[:n])
			if msg == "who" {
				_, _ = conn.Write([]byte("user list:\n"))
				for _, tmp := range onlineMap {
					userAddr := tmp.Addr + ":" + tmp.Name + "\n"
					_, _ = conn.Write([]byte(userAddr))

				}
			} else if len(msg) >= 8 && msg[:6] == "rename" {
				//rename
				newName := strings.Split(msg, "|")[1]
				cli.Name = newName
				onlineMap[cliAddr] = cli
				_, _ = conn.Write([]byte("rename ok\n"))

			} else {
				message <- MakeMsg(cli, msg)
			}
			hasData <- true // Representative has data
		}

	}()

	// Disallow disconnection of server clients
	for {
		select {
		case <-hasQuit:
			delete(onlineMap, cliAddr)
			message <- MakeMsg(cli, "login out")
		case <-hasData:
		case <-time.After(30 * time.Second):
			delete(onlineMap, cliAddr)
			message <- MakeMsg(cli, "time out leave out")
			return
		}
	}
}

func Manager() {
	onlineMap = make(map[string]Client)
	for {
		msg := <-message
		for _, cli := range onlineMap {
			cli.C <- msg
		}
	}
}

func main() {
	//listen
	lis, err := net.Listen("tcp", ":8888")
	if err != nil {
		fmt.Println("net.Listen err =", err)
		return
	}
	defer lis.Close()
	// new association,As soon as there is a message is coming,it will traverse the map and send the message.
	go Manager()
	// Master association, waiting for user link
	for {
		conn, err1 := lis.Accept()
		if err1 != nil {
			fmt.Println("is.Accept err1=", err1)
			continue
		}
		// Handle user links
		go HandleConn(conn)
	}

}

*/

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
		go handleconn(conn)
	}
}

//type client chan<- string

type MessageStruct struct {
	User *chatroom2.ClientInfo //消息的发送者信息
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

func handleconn(conn net.Conn) {
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
			//TODO list all room
			r := room.RoomList()
			var roomList string
			for _, room := range r {
				roomList += room.Name + " " + room.CreateId + "\t"
			}
			ch <- roomList
		case "enter":
			//TODO enter a room
			roomName, ok := room.FindRoom(paras[1])
			if ok {
				currentRoom := cli.CurrentRoom
				room.RemoveClient(currentRoom, cli)
				cli.CurrentRoom = roomName
				room.AddClient(roomName, cli)
				ch <- cli.NickName + " enter " + cli.CurrentRoom.Name
			}
		case "exit":
			//TODO exit current room
			//如果不在默认房间，则进入默认房间
			//如果已经在默认房间了，则跳出循环
			if cli.CurrentRoom == defaultRoom {
				fmt.Println(111)
				room.RemoveClient(defaultRoom, cli)
				goto exit
			}
			room.RemoveClient(cli.CurrentRoom, cli)
			cli.CurrentRoom = defaultRoom
			room.AddClient(defaultRoom, cli)
			ch <- cli.NickName + " enter " + cli.CurrentRoom.Name
		case "creat":
			//TODO create a room
			newRoom := chatroom2.RoomInfo{Name: paras[1], CreateId: who}
			fmt.Println(newRoom.Name, newRoom.CreateId)
			room.AddRoom(newRoom)
		case "delete":
			//TODO delete a room
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

