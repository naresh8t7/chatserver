package client

import "log"

type Channel struct {
	name     string
	messages chan string
	members  map[string]*Client
}

var Channels = map[string]*Channel{}

func (c *Client) JoinChannel() {

	name, err := inputCommand(c.Connection, "Please enter Channel name: ")
	if err != nil {
		panic(err)
	}

	c.join(name)
}

//Leave current Channel
func (c *Client) LeaveChannel() {
	//only if Channel is not empty
	if c.Channel != "" {
		delete(Channels[c.Channel].members, c.Connection.RemoteAddr().String())
		log.Printf("Leave: removing user %v from Channel %v: current members: %v", c.Name, c.Channel, Channels[c.Channel].members)
		printMessage(c.Connection, "leaving "+c.Channel)
	}
}

func (c *Client) AddChannel() {
	name, err := inputCommand(c.Connection, "Please enter Channel name: ")
	if err != nil {
		panic(err)
	}
	c.createChannel(name)
}

func (c *Client) createChannel(name string) {
	//if already a member of another Channel, Leave that one first
	if _, ok := Channels[name]; ok {
		c.join(name)
		return
	}
	if name != "" {
		cr := NewChannel(name)
		go cr.listenPublish()
		cr.members[c.Connection.RemoteAddr().String()] = c

		if name != "default" {
			c.LeaveChannel()
			Channels[c.Channel].notification(c.Name + " has left..")
		}
		// set Clients Channel to new Channel
		c.Channel = cr.name
		// Add new Channel to map
		Channels[cr.name] = cr
		cr.notification(c.Name + " has joined!")

		printMessage(c.Connection, "* Channel "+cr.name+" has been Addd *")
	} else {
		printMessage(c.Connection, "* error: could not Add Channel \""+name+"\" *")
	}
}

func (c *Client) join(name string) {
	if r := Channels[name]; r != nil {
		r.members[c.Connection.RemoteAddr().String()] = c

		if c.Channel != "default" {
			c.LeaveChannel()
			r.notification(c.Name + " has left..")
		}

		c.Channel = name
		printMessage(c.Connection, c.Name+" has joined "+r.name)
		r.notification(c.Name + " has joined!")
	} else {
		printMessage(c.Connection, "error: could not join Channel")
	}
}

func NewChannel(name string) *Channel {
	return &Channel{
		name:     name,
		members:  make(map[string]*Client),
		messages: make(chan string),
	}
}

func (r *Channel) notification(msg string) {
	r.messages <- "* " + msg + " *"
}

func (r *Channel) listenPublish() {
	for {
		out := <-r.messages
		for _, v := range r.members {
			v.Msg <- out
			log.Printf("AddChannel: broadcasting msg in Channel: %v to member: %v", r.name, v.Name)
		}
	}
}
