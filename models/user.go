package models

import (
	"fmt"
	"net/mail"
	"strings"
)

type User struct {
	ID        int     `json:"id"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	UserRoles []Roles `json:"roles"`
}

func (u User) Validate() []string {
	var validationErrors []string
	const missingRequiredFieldMsg = "missing required field %s"

	if u.Email == "" {
		validationErrors = append(validationErrors, fmt.Sprintf(missingRequiredFieldMsg, "email"))
	} else {
		_, err := mail.ParseAddress(u.Email)
		if err != nil {
			errStr := strings.ReplaceAll(err.Error(), "mail: ", "")
			validationErrors = append(validationErrors, fmt.Sprintf("invalid email: %s", errStr))
		}
	}
	if len(strings.Trim(u.Password, " ")) == 0 {
		validationErrors = append(validationErrors, fmt.Sprintf(missingRequiredFieldMsg, "password"))
	}

	return validationErrors
}
