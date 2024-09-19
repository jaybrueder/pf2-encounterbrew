package main

import (
	"fmt"

	"pf2.encounterbrew.com/internal/server"
)

func main() {
	server := server.NewServer()
	err := server.ListenAndServe()

	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
