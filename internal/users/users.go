package users

import (
	"errors"

	"github.com/Groskilled/Chirpy/internal/database"
	"github.com/Groskilled/Chirpy/internal/entities"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string
}

func (u User) GetID() int {
	return u.Id
}

func (u User) GetEmail() string {
	return u.Email
}

func (u User) GetPassword() string {
	return u.Password
}

func GetUserByEmail(db database.DB, email string) (User, error) {
	dbStruct, err := db.LoadDB()
	res := User{}
	if err != nil {
		return res, err
	}
	if dbStruct.Users == nil {
		return res, errors.New("no user available")
	}
	for _, usr := range dbStruct.Users {
		if usr.GetEmail() == email {
			return User{Id: usr.GetID(), Email: usr.GetEmail(), Password: usr.GetEmail()}, nil
		}
	}
	return res, nil
}

func CreateUser(db database.DB, id int, email string, password string) (User, error) {
	newUser := User{
		Id:       id,
		Email:    email,
		Password: password,
	}
	dbStruct, err := db.LoadDB()
	if err != nil {
		return User{}, err
	}
	if dbStruct.Users == nil {
		dbStruct.Users = make(map[int]entities.UserInterface)
	}
	dbStruct.Users[newUser.Id] = newUser
	err = db.WriteDB(dbStruct)
	if err != nil {
		return User{}, nil
	}
	return newUser, nil
}
