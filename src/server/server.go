package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Server struct {
	IP               string
	Port             int
	OnlineUsers      map[string]*User
	Lock             sync.Mutex
	BroadcastChannel chan string
}

// NewServer 创建一个Server
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:               ip,
		Port:             port,
		OnlineUsers:      make(map[string]*User),
		BroadcastChannel: make(chan string),
	}

	return server
}

func (thisServer *Server) Broadcast(user *User, msg string) {
	broadMsg := "[" + user.IP + "]" + user.Name + ":" + msg + "\n"
	thisServer.BroadcastChannel <- broadMsg
}

func (thisServer *Server) HandleBroadcast() {
	for {
		msg := <-thisServer.BroadcastChannel
		thisServer.Lock.Lock()

		for _, user := range thisServer.OnlineUsers {
			// user.MessageChannel <- msg
			user.SendMessage(msg)
		}

		thisServer.Lock.Unlock()
	}
}

// HandleUser 处理一个user的业务
func (thisServer *Server) HandleUser(conn net.Conn) {
	user := NewUser(conn, thisServer)

	user.Online()
	defer user.Offline()

	isAlive := make(chan bool)
	buffer := make([]byte, 4096)
	for {
		// 用户保活检测
		select {
		case <-isAlive:
			break
			// 不做任何事，但是刷新time，表明用户连接活动
		case <-time.After(3600 * time.Second):
			user.SendMessage("you are over_time and kicked")
			// close(user.MessageChannel)
			err := conn.Close()
			if err != nil {
				fmt.Println("Server.HandleUser conn.Close error: ", err)
				return
			}
			break
		}

		// 用户业务处理
		sz, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Server.HandleUser conn.Read error: ", err)
		} else if sz == 0 {
			user.Offline()
			return
		} else {
			msg := string(buffer[:sz-1])
			user.HandleMessage(msg)
			isAlive <- true
		}
	}
}

func (thisServer *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", thisServer.IP, thisServer.Port))
	if err != nil {
		fmt.Println("Server.Start net.Listen error: ", err)
		return
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Server.Start, net.Close error : ", err)
			return
		}
	}(listener)

	go thisServer.HandleBroadcast()
	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Server.Start, listener.Accept error: ", err)
			continue
		}
		fmt.Printf("a new connection from %v\n", conn.RemoteAddr().String())

		// do handler
		go thisServer.HandleUser(conn)
	}
}

func main() {
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}
