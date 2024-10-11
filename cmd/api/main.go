package main

import (
	"encoding/gob"
	"fmt"

	"pf2.encounterbrew.com/internal/models"
	"pf2.encounterbrew.com/internal/server"
)

func init() {
	gob.Register(&models.Encounter{})
	gob.Register(&models.Player{})
	gob.Register(&models.Monster{})
	gob.Register(map[string]interface{}{})
}

func main() {
	server, err := server.NewServer()
	if err != nil {
		panic(fmt.Sprintf("cannot initialize server: %s", err))
	}

	err = server.ListenAndServe()

	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
