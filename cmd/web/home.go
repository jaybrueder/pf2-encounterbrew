package web

import (
	"log"
	"net/http"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/models"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Get the database connection
	db := database.New()
	defer db.Close()

	// Fetch all monsters
	monsters, err := models.GetAllMonsters(db)
	if err != nil {
		http.Error(w, "Error fetching monsters", http.StatusInternalServerError)
		log.Printf("Error fetching monsters: %v", err)
		return
	}

	// Render the template with the monsters
	component := Home(monsters)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error rendering Home template: %v", err)
	}

}
