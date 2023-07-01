package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	Status     int
}

const (
	Exit        = 0
	OpenChat    = 1
	PrivateChat = 2
	UpdateName  = 3
	Init        = 999
)

func NewClient(serverIP string, serverPort int) *Client {
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		Status:     Init,
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		fmt.Println("NewClient() in client.go, error in net.Dial(): ", err)
		return nil
	}

	client.conn = conn
	return client
}

func (thisClient *Client) menu() bool {
	var status int
	fmt.Println("1.Open Chat")
	fmt.Println("2.Private Chat")
	fmt.Println("3.update your name")

	fmt.Println("0.exit")

	_, err := fmt.Scanln(&status)
	if err != nil {
		return false
	}

	if status >= 0 && status <= 3 {
		thisClient.Status = status
		return true
	} else {
		fmt.Println(">< please make sure your input valid")
		return false
	}

}

func (thisClient *Client) DealResponse() int {
	_, err := io.Copy(os.Stdout, thisClient.conn)
	if err != nil {
		fmt.Println("Client.DealResponse io.Copy error, ", err)
		thisClient.conn.Close()
		return 1
	}
	return 0
}

func (thisClient *Client) OpenChat() {
	// 提示用户输入消息
	var chatmsg string
	fmt.Println(">>>please input your message content >>> Open")
	_, err := fmt.Scanln(&chatmsg)
	if err != nil {
		fmt.Println("Client.OpenChat ScanLn error, ", err)
		return
	}

	// 消息不为空则发送
	for chatmsg != "exit" {
		if len(chatmsg) != 0 {
			// sendmsg := chatmsg + "\n"
			chatmsg += "\n"
			_, err := thisClient.conn.Write([]byte(chatmsg))
			if err != nil {
				fmt.Println("from Client.OpenChat() in client.go, error : ", err)
				break
			}
		}

		chatmsg = ""
		fmt.Println(">>>message've sent success")
		fmt.Println(">>>please input your message content >>> Open")
		_, err := fmt.Scanln(&chatmsg)
		if err != nil {
			fmt.Println("Client.OpenChat ScanLn error, ", err)
			return
		}

	}
}

func (thisClient *Client) SelectUser() {
	sendmsg := "who\n"
	_, err := thisClient.conn.Write([]byte(sendmsg))
	if err != nil {
		fmt.Println("from Client.SelectUser() in client.go, error : ", err)
		return
	}
}

func (thisClient *Client) PrivateChat() {
	var remoteName string
	var chatmsg string

	thisClient.SelectUser()
	fmt.Println("please input who you want to chat with :")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>please input your message content >>> Private with", remoteName)
		fmt.Scanln(&chatmsg)

		for chatmsg != "exit" {
			if len(chatmsg) != 0 {
				sendmsg := "to|" + remoteName + "|" + chatmsg + "\n\n"

				_, err := thisClient.conn.Write([]byte(sendmsg))
				if err != nil {
					fmt.Println("from Client.OpenChat() in client.go, error : ", err)
					break
				}

			}

			chatmsg = ""
			fmt.Println(">>>message've sent success")
			fmt.Println(">>>please input your message content >>> Private")
			fmt.Scanln(&chatmsg)
		}

		thisClient.SelectUser()
		fmt.Println("please input who you want to chat with :")
		fmt.Scanln(&remoteName)

	}

}

func (thisClient *Client) UpdateName() bool {
	fmt.Println(">>>please input your new name")

	fmt.Scanln(&thisClient.Name)

	updnmsg := "rename|" + thisClient.Name + "\n"

	_, err := thisClient.conn.Write([]byte(updnmsg))
	if err != nil {
		fmt.Println("from Client.UpdateName() in client.go, error : ", err)
		return false
	}

	return true

}

func (thisClient *Client) Run() {
	for thisClient.Status != Exit {
		for thisClient.menu() != true {

		}
		// 根据flag不同，处理不同业务
		switch thisClient.Status {
		case OpenChat:
			// Open Chat
			fmt.Println(">>> Open Chat mode")
			thisClient.OpenChat()
			break
		case PrivateChat:
			// Private Chat
			fmt.Println(">>> Private Chat mode")
			thisClient.PrivateChat()
			break
		case UpdateName:
			// update name
			fmt.Println(">>> update name")
			thisClient.UpdateName()
			break
		}
	}
}

func main() {
	// 命令行解析
	flag.Parse()
	serverIP, serverPort := "127.0.0.1", 8888
	client := NewClient(serverIP, serverPort)
	if client == nil {
		fmt.Println(">>>>>client connect server failed, please retry")
		return
	}

	fmt.Println(">>>>>client connect server success")

	// 启动客户端业务
	go client.Run()
	for {
		if client.DealResponse() == 1 {
			return
		}
	}

	/*go client.DealResponse()
	client.Run()*/
}
