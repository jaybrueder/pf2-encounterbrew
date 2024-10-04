package main

import (
	"fmt"
	"encoding/gob"

	"pf2.encounterbrew.com/internal/server"
	"pf2.encounterbrew.com/internal/models"
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
