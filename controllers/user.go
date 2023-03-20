package controllers

import (
	"jwt-auth-service/models"
	"jwt-auth-service/repositories"

	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	UserRepository repositories.IUserRepository
}

func (uc UserController) GetUserByID(id int) (models.User, error) {
	return uc.UserRepository.GetUserByID(id)
}

func (uc UserController) GetUserByEmail(email string) (models.User, error) {
	return uc.UserRepository.GetUserByEmail(email)
}

func (uc UserController) AddUser(user models.User) (models.User, models.ErrorResponse) {
	if errors := user.Validate(); errors != nil {
		return user, models.ErrorResponse{ErrorMessage: "validation errors occurred", Errors: errors}
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		return user, models.ErrorResponse{ErrorMessage: "failed to encrypt password"}
	}

	user.Password = string(hashedPass)

	addedUser, err := uc.UserRepository.AddUser(user)
	if err != nil {
		return addedUser, models.ErrorResponse{ErrorMessage: err.Error()}
	}

	return addedUser, models.ErrorResponse{}
}

func (uc UserController) GetUserWithCredentials(email string, password string) (models.User, error) {
	return uc.UserRepository.GetUserWithCredentials(email, password)
}

func (uc UserController) DeleteUser(id int) error {
	return uc.UserRepository.DeleteUser(id)
}
