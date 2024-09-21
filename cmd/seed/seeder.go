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

	// Seed conditions
	err := filepath.Walk("data/conditions", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if filepath.Ext(path) == ".json" {
            if err := seedFile(dbService, path, "conditions"); err != nil {
                return err
            }
        }
        return nil
    })

	// Seed monsters
	err = filepath.Walk("data/bestiaries", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if filepath.Ext(path) == ".json" {
            if err := seedFile(dbService, path, "monsters"); err != nil {
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

func seedFile(db database.Service, filePath string, table string) error {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("unable to read file %s: %v", filePath, err)
    }

    var jsonData map[string]interface{}
    err = json.Unmarshal(data, &jsonData)
    if err != nil {
        return fmt.Errorf("unable to parse JSON from %s: %v", filePath, err)
    }

    _, err = db.Insert(table, []string{"data"}, jsonData)
    if err != nil {
        return fmt.Errorf("unable to insert data from %s: %v", filePath, err)
    }

    fmt.Printf("Seeded data from %s\n", filePath)
    return nil
}
