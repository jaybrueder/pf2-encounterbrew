package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"pf2.encounterbrew.com/internal/database"
)

func main() {
	dbService := database.New()

	err := filepath.Walk("data/bestiary", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if filepath.Ext(path) == ".json" {
            if err := seedFile(dbService, path); err != nil {
                return err
            }
        }
        return nil
    })

 	if err != nil {
        fmt.Fprintf(os.Stderr, "Error seeding data: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("Data seeded successfully")
}

func seedFile(conn database.Service, filePath string) error {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("unable to read file %s: %v", filePath, err)
    }

    var character map[string]interface{}
    err = json.Unmarshal(data, &character)
    if err != nil {
        return fmt.Errorf("unable to parse JSON from %s: %v", filePath, err)
    }

    err = conn.InsertJson("monsters", character)
    if err != nil {
        return fmt.Errorf("unable to insert data from %s: %v", filePath, err)
    }

    fmt.Printf("Seeded data from %s\n", filePath)
    return nil
}
