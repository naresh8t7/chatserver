// Package min provides application bootstrap
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"chatserver/internal/client"
	"chatserver/internal/config"
	"chatserver/internal/httpserver"

	"github.com/go-redis/redis"
	"gopkg.in/natefinch/lumberjack.v2"
)

var clients = map[string]*client.Client{}

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf(" err %v", err)
	}
	if config.Address == "" || config.TcpPort == "" || config.HttpPort == "" || config.LogFile == "" || config.RedisPort == "" {
		log.Fatal("inavlid config")
	}
	// lumberjack config for rotating log file based on file size , other  params can also set in config.
	log.SetOutput(&lumberjack.Logger{
		Filename:   config.LogFile,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	})
	// redis for storing messages by client. Used for rest api endpoints
	redisURL := fmt.Sprintf("redis://%s:%s", config.Address, config.RedisPort)
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("unable to set up redis conn %v", err)
	}
	rdc := redis.NewClient(opts)
	tcpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.Address, config.TcpPort))
	if err != nil {
		log.Fatalf("unable to set up tcp listener %v", err)
	}
	defer tcpListener.Close()
	log.Printf("\n Listening On %v", tcpListener.Addr())

	mu := sync.Mutex{}

	server := httpserver.NewHttpServer(rdc, clients, &mu)
	server.AddRoutes()
	go func() {
		log.Println("starting HTTP sever")
		http.ListenAndServe(fmt.Sprintf(":%s", config.HttpPort), server.Router)
	}()

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			log.Fatalf("unable to accept connection %v", err)
		}
		client := client.NewClient(conn, rdc)
		mu.Lock()
		clients[client.Name] = client
		mu.Unlock()

		go client.SendMessages()
		go client.RecieveMessages()

	}
}
