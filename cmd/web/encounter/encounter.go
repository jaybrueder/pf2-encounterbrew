package encounter

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/models"
)

func EncounterListHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Fetch all encounters for given user
		encounters, err := models.GetAllEncounters(db)
		if err != nil {
			log.Printf("Error fetching encounters: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounters")
		}

		// Render the template with the encounters
		component := EncounterList(encounters)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterShowHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the encounter ID from the URL path parameter
		id := c.Param("encounter_id")

		// Get active party
		// TODO: Use actual user ID
		user, err := models.GetUserByID(db, 1)
		if err != nil {
			log.Printf("Error fetching users: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching user")
		}

		// Fetch the encounter from the database
		encounter, err := models.GetEncounterWithCombatants(db, id, strconv.Itoa(user.ActivePartyID))
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Get all conditions
		groupedConditions, err := models.GetGroupedConditions(db)
		if err != nil {
			log.Printf("Error fetching grouped conditions: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching conditions")
		}
		encounter.GroupedConditions = groupedConditions

		// Store encounter in session
		sess, _ := session.Get("encounter-session", c)
		sess.Values["encounter"] = encounter
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("Error saving session: %v", err)
		}

		// Render the template with the encounter
		component := EncounterShow(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

// func EncounterEditHandler(db database.Service) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		// Get the encounter ID from the URL path parameter
// 		id := c.Param("encounter_id")

// 		// Fetch the encounter from the database
// 		encounter, err := models.GetEncounter(db, id)
// 		if err != nil {
// 			log.Printf("Error fetching encounter: %v", err)
// 			return c.String(http.StatusInternalServerError, "Error fetching encounter")
// 		}

// 		// Render the template with the encounter
// 		component := EncounterEdit(encounter)
// 		return component.Render(c.Request().Context(), c.Response().Writer)
// 	}
// }

func EncounterSearchMonster(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		search := c.FormValue("search")
		encounterID := c.Param("encounter_id")

		monsters, err := models.SearchMonsters(db, search)
		if err != nil {
			log.Printf("Error searching for monster: %v", err)
			return c.String(http.StatusInternalServerError, "Error searching for monster")
		}

		component := MonsterSearchResults(encounterID, monsters)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterAddMonster(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID := c.Param("encounter_id")
		monsterID := c.Param("monster_id")
		levelAdjustment, _ := strconv.Atoi(c.FormValue("level_adjustment"))

		encounter, err := models.AddMonsterToEncounter(db, encounterID, monsterID, levelAdjustment)
		if err != nil {
			log.Printf("Error adding monster: %v", err)
			return c.String(http.StatusInternalServerError, "Error adding monster")
		}

		component := MonstersAdded(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterRemoveMonster(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID := c.Param("encounter_id")
		associationID := c.Param("association_id")

		encounter, err := models.RemoveMonsterFromEncounter(db, encounterID, associationID)
		if err != nil {
			log.Printf("Error removing monster: %v", err)
			return c.String(http.StatusInternalServerError, "Error removing monster")
		}

		component := MonstersAdded(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func UpdateCombatant() echo.HandlerFunc {
	return func(c echo.Context) error {
		combatantIndex, _ := strconv.Atoi(c.Param("index"))

		// Get session
		sess, _ := session.Get("encounter-session", c)

		// Get encounter from session
		encounterData, ok := sess.Values["encounter"]
		if !ok {
			return c.String(http.StatusInternalServerError, "Encounter not found in session")
		}

		encounter, ok := encounterData.(*models.Encounter)
		if !ok {
			log.Printf("Type assertion failed. Actual type: %T", encounterData)
			return c.String(http.StatusInternalServerError, "Invalid encounter data in session")
		}

		// Update the specific combatant's values
		if combatantIndex < len(encounter.Combatants) {
			// Check if initiative was provided
			if initiativeStr := c.FormValue("initiative"); initiativeStr != "" {
				if newInitiative, err := strconv.Atoi(initiativeStr); err == nil {
					encounter.Combatants[combatantIndex].SetInitiative(newInitiative)
					// Re-sort combatants by initiative only if initiative was updated
					models.SortCombatantsByInitiative(encounter.Combatants)
				}
			}

			// Check if damage was provided
			if damageStr := c.FormValue("damage"); damageStr != "" {
				if damage, err := strconv.Atoi(damageStr); err == nil {
					encounter.Combatants[combatantIndex].SetHp(damage)
				}
			}
		}

		// Save updated encounter back to session
		sess.Values["encounter"] = encounter
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("Error saving session: %v", err)
			return c.String(http.StatusInternalServerError, "Error saving session")
		}

		// Render and return the updated combatant list
		component := CombatantList(*encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func BulkUpdateInitiative() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session
		sess, _ := session.Get("encounter-session", c)

		// Get encounter from session
		encounterData, ok := sess.Values["encounter"]
		if !ok {
			return c.String(http.StatusInternalServerError, "Encounter not found in session")
		}

		encounter, ok := encounterData.(*models.Encounter)
		if !ok {
			log.Printf("Type assertion failed. Actual type: %T", encounterData)
			return c.String(http.StatusInternalServerError, "Invalid encounter data in session")
		}

		// Update the each combatant's initiative
		for i, combatant := range encounter.Combatants {
			newInitiative, _ := strconv.Atoi(c.FormValue("initiative-" + strconv.Itoa(i)))
			combatant.SetInitiative(newInitiative)
		}

		// Re-sort combatants by initiative
		models.SortCombatantsByInitiative(encounter.Combatants)

		// Save updated encounter back to session
		sess.Values["encounter"] = encounter
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("Error saving session: %v", err)
			return c.String(http.StatusInternalServerError, "Error saving session")
		}

		component := CombatantList(*encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func ChangeTurn(next bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session
		sess, _ := session.Get("encounter-session", c)

		// Get encounter from session
		encounterData, ok := sess.Values["encounter"]
		if !ok {
			return c.String(http.StatusInternalServerError, "Encounter not found in session")
		}

		encounter, ok := encounterData.(*models.Encounter)
		if !ok {
			log.Printf("Type assertion failed. Actual type: %T", encounterData)
			return c.String(http.StatusInternalServerError, "Invalid encounter data in session")
		}

		numberOfCombatants := len(encounter.Combatants)

		if next {
			if encounter.Turn == numberOfCombatants-1 {
				encounter.Turn = 0
				encounter.Round += 1
			} else {
				encounter.Turn += 1
			}
		} else {
			if encounter.Turn == 0 {
				encounter.Turn = numberOfCombatants - 1
				encounter.Round -= 1
			} else {
				encounter.Turn -= 1
			}
		}

		if encounter.Round < 0 {
			encounter.Round = 0
			encounter.Turn = 0
		}

		// Save updated encounter back to session
		sess.Values["encounter"] = encounter
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("Error saving session: %v", err)
			return c.String(http.StatusInternalServerError, "Error saving session")
		}

		// Render and return the updated combatant list
		component := EncounterShow(*encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)

	}
}

func AddCondition(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		conditionID, _ := strconv.Atoi(c.Param("condition_id"))
		combatantIndex, _ := strconv.Atoi(c.Param("index"))

		// Get session
		sess, _ := session.Get("encounter-session", c)

		// Get encounter from session
		encounterData, ok := sess.Values["encounter"]
		if !ok {
			return c.String(http.StatusInternalServerError, "Encounter not found in session")
		}

		encounter, ok := encounterData.(*models.Encounter)
		if !ok {
			log.Printf("Type assertion failed. Actual type: %T", encounterData)
			return c.String(http.StatusInternalServerError, "Invalid encounter data in session")
		}

		// Update the specific combatant's values
		// TODO Increasr value if already there
		encounter.Combatants[combatantIndex].SetCondition(db, conditionID, 0)

		// Save updated encounter back to session
		sess.Values["encounter"] = encounter
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("Error saving session: %v", err)
			return c.String(http.StatusInternalServerError, "Error saving session")
		}

		// Render and return the updated combatant list
		component := CombatantList(*encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func RemoveCondition() echo.HandlerFunc {
	return func(c echo.Context) error {
		conditionID, _ := strconv.Atoi(c.Param("condition_id"))
		combatantIndex, _ := strconv.Atoi(c.Param("index"))

		// Get session
		sess, _ := session.Get("encounter-session", c)

		// Get encounter from session
		encounterData, ok := sess.Values["encounter"]
		if !ok {
			return c.String(http.StatusInternalServerError, "Encounter not found in session")
		}

		encounter, ok := encounterData.(*models.Encounter)
		if !ok {
			log.Printf("Type assertion failed. Actual type: %T", encounterData)
			return c.String(http.StatusInternalServerError, "Invalid encounter data in session")
		}

		// Update the specific combatant's values
		encounter.Combatants[combatantIndex].RemoveCondition(conditionID)

		// Save updated encounter back to session
		sess.Values["encounter"] = encounter
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			log.Printf("Error saving session: %v", err)
			return c.String(http.StatusInternalServerError, "Error saving session")
		}

		// Render and return the updated combatant list
		component := CombatantList(*encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}
