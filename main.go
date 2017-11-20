package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/golang/glog"

	"github.com/googollee/go-socket.io"
)

type SocketioClient struct {
	conn         socketio.Socket
	id           string
	responses    chan string
	responsesMut sync.RWMutex
}

func (c *SocketioClient) ToString() string {
	return c.id
}

func NewSocketioClient(conn socketio.Socket) (*SocketioClient, error) {
	c := &SocketioClient{conn: conn, id: conn.Id(), responses: make(chan string)}
	conn.On("message", func(msg string) {
		// echo service
		glog.Infof("message %#v received, echo back ...", msg)
		c.Send(fmt.Sprintf("node#%d: %s", os.Getpid(), msg))
	})

	go c.Run()

	return c, nil
}

func (c *SocketioClient) Run() {
	// Send response one by one.
	for response := range c.responses {
		if response == "" {
			break
		}
		if err := c.conn.Emit("message", response); err != nil {
			glog.Errorf("socketio client %#v send response %#v error: %s", c.ToString(), response, err)
			break
		}
		glog.Infof("response to client %#v: %s", c.ToString(), response)
	}

	c.responsesMut.Lock()
	close(c.responses)
	c.responses = nil
	c.responsesMut.Unlock()

	c.conn.Disconnect()

	glog.Infof("socketio client %#v closed", c.ToString())
}

func (c *SocketioClient) Close() {
	c.responsesMut.RLock()
	defer c.responsesMut.RUnlock()
	if c.responses != nil {
		c.responses <- ""
		glog.Warningf("socketio client %#v closing", c.ToString())
	}
}

func (c *SocketioClient) Send(res string) error {
	c.responsesMut.RLock()
	defer c.responsesMut.RUnlock()
	if c.responses != nil {
		c.responses <- res
		return nil
	}
	return fmt.Errorf("socketio client %#v already closed", c.ToString())
}

type SocketioServer struct {
	server     *socketio.Server
	clients    map[string]*SocketioClient
	clientsMut sync.RWMutex
}

func NewSocketioServer() (*SocketioServer, error) {
	s := &SocketioServer{clients: map[string]*SocketioClient{}}
	var err error
	s.server, err = socketio.NewServer(nil)
	if err != nil {
		return nil, err
	}
	s.server.On("connection", func(conn socketio.Socket) error {
		client, err := NewSocketioClient(conn)
		if err != nil {
			return err
		}
		s.AddClient(client)
		conn.On("disconnection", func() {
			s.DelClient(client)
			client.Close()
		})
		client.Send(fmt.Sprintf("node#%d: welcome", os.Getpid()))
		return nil
	})
	s.server.On("error", func(conn socketio.Socket, err error) {
		glog.Errorf("socketio client %#v error: %#v", conn.Id(), err)
	})

	http.Handle("/socket.io/", s.server)
	http.Handle("/", http.FileServer(http.Dir("./asset"))) // For client demo
	return s, nil
}

func (s *SocketioServer) AddClient(client *SocketioClient) {
	s.DelClient(client)

	s.clientsMut.Lock()
	defer s.clientsMut.Unlock()
	s.clients[client.id] = client
	glog.Infof("socketio client %#v added, %d clients online", client.ToString(), len(s.clients))
}

func (s *SocketioServer) DelClient(client *SocketioClient) {
	s.clientsMut.Lock()
	defer s.clientsMut.Unlock()
	_, ok := s.clients[client.id]
	delete(s.clients, client.id)
	if ok {
		glog.Infof("socketio client %#v deleted, %d clients online", client.ToString(), len(s.clients))
	}
}

func (s *SocketioServer) Run() error {
	glog.Infof("socketio server running at localhost:8001 ...")
	return http.ListenAndServe(":8001", nil)
}

func main() {
	flag.Parse()

	glog.Infof("socketio server init ...")
	server, err := NewSocketioServer()
	if err != nil {
		glog.Fatalf("socketio server init error: %s", err)
	}

	err = server.Run()
	glog.Fatalf("server exited: %s", err)
}
