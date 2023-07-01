package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	IP   string
	// MessageChannel chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userIP := conn.RemoteAddr().String()

	user := &User{
		Name: userIP,
		IP:   userIP,
		// MessageChannel: make(chan string),
		conn:   conn,
		server: server,
	}

	// go user.HandleMessageFromServer()
	return user
}

// **** MessageChannel 只接收广播消息，弃用
// HandleMessageFromServer 处理其他用户新发来的消息，将消息通过Conn发送给客户端
/*func (thisUser *User) HandleMessageFromServer() {
	for {
		msg := <- thisUser.MessageChannel
		_, err := thisUser.conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("user.HandleMessage Write error: ", err)
			// return
		}

	}
}*/

func (thisUser *User) Online() {
	thisUser.server.Lock.Lock()
	thisUser.server.OnlineUsers[thisUser.Name] = thisUser
	thisUser.server.Lock.Unlock()

	thisUser.server.Broadcast(thisUser, "Login")
}

func (thisUser *User) Offline() {
	thisUser.server.Lock.Lock()
	delete(thisUser.server.OnlineUsers, thisUser.Name)
	thisUser.server.Lock.Unlock()

	thisUser.server.Broadcast(thisUser, "Log out")
}

func (thisUser *User) SendMessage(msg string) {
	_, err := thisUser.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("User.SendMessage write error: ", err)
		return
	}
}

func (thisUser *User) HandleMessage(msg string) {
	if msg == "who" {
		thisUser.server.Lock.Lock()

		for _, user := range thisUser.server.OnlineUsers {
			whoOnline := "[" + user.IP + "]" + user.Name + ":" + "Online\n"
			thisUser.SendMessage(whoOnline)
		}

		thisUser.server.Lock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]

		_, hasBennUsed := thisUser.server.OnlineUsers[newName]
		if hasBennUsed {
			thisUser.SendMessage("this name has been used...\n")
		} else {
			thisUser.server.Lock.Lock()

			delete(thisUser.server.OnlineUsers, thisUser.Name)
			thisUser.server.OnlineUsers[newName] = thisUser

			thisUser.server.Lock.Unlock()
			thisUser.Name = newName
			thisUser.SendMessage("your name update success : " + thisUser.Name + "\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式：to|张三|hello
		// 获取用户名

		splitMessage := strings.Split(msg, "|")
		remoteName := splitMessage[1]
		if remoteName == "" {
			thisUser.SendMessage("message format is wrong(remote name is empty), please resend the message\n")
		}

		// 根据用户名得到User对象
		remoteUser, ok := thisUser.server.OnlineUsers[remoteName]
		if !ok {
			thisUser.SendMessage("the user is not exist or offline\n")
			return
		}

		// 获取信息
		content := splitMessage[2]
		if content == "" {
			thisUser.SendMessage("your message content is empty\n")
			return

		} else {
			remoteUser.SendMessage(thisUser.Name + "send to you : " + content + "\n")

		}
	} else {
		thisUser.server.Broadcast(thisUser, msg)
	}
}
