package main

import (
	"fmt"

	"github.com/stolostron/recommends/pkg/server"
)

func main() {
	fmt.Println("Starting Recommender")
	server.StartServer()

}
