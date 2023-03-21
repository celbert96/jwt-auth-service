package repositories

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"jwt-auth-service/models"
	"testing"
)

type MockUserRepository struct {
}

var testUserData = map[int]models.User{
	0: {
		ID:        0,
		Email:     "email0@gmail.com",
		Password:  "password0",
		UserRoles: testUserRoles,
	},
	1: {
		ID:        1,
		Email:     "email1@gmail.com",
		Password:  "password1",
		UserRoles: testUserRoles,
	},
	2: {
		ID:        2,
		Email:     "email2@gmail.com",
		Password:  "password2",
		UserRoles: testUserRoles,
	},
	3: {
		ID:        3,
		Email:     "email3@gmail.com",
		Password:  "password3",
		UserRoles: testUserRoles,
	},
}

var testExistingUser = testUserData[1]
var testNewUserId = 4
var testUserRoles = []models.Roles{models.UserRole}

func (repo MockUserRepository) AddUser(user models.User) (models.User, error) {
	copyData := copyTestUserData(testUserData)
	user.ID = testNewUserId
	user.UserRoles = testUserRoles

	el, ok := copyData[user.ID]
	if ok {
		return el, fmt.Errorf("user already exists with id %d", user.ID)
	}

	for _, v := range testUserData {
		if v.Email == user.Email {
			return v, fmt.Errorf("user already exists with email %s", user.Email)
		}
	}

	copyData[user.ID] = user

	el, ok = copyData[user.ID]
	if !ok {
		return el, fmt.Errorf("failed to add user to user data")
	}

	return el, nil
}

func (repo MockUserRepository) DeleteUser(userId int) error {
	copyData := copyTestUserData(testUserData)
	_, ok := copyData[userId]
	if !ok {
		return fmt.Errorf("user does not exist")
	}

	delete(copyData, userId)
	_, ok = copyData[userId]
	if ok {
		return fmt.Errorf("failed to delete user from user data")
	}

	return nil
}

func (repo MockUserRepository) GetUserById(userId int) (models.User, error) {
	el, ok := testUserData[userId]
	if !ok {
		return el, fmt.Errorf("no user exists with id %d", userId)
	}

	return el, nil
}

func (repo MockUserRepository) GetUserByEmail(email string) (models.User, error) {
	for _, v := range testUserData {
		if v.Email == email {
			return v, nil
		}
	}

	return models.User{}, fmt.Errorf("no user exists with email %s", email)
}

func (repo MockUserRepository) GetUserWithCredentials(email string, password string) (models.User, error) {
	for _, v := range testUserData {
		if v.Email == email && v.Password == password {
			return v, nil
		}
	}

	return models.User{}, fmt.Errorf("no user exists with email %s and password %s", email, password)
}

func TestAddUser(t *testing.T) {
	user := models.User{
		Email:    "brandnewemail@email.com",
		Password: "aosdkpoak",
	}

	expectedUser := models.User{
		ID:        testNewUserId,
		Email:     user.Email,
		Password:  user.Password,
		UserRoles: testUserRoles,
	}

	addedUser, err := MockUserRepository{}.AddUser(user)

	if err != nil {
		t.Fatalf("failed to add user: %q", err)
	}

	if !cmp.Equal(addedUser, expectedUser) {
		t.Fatalf("added user had unexpected values: \n\tactual: %q\n\texpected: %q", addedUser, expectedUser)
	}
}

func TestAddUserFailsExistingEmail(t *testing.T) {
	user := models.User{
		Email:    testExistingUser.Email,
		Password: "aosdkpoak",
	}

	_, err := MockUserRepository{}.AddUser(user)

	if err == nil {
		t.Fatalf("no error was thrown when adding a duplicate user with email %s", user.Email)
	}
}

func TestDeleteUser(t *testing.T) {
	err := MockUserRepository{}.DeleteUser(testExistingUser.ID)
	if err != nil {
		t.Fatalf("failed to delete user: %q", err)
	}
}

func TestDeleteUserFailsNonExistentId(t *testing.T) {
	err := MockUserRepository{}.DeleteUser(99)
	if err == nil {
		t.Fatalf("no error thrown when deleting user id that doesn't exist")
	}
}

func TestGetUserById(t *testing.T) {
	user, err := MockUserRepository{}.GetUserById(testExistingUser.ID)
	if err != nil {
		t.Fatalf("error occurred when getting user with id: %d > %q", testExistingUser.ID, err)
	}

	if user.ID != testExistingUser.ID {
		t.Fatalf("unexpected result\n\texpected: user with id %d\n\tactual: user with id %d", testExistingUser.ID, user.ID)
	}
}

func TestGetUserByIdFailsNonExistentId(t *testing.T) {
	_, err := MockUserRepository{}.GetUserById(99)
	if err == nil {
		t.Fatalf("no error was thrown when fetching user with non-existent id %d", 99)
	}
}

func TestGetUserByEmail(t *testing.T) {
	user, err := MockUserRepository{}.GetUserByEmail(testExistingUser.Email)
	if err != nil {
		t.Fatalf("error occurred when getting user with email %s: %q", testExistingUser.Email, err)
	}

	if user.Email != testExistingUser.Email {
		t.Fatalf("unexpected result\n\texpected: user with email %s\n\tactual: user with email%s\n\t", testExistingUser.Email, err)
	}
}

func TestGetUserByEmailFailsNonExistentEmail(t *testing.T) {
	fakeEmail := "nonexistentemail@email.com"
	_, err := MockUserRepository{}.GetUserByEmail(fakeEmail)
	if err == nil {
		t.Fatalf("no error was thrown when fetching user with non-existent email %s", fakeEmail)
	}
}

func TestGetUserWithCredentials(t *testing.T) {
	existingEmail := testExistingUser.Email
	existingPassword := testExistingUser.Password

	user, err := MockUserRepository{}.GetUserWithCredentials(existingEmail, existingPassword)
	if err != nil {
		t.Fatalf("error occurred when getting user with email %s and password %s: %q", existingEmail, existingPassword, err)
	}

	if user.Email != existingEmail || user.Password != existingPassword {
		t.Fatalf(
			"unexpected result\n\texpected: user with email %s and password %s\n\tactual: user with email %s and password %s",
			existingEmail,
			existingPassword,
			user.Email,
			user.Password)
	}
}

func TestGetUserWithCredentialsFailsInvalidEmail(t *testing.T) {
	nonExistentEmail := "nonexistentemail@email.com"
	existingPassword := testExistingUser.Password

	_, err := MockUserRepository{}.GetUserWithCredentials(nonExistentEmail, existingPassword)

	if err == nil {
		t.Fatalf("no error was thrown when getting user with non-existent email %s", nonExistentEmail)
	}
}

func TestGetUserWithCredentialsFailsInvalidPassword(t *testing.T) {
	existingEmail := testExistingUser.Email
	nonExistentPassword := "notrealpasswd"

	_, err := MockUserRepository{}.GetUserWithCredentials(existingEmail, nonExistentPassword)

	if err == nil {
		t.Fatalf("no error was thrown when getting user with incorrect password %s", nonExistentPassword)
	}
}

func copyTestUserData(m map[int]models.User) map[int]models.User {
	newMap := make(map[int]models.User)
	for k, v := range m {
		newMap[k] = v
	}

	return newMap
}
