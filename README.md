# Chat Server App

- [Instructions](#instruction-to-run)
- [Features](#features)
- [TODO](#todo)
- [Assumptions](#design-assumptions)
- [Libraries](#libraries)
- [References](#references)


# Instruction to run and test

## To run application
The app can be run in following ways
```
 cd chatserver
 docker-compose up // for starting a redis server
 cd cmd/chatapp
 go run main.go
 telnet 127.0.0.1 9001 // for connecting to chat server and sending messages
 
```

# Features 
The following features/requirements are implemented.
- Chat server with multiple clients by telnet
- Message sending and relaying to all clients
    ```
    ---------------------------
    * Nash has joined! *
    ---------------------------
    Message>Hi
    ---------------------------
    04/04/2022 12:01:18(User:Nash): "Hi"
    ```
- Configuration provided through app.env
- Logging messages to log file
- HTTP Rest end points to post message
    ```
        curl --location --request POST 'http://localhost:8085/message' \
        --header 'Content-Type: application/json' \
        --data-raw '{
            "name": "Test",
            "message": "Helllo 2"
        }'
    ```
- HTTP Rest end points to query message
    ```
    curl --location --request GET 'http://localhost:8085/message/{username}'
    ```
- Support of channels, so that only clients connected in the same room / channel receive the messages! 

# TODO
 - Test Scripts and mocking
 - Validations
 - Ignore option, where a client can choose to ignore (unsubscribe) from another client's messages
 - Dockerization of the app
 - Adding support for web clients using web scokets

# Design Assumptions
I made following assumptions during my development
- HTTP endpoints for POST and query only by already established user connection
- Temporary persistence for showcasing chat history to support the query mesage feature

# Libraries
- Lumberjack (gopkg.in/natefinch/lumberjack.v2) for logging messages to file with log rotation support
- Viper (github.com/spf13/viper) For configiring application parms in file.
- Mux (github.com/gorilla/mux) For http server implementation
- Go-Redis (github.com/go-redis/redis) Go client for redis

# Libraries
- Reference google and some medium posts for the implementation.
