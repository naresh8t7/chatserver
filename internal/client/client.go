package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

var (
	commands = map[string]string{
		"\\quit":     "quit\n",
		"\\channels": "list all channels and online users\n",
		"\\add":      "add a new Channel\n",
		"\\join":     "join a Channel\n",
		"\\help":     "prints all available commands\n",
	}
)

const (
	timeFormat = "02/01/2006 15:04:05"
)

type Client struct {
	Name        string
	Msg         chan string
	Connection  net.Conn
	Channel     string
	RedisClient *redis.Client
}

type ChatClient interface {
	SendMessages()
	RecieveMessages()
	Command(msg string) bool
	Close()
	JoinChannel()
	LeaveChannel()
	AddChannel()
}

func NewClient(conn net.Conn, rdc *redis.Client) *Client {
	name, err := inputCommand(conn, "Please Enter Name: ")
	if err != nil {
		log.Fatal(err)
	}

	printMessage(conn, "Welcome "+name)

	//init Client struct
	client := &Client{
		Msg:         make(chan string),
		Connection:  conn,
		Name:        name,
		Channel:     "default",
		RedisClient: rdc,
	}
	client.createChannel(client.Channel)

	//print commands
	printMessage(conn, commands)
	return client
}

func (c *Client) Close() {
	c.LeaveChannel()
	c.Connection.Close()
	c.Msg <- "\\quit"
}

func (c *Client) RecieveMessages() {
	for {
		msg := <-c.Msg
		if msg == "\\quit" {
			Channels[c.Channel].notification(c.Name + " has left..")
			break
		}
		printMessage(c.Connection, msg)
	}
}

func (c *Client) Send(msg, name string) {
	if msg == "\\quit" {
		c.Close()
		log.Printf("%v has left..", name)
	}

	if c.Command(msg) {
		cmsg := fmt.Sprintf("%s(User:%s): \"%s\"", time.Now().Format(timeFormat), name, msg)
		log.Println(cmsg)

		for _, v := range Channels {
			for k := range v.members {
				if k == c.Connection.RemoteAddr().String() {
					c.RedisClient.RPush(name, fmt.Sprintf("%s : %s", time.Now().Format(timeFormat), msg))
					v.messages <- cmsg
				}
			}
		}
		time.Sleep(5 * time.Millisecond)

	}
}

func (c *Client) SendMessages() {
	for {
		msg, err := inputCommand(c.Connection, "Message>")
		if err != nil {
			log.Fatal(err)
		}
		c.Send(msg, c.Name)

	}
}

func (c *Client) Command(msg string) bool {
	switch {
	case msg == "\\channels":
		c.Connection.Write([]byte("-------------------\n"))
		for k := range Channels {
			count := 0
			for range Channels[k].members {
				count++
			}
			c.Connection.Write([]byte(k + " : online members(" + strconv.Itoa(count) + ")\n"))
		}
		c.Connection.Write([]byte("-------------------\n"))
		return false
	case msg == "\\join":
		c.JoinChannel()
		return false
	case msg == "\\help":
		printMessage(c.Connection, commands)
		return false
	case msg == "\\add":
		c.AddChannel()
		return false
	}
	return true
}

func inputCommand(conn net.Conn, cmd string) (string, error) {
	conn.Write([]byte(cmd))
	var s string
	var err error
	if s, err = bufio.NewReader(conn).ReadString('\n'); err != nil {
		return "", fmt.Errorf("inputCommand: could not read input from stdin: %v from Client %v", err, conn.RemoteAddr().String())
	}
	return strings.Trim(s, "\r\n"), nil
}

func printMessage(conn net.Conn, msg interface{}) error {
	if _, err := conn.Write([]byte("---------------------------\n")); err != nil {
		return err
	}
	t := reflect.ValueOf(msg)
	switch t.Kind() {
	case reflect.Map:
		for k, v := range msg.(map[string]string) {
			if _, err := conn.Write([]byte(k + " : " + v)); err != nil {
				return err
			}
		}
	case reflect.String:
		v := reflect.ValueOf(msg).String()
		if _, err := conn.Write([]byte(v + "\n")); err != nil {
			return err
		}
	}
	conn.Write([]byte("---------------------------\n"))

	return nil
}
