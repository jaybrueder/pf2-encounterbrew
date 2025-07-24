package encounter

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"pf2.encounterbrew.com/internal/database"
	"pf2.encounterbrew.com/internal/models"
)

func EncounterNewHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		parties, err := models.GetAllParties(db)
		if err != nil {
			log.Printf("Error getting parties: %v", err)
			return c.String(http.StatusInternalServerError, "Error getting parties")
		}

		component := EncounterNew(parties)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterCreateHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		partyID, err := strconv.Atoi(c.FormValue("party_id"))
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid party ID")
		}

		// First, verify that the party exists
		exists, err := models.PartyExists(db, partyID)
		if err != nil {
			log.Printf("Error checking party existence: %v", err)
			return c.String(http.StatusInternalServerError, "Error creating encounter")
		}
		if !exists {
			return c.String(http.StatusBadRequest, "Selected party does not exist")
		}

		_, err = models.CreateEncounter(db, name, partyID)
		if err != nil {
			log.Printf("Error creating encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error creating encounter")
		}

		encounters, err := models.GetAllEncounters(db)
		if err != nil {
			log.Printf("Error fetching encounters: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounters")
		}

		// Render the template with the encounter
		component := EncounterList(encounters)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterEditHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("encounter_id"))
		if err != nil {
			log.Printf("Invalid encounter ID: %v", err)
			return c.String(http.StatusBadRequest, "Invalid encounter ID")
		}

		encounter, err := models.GetEncounter(db, id)
		if err != nil {
			log.Printf("Error getting encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error getting encounter")
		}

		parties, err := models.GetAllParties(db)
		if err != nil {
			log.Printf("Error getting parties: %v", err)
			return c.String(http.StatusInternalServerError, "Error getting parties")
		}

		component := EncounterEdit(encounter, parties)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterUpdateHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get encounter ID from URL parameter
		encounterID, err := strconv.Atoi(c.Param("encounter_id"))
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid encounter ID")
		}

		// Get form values
		name := c.FormValue("name")
		partyID, err := strconv.Atoi(c.FormValue("party_id"))
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid party ID")
		}

		// Verify that the party exists
		exists, err := models.PartyExists(db, partyID)
		if err != nil {
			log.Printf("Error checking party existence: %v", err)
			return c.String(http.StatusInternalServerError, "Error updating encounter")
		}
		if !exists {
			return c.String(http.StatusBadRequest, "Selected party does not exist")
		}

		// Update the encounter
		err = models.UpdateEncounter(db, encounterID, name, partyID)
		if err != nil {
			log.Printf("Error updating encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error updating encounter")
		}

		// Fetch the encounter from the database
		encounter, err := models.GetEncounterWithCombatants(db, encounterID)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Render the template with the encounter
		component := EncounterShow(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterDeleteHandler(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get encounter ID from URL parameter
		encounterID, err := strconv.Atoi(c.Param("encounter_id"))
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid encounter ID")
		}

		// Delete the encounter
		err = models.DeleteEncounter(db, encounterID)
		if err != nil {
			log.Printf("Error deleting encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error deleting encounter")
		}

		// Fetch all encounters to refresh the list
		encounters, err := models.GetAllEncounters(db)
		if err != nil {
			log.Printf("Error fetching encounters: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounters")
		}

		// Render the updated list
		component := EncounterList(encounters)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

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
		encounterID, _ := strconv.Atoi(c.Param("encounter_id"))

		// Fetch the encounter from the database
		encounter, err := getEncounter(db, encounterID)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Render the template with the encounter
		component := EncounterShow(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterSearchMonster(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		search := c.FormValue("search")
		encounterID := c.Param("encounter_id")

		// Parse the form to get array values
		if err := c.Request().ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
		}

		// Parse filter parameters
		filters := models.MonsterSearchFilters{}

		// Level range filter
		if minLevelStr := c.FormValue("min_level"); minLevelStr != "" {
			if minLevel, err := strconv.Atoi(minLevelStr); err == nil {
				filters.MinLevel = &minLevel
			}
		}

		if maxLevelStr := c.FormValue("max_level"); maxLevelStr != "" {
			if maxLevel, err := strconv.Atoi(maxLevelStr); err == nil {
				filters.MaxLevel = &maxLevel
			}
		}

		// Excluded sources filter
		if excludedSources := c.Request().Form["excluded_sources[]"]; len(excludedSources) > 0 {
			filters.ExcludedSources = excludedSources
		}

		// Excluded sizes filter
		if excludedSizes := c.Request().Form["excluded_sizes[]"]; len(excludedSizes) > 0 {
			filters.ExcludedSizes = excludedSizes
		}

		monsters, err := models.SearchMonstersWithFilters(db, search, filters)
		if err != nil {
			log.Printf("Error searching for monster: %v", err)
			return c.String(http.StatusInternalServerError, "Error searching for monster")
		}

		// Get the encounter to access party level
		encID, _ := strconv.Atoi(encounterID)
		encounter, err := models.GetEncounter(db, encID)
		if err != nil {
			log.Printf("Error getting encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error getting encounter")
		}

		// Get party
		party, err := models.GetParty(db, encounter.PartyID)
		if err != nil {
			log.Printf("Error getting party: %v", err)
			party = models.Party{} // Use empty party if error
		}

		component := MonsterSearchResults(encounterID, monsters, party.GetLevel())
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterAddMonster(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID, _ := strconv.Atoi(c.Param("encounter_id"))
		monsterID, _ := strconv.Atoi(c.Param("monster_id"))
		levelAdjustment, _ := strconv.Atoi(c.FormValue("level_adjustment"))

		monster, err := models.GetMonster(db, monsterID)
		if err != nil {
			log.Printf("Error finding monster: %v", err)
			return c.String(http.StatusInternalServerError, "Error finding monster")
		}

		encounter, err := models.AddMonsterToEncounter(db, encounterID, monsterID, levelAdjustment, monster.GenerateInitiative())
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
		encounterID, _ := strconv.Atoi(c.Param("encounter_id"))
		associationID, _ := strconv.Atoi(c.Param("association_id"))

		err := models.RemoveMonsterFromEncounter(db, encounterID, associationID)
		if err != nil {
			log.Printf("Error removing monster: %v", err)
			return c.String(http.StatusInternalServerError, "Error removing monster")
		}

		// Fetch the encounter from the database
		encounter, err := getEncounter(db, encounterID)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		component := MonstersAdded(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func EncounterRemoveCombatant(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID, _ := strconv.Atoi(c.Param("encounter_id"))
		associationID, _ := strconv.Atoi(c.Param("association_id"))
		isMonster, _ := strconv.ParseBool(c.Param("is_monster"))

		if isMonster {
			log.Printf("Removing monster: %v", associationID)
			err := models.RemoveMonsterFromEncounter(db, encounterID, associationID)
			if err != nil {
				log.Printf("Error removing monster: %v", err)
				return c.String(http.StatusInternalServerError, "Error removing monster")
			}
		} else {
			log.Printf("Removing player: %v", associationID)
			err := models.RemovePlayerFromEncounter(db, encounterID, associationID)
			if err != nil {
				log.Printf("Error removing monster: %v", err)
				return c.String(http.StatusInternalServerError, "Error removing monster")
			}
		}

		// Fetch the encounter from the database
		encounter, err := getEncounter(db, encounterID)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Render and return the updated combatant list
		component := CombatantList(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func UpdateCombatant(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID, _ := strconv.Atoi(c.Param("encounter_id"))
		combatantIndex, _ := strconv.Atoi(c.Param("index"))

		// Fetch the encounter from the database
		encounter, err := getEncounter(db, encounterID)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Update the specific combatant's values
		if combatantIndex < len(encounter.Combatants) {
			// Check if initiative was provided
			if initiativeStr := c.FormValue("initiative"); initiativeStr != "" {
				if newInitiative, err := strconv.Atoi(initiativeStr); err == nil {
					if err := encounter.Combatants[combatantIndex].SetInitiative(db, newInitiative); err != nil {
						log.Printf("Error updating initiative: %v", err)
					}
					// Re-sort combatants by initiative only if initiative was updated
					models.SortCombatantsByInitiative(encounter.Combatants)
				}
			}

			// Check if damage was provided
			if damageStr := c.FormValue("damage"); damageStr != "" {
				if damage, err := strconv.Atoi(damageStr); err == nil {
					if err := encounter.Combatants[combatantIndex].SetHp(db, damage); err != nil {
						log.Printf("Error updating hp: %v", err)
					}
				}
			}
		}

		// Render and return the updated combatant list
		component := CombatantList(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func BulkUpdateInitiative(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID, _ := strconv.Atoi(c.Param("encounter_id"))

		// Fetch the encounter from the database
		encounter, err := getEncounter(db, encounterID)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Update the each combatant's initiative
		for i, combatant := range encounter.Combatants {
			newInitiative, _ := strconv.Atoi(c.FormValue("initiative-" + strconv.Itoa(i)))
			if err := combatant.SetInitiative(db, newInitiative); err != nil {
				log.Printf("Error updating initiative: %v", err)
			}
		}

		// Re-sort combatants by initiative
		models.SortCombatantsByInitiative(encounter.Combatants)

		component := CombatantList(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func ChangeTurn(db database.Service, next bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID, _ := strconv.Atoi(c.Param("encounter_id"))

		// Fetch the encounter from the database
		encounter, err := getEncounter(db, encounterID)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		numberOfCombatants := len(encounter.Combatants)
		fmt.Printf("Combatants: %d", numberOfCombatants)

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

		fmt.Printf("Turn: %d", encounter.Turn)

		// Persist turn and round
		err = models.UpdateTurnAndRound(db, encounter.Turn, encounter.Round, encounterID)
		if err != nil {
			log.Printf("Error updating turn and round: %v", err)
			return c.String(http.StatusInternalServerError, "Error updating turn and round")
		}

		// Render and return the updated combatant list
		component := EncounterShow(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)

	}
}

func AddCondition(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID, _ := strconv.Atoi(c.Param("encounter_id"))
		conditionID, _ := strconv.Atoi(c.Param("condition_id"))
		combatantIndex, _ := strconv.Atoi(c.Param("index"))

		// Fetch the encounter from the database
		encounter, err := getEncounter(db, encounterID)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Update the specific combatant's values
		condition, err := models.GetCondition(db, conditionID)
		if err != nil {
			log.Printf("Error getting condition: %v", err)
			return c.String(http.StatusInternalServerError, "Error getting condition")
		}

		combatant := encounter.Combatants[combatantIndex]
		hasCondition := combatant.HasCondition(conditionID)
		isValued := condition.IsValued()

		if hasCondition && isValued {
			// Increment existing valued condition by 1
			err = combatant.SetCondition(db, encounterID, conditionID, 1)
			if err != nil {
				log.Printf("Error incrementing condition: %v", err)
				return c.String(http.StatusInternalServerError, "Error incrementing condition")
			}
		} else if !hasCondition {
			// Set new condition with value 1 if valued, 0 if not
			value := 0
			if isValued {
				value = 1
			}

			err = combatant.SetCondition(db, encounterID, conditionID, value)
			if err != nil {
				log.Printf("Error setting condition: %v", err)
				return c.String(http.StatusInternalServerError, "Error setting condition")
			}
		}

		// Render and return the updated combatant list
		component := CombatantList(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func RemoveCondition(db database.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		encounterID, _ := strconv.Atoi(c.Param("encounter_id"))
		conditionID, _ := strconv.Atoi(c.Param("condition_id"))
		combatantIndex, _ := strconv.Atoi(c.Param("index"))

		// Fetch the encounter from the database
		encounter, err := getEncounter(db, encounterID)
		if err != nil {
			log.Printf("Error fetching encounter: %v", err)
			return c.String(http.StatusInternalServerError, "Error fetching encounter")
		}

		// Update the specific combatant's values
		err = encounter.Combatants[combatantIndex].RemoveCondition(db, encounterID, conditionID)
		if err != nil {
			log.Printf("Error removing condition: %v", err)
			return c.String(http.StatusInternalServerError, "Error removing condition")
		}

		// Render and return the updated combatant list
		component := CombatantList(encounter)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func getEncounter(db database.Service, encounterID int) (models.Encounter, error) {
	// Fetch the encounter from the database
	encounter, err := models.GetEncounterWithCombatants(db, encounterID)
	if err != nil {
		log.Printf("Error fetching encounter: %v", err)
		return models.Encounter{}, err
	}

	// Get all conditions
	groupedConditions, err := models.GetGroupedConditions(db)
	if err != nil {
		log.Printf("Error fetching grouped conditions: %v", err)
		return models.Encounter{}, err
	}
	encounter.GroupedConditions = groupedConditions

	// Sort the combatants by initiative
	models.SortCombatantsByInitiative(encounter.Combatants)

	return encounter, nil
}
