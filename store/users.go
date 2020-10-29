package store

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type UserExistError struct {
	err string
}

func (exist UserExistError) Error() string {
	return exist.err
}

type UserDoesNotExistError struct {
	err string
}

func (exist UserDoesNotExistError) Error() string {
	return exist.err
}

type UserAddress struct {
	UserID int    `db:"uid" json:"-"`
	Value  string `db:"address" json:"address"`
}

type User struct {
	UID       int            `db:"id" json:"user_id"`
	FirstName string         `db:"first_name" json:"first_name"`
	LastName  string         `db:"last_name" json:"last_name"`
	Addresses []*UserAddress `json:"addresses"`
}

const (
	INSERT_USER           = "INSERT INTO users (id, first_name, last_name) VALUES ($1, $2, $3) RETURNING id"
	INSERT_ADDRESS        = "INSERT INTO addresses (uid, address) VALUES (:uid, :address);"
	SELECT_USER           = "SELECT id, first_name, last_name FROM users WHERE id=$1"
	SELECT_USER_ADDRESSES = "SELECT address FROM addresses WHERE uid=$1"
)

type UserStore struct {
	dbDriver *sqlx.DB
}

func NewUserStore(DSN string) (*UserStore, error) {
	db, err := sqlx.Connect("postgres", DSN)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to DB: %v ", err)
	}

	store := &UserStore{dbDriver: db}
	return store, nil
}

func (uStore *UserStore) CloseDB() error {
	err := uStore.dbDriver.Close()
	return err
}

func (uStore *UserStore) CreateUser(user *User) error {
	userExist, err := uStore.checkUserExist(user)
	if err != nil {
		return fmt.Errorf("Error checking if user exist %v ", err)
	}
	if userExist {
		return UserExistError{err: "User exist"}
	}
	lastInsertId := 0
	err = uStore.dbDriver.QueryRow(INSERT_USER, user.UID, user.FirstName, user.LastName).Scan(&lastInsertId)
	if err != nil {
		return fmt.Errorf("Error creating user (inserting user) %v ", err)
	}
	user.UID = lastInsertId
	log.Printf("User created: %v", user)
	var lastError error
	for _, address := range user.Addresses {
		address.UserID = user.UID
		err := uStore.createAddress(address)
		if err != nil {
			lastError = fmt.Errorf("Error inserting address %s to user with ID %d: %v ",
				address.Value, user.UID, err)
			break
		}
	}
	if lastError != nil {
		return lastError
	}

	return nil
}

func (uStore *UserStore) checkUserExist(user *User) (bool, error) {
	var resUser User
	err := uStore.dbDriver.Get(&resUser, SELECT_USER, user.UID)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (uStore *UserStore) createAddress(address *UserAddress) error {
	_, err := uStore.dbDriver.NamedExec(INSERT_ADDRESS, address)
	if err != nil {
		return fmt.Errorf("Error inserting address %v ", err)
	}
	return nil
}

func (uStore *UserStore) GetUserByID(userID int) (*User, error) {
	var resUser User
	err := uStore.dbDriver.Get(&resUser, SELECT_USER, userID)
	if err != nil {
		return nil, UserDoesNotExistError{err: fmt.Sprintf("No user with id: %d", userID)}
	}
	resUser.Addresses, err = uStore.getUserAddresses(userID)
	if err != nil {
		return nil, fmt.Errorf("Error getting user addresses %v ", err)
	}

	return &resUser, nil
}

func (uStore *UserStore) getUserAddresses(userID int) ([]*UserAddress, error) {
	var dbAddresses []UserAddress
	err := uStore.dbDriver.Select(&dbAddresses, SELECT_USER_ADDRESSES, userID)
	if err != nil {
		return nil, fmt.Errorf("Error getting address for UID %d: %v ", userID, err)
	}

	addresses := make([]*UserAddress, 0)
	for _, addr := range dbAddresses {
		addresses = append(addresses, &addr)
	}

	return addresses, nil
}
