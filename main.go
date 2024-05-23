package main

import (
	"Forum/go/Server"
	"Forum/go/database"
)

func main() {
	database.InitDB("./db/database.db")
	server.HandleFunc()
}
