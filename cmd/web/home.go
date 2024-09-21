package web

import (
	"log"
	"net/http"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/models"
)

func HomeHandler(db database.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Fetch all monsters (now using cache)
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
}
