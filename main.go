package main

import (
    "Forum/go/Server" 
)

func main() {
	server.InitDB("./db/database.db")
    server.HandleFunc()
}
