package main

import (
	"encoding/gob"
	"log"

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
		log.Fatalf("cannot initialize server: %s", err)
	}

	err = server.ListenAndServe()

	if err != nil {
		log.Fatalf("cannot initialize server: %s", err)
	}
}
