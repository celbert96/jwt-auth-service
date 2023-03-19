package repositories

import (
	"database/sql"
	"fmt"
	"jwt-auth-service/models"
	"log"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type IUserRepository interface {
	AddUser(models.User) (int, error)
	DeleteUser(int) error
	GetUserByID(int) (models.User, error)
	GetUserByEmail(string) (models.User, error)
	GetUserWithCredentials(string, string) (models.User, error)
}

type UserRepository struct {
	DBConn *sql.DB
}

func (repo UserRepository) AddUser(user models.User) (int, error) {
	dbConn := repo.DBConn

	result, err := dbConn.Exec("INSERT INTO USERS (EMAIL, PASSWORD) VALUES (?, ?)",
		user.Email, user.Password)

	if err != nil {
		mysqlerr, _ := err.(*mysql.MySQLError)
		if mysqlerr != nil {
			if mysqlerr.Number == 1062 {
				return 0, fmt.Errorf("user already exists")
			}
		}

		log.Printf("repositories > user.go > AddUser > error: %s", err.Error())
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = dbConn.Exec("INSERT INTO USER_ROLES (USER_ID, ROLE_ID) VALUES (?, ?)", id, models.USER_ROLE)

	if err != nil {
		log.Println("repositories > user.go > AddUser > error: %s" + err.Error())
	}

	return int(id), nil
}

func (repo UserRepository) GetUserByID(id int) (models.User, error) {
	dbConn := repo.DBConn

	row := dbConn.QueryRow("SELECT ID, EMAIL FROM USERS WHERE ID = ?", id)

	var user models.User
	err := row.Scan(&user.ID, &user.Email)

	if err != nil {
		return user, err
	}

	rolesResult, err := dbConn.Query("SELECT ROLE_ID FROM USER_ROLES WHERE USER_ID = ?", id)

	if err != nil {
		return user, err
	}

	for rolesResult.Next() {
		var roleId models.Roles
		err := rolesResult.Scan(&roleId)
		if err != nil {
			return user, err
		}

		user.UserRoles = append(user.UserRoles, roleId)
	}

	return user, nil
}

func (repo UserRepository) GetUserByEmail(email string) (models.User, error) {
	dbConn := repo.DBConn

	row := dbConn.QueryRow("SELECT ID, EMAIL FROM USERS WHERE EMAIL = ?", email)

	var user models.User
	err := row.Scan(&user.ID, &user.Email)

	if err != nil {
		return user, err
	}

	rolesResult, err := dbConn.Query("SELECT ROLE_ID FROM USER_ROLES WHERE USER_ID = ?", user.ID)

	if err != nil {
		return user, err
	}

	for rolesResult.Next() {
		var roleId models.Roles
		err := rolesResult.Scan(&roleId)
		if err != nil {
			return user, err
		}

		user.UserRoles = append(user.UserRoles, roleId)
	}

	return user, nil
}

func (repo UserRepository) GetUserWithCredentials(email string, password string) (models.User, error) {
	dbConn := repo.DBConn

	row := dbConn.QueryRow("SELECT ID, EMAIL, PASSWORD FROM USERS WHERE EMAIL = ?", email)

	var user models.User
	err := row.Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("invalid credentials")
		}
		mysqlerr, _ := err.(*mysql.MySQLError)
		if mysqlerr != nil {
			log.Printf("routes > user.go > GetUserWithCredentials > error: %d", mysqlerr.Number)
			err = fmt.Errorf("internal error")
		}

		log.Printf("routes > user.go > GetUserWithCredentials > error: " + err.Error())
		return models.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println(err.Error())
		err = fmt.Errorf("invalid credentials")
		return models.User{}, err
	}

	rolesResult, err := dbConn.Query("SELECT ROLE_ID FROM USER_ROLES WHERE USER_ID = ?", user.ID)

	if err != nil {
		return user, err
	}

	for rolesResult.Next() {
		var roleId models.Roles
		err := rolesResult.Scan(&roleId)
		if err != nil {
			return user, err
		}

		user.UserRoles = append(user.UserRoles, roleId)
	}

	return user, nil
}

func (repo UserRepository) DeleteUser(id int) error {
	dbConn := repo.DBConn
	_, err := dbConn.Exec("DELETE FROM USERS WHERE ID = ?", id)
	if err != nil {
		log.Printf("repositories > user.go > DeleteUser > error deleting user with ID %d\n", id)
		return err
	}
	_, err = dbConn.Exec("DELETE FROM USER_ROLES WHERE ID = ?", id)
	if err != nil {
		log.Printf("repositories > user.go > DeleteUser > error deleting user roles for User ID %d\n", id)
	}

	return err
}
