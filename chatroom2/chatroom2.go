package chatroom2

import "fmt"

//房间信息
type RoomInfo struct {
	Name     string //房间名
	CreateId string //创建者id，用于删除房间
}

//用户信息
type ClientInfo struct {
	Id             string   //用户ID
	NickName       string   //用户昵称
	CurrentRoom    RoomInfo //当前房间
	ClientResource chan<- string
}

//一个房间对应多个人
type Clients []*ClientInfo
type ChatRoom struct {
	RoomMap map[RoomInfo]Clients
}

//新建房间
func NewChatRoom() *ChatRoom {
	return &ChatRoom{RoomMap: make(map[RoomInfo]Clients)}
}

//增加房间
//是一个方法
func (c *ChatRoom) AddRoom(r RoomInfo) {
	_, ok := c.RoomMap[r]
	if !ok {
		c.RoomMap[r] = Clients{}
		fmt.Println("ok")
	}
}

//删除房间
//在外面做房间创建者的判断
func (c *ChatRoom) RemoveRoom(r RoomInfo) {
	delete(c.RoomMap, r)
}

//往房间中增加用户
func (c *ChatRoom) AddClient(r RoomInfo, cli *ClientInfo) {
	flag := false
	//避免重复添加
	for _, client := range c.RoomMap[r] {
		if client.Id == cli.Id {
			flag = true
		}
	}
	if flag == false {
		cli.CurrentRoom = r
		c.RoomMap[r] = append(c.RoomMap[r], cli)
	}
}

//从房间中删除用户
func (c *ChatRoom) RemoveClient(r RoomInfo, cli *ClientInfo) {
	_, ok := c.RoomMap[r]
	//房间不存在
	if !ok {
		return
	}
	//如果是创建者被删除，则直接删除房间
	if cli.Id == r.CreateId {
		c.RemoveRoom(r)
		return
	}

	for idx, client := range c.RoomMap[r] {
		if client.Id == cli.Id {
			c.RoomMap[r] = append(c.RoomMap[r][:idx], c.RoomMap[r][idx+1:]...) //要加三个点
			//fmt.Println("remove client ",client)
			//fmt.Println(len(c.RoomMap[r]))
			return
		}
	}
}

//返回房间用户
func (c *ChatRoom) RoomUsers(r RoomInfo) Clients {
	return c.RoomMap[r]
}

//返回房间列表
func (c *ChatRoom) RoomList() (results []RoomInfo) {
	for k := range c.RoomMap {
		results = append(results, k)
	}
	return
}

//根据房间名找到房间
func (c *ChatRoom) FindRoom(RoomName string) (RoomInfo, bool) {
	for k := range c.RoomMap {
		if k.Name == RoomName {
			return k, true
		}
	}
	return RoomInfo{}, false
}

