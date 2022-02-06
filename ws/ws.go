package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ClientManager is a websocket manager
type ClientManager struct {
	Clients    map[string]*Client
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

// Client is a websocket client
type Client struct {
	ID       string
	Socket   *websocket.Conn
	Send     chan []byte
	Username string
	RoomId   int
}

// Message is return msg
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

// Manager define a ws server manager
var Manager = ClientManager{
	Clients:    make(map[string]*Client),
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
}

// Rooms 房间
var Rooms = make(map[int][]*Client)

var mutex = sync.Mutex{}

func (manager *ClientManager) Start() {
	for {
		select {
		case conn := <-Manager.Register:
			log.Printf("新用户加入:%v", conn.ID)
			//用户进入房间
			Rooms[conn.RoomId] = append(Rooms[conn.RoomId], conn)
		case conn := <-Manager.Unregister:
			log.Printf("用户离开:%v", conn.ID)
			temUser := Rooms[conn.RoomId]
			for k, v := range temUser {
				if v.Username == conn.Username {
					//索引排除法
					mutex.Lock()
					Rooms[conn.RoomId] = append(temUser[:k], temUser[k+1:]...)
					mutex.Unlock()
					close(conn.Send)
				}
			}
		case message := <-Manager.Broadcast:
			MessageStruct := Message{}
			json.Unmarshal(message, &MessageStruct)
			rommId, _ := strconv.Atoi(MessageStruct.Recipient)
			for _, conn := range Rooms[rommId] {
				if conn.RoomId != rommId {
					continue
				}
				select {
				case conn.Send <- message:
				default:
					temUser := Rooms[conn.RoomId]
					for k, v := range temUser {
						if v.Username == conn.Username {
							//索引排除法
							Rooms[conn.RoomId] = append(temUser[:k], temUser[k+1:]...)
							close(conn.Send)
						}
					}
				}
			}
		}
	}
}

func creatId(uid, touid string) string {
	return uid + "_" + touid
}

func (c *Client) Read() {
	defer func() {
		Manager.Unregister <- c
		c.Socket.Close()
	}()
	for {
		fmt.Println("进入房间")
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			Manager.Unregister <- c
			c.Socket.Close()
			break
		}
		log.Printf("读取到客户端的信息:%s", string(message))
		Manager.Broadcast <- message
	}
}

func (c *Client) Write() {
	defer func() {
		c.Socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			MessageStruct := Message{}
			json.Unmarshal(message, &MessageStruct)
			log.Printf("发送到到客户端的信息:%s", string(message))
			sendMsg := MessageStruct.Sender + "：" + MessageStruct.Content
			c.Socket.WriteMessage(websocket.TextMessage, []byte(sendMsg))
		}
	}
}

func WsHandler(c *gin.Context) {
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		http.NotFound(c.Writer, c.Request)
		return
	}
	uid := c.Query("uid")
	touid := c.Query("to_uid")
	roomId, _ := strconv.Atoi(touid)
	client := &Client{
		ID:       creatId(uid, touid),
		Socket:   conn,
		Send:     make(chan []byte),
		Username: uid,
		RoomId:   roomId,
	}
	//注册信息
	Manager.Register <- client
	go client.Read()
	go client.Write()
}

func GetOnlineRoomUserCount(roomId int) int {
	return len(Rooms[roomId])
}
