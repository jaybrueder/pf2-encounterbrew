package models

import ()

type Party struct {
	ID   int    	`json:"id"`
	Name string 	`json:"name"`
 	UserID int		`json:"user_id"`
    User   *User	`json:"user,omitempty"`
}
