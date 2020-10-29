package store

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"testing"
)

var DSN = "user=postgres password=testPWD99 host=localhost dbname=usersapi sslmode=disable"

func cleanDB(db *sqlx.DB) error {
	_, err := db.Exec("TRUNCATE addresses, users;")
	if err != nil {
		return fmt.Errorf("Error truncating tables: %v ", err)
	}
	return nil
}

func TestUserStore_CreateUser(t *testing.T) {
	db, err := sqlx.Connect("postgres", DSN)
	if err != nil {
		t.Fatalf("Error connecting to DB: %v ", err)
	}
	err = cleanDB(db)
	if err != nil {
		t.Fatalf("Error cleaning DB: %v", err)
	}
	type fields struct {
		dbDriver *sqlx.DB
	}
	type args struct {
		user *User
	}
	testUser1 := &User{
		UID:       1,
		FirstName: "TestName",
		LastName:  "LastName",
		Addresses: make([]*UserAddress, 0),
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Create user",
			fields:  fields{},
			args:    args{user: testUser1},
			wantErr: false,
		},
	}
	uStore := &UserStore{
		dbDriver: db,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := uStore.CreateUser(tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	err = cleanDB(db)
	if err != nil {
		t.Fatalf("Error cleaning DB: %v", err)
	}
	err = db.Close()
	if err != nil {
		t.Fatalf("Error closing DB: %v", err)
	}
}

func TestUserStore_createAddress(t *testing.T) {
	db, err := sqlx.Connect("postgres", DSN)
	if err != nil {
		t.Fatalf("Error connecting to DB: %v ", err)
	}
	err = cleanDB(db)
	if err != nil {
		t.Fatalf("Error cleaning DB: %v", err)
	}
	type fields struct {
		dbDriver *sqlx.DB
	}
	type args struct {
		address *UserAddress
	}
	testUser1 := &User{
		UID:       1,
		FirstName: "TestName",
		LastName:  "LastName",
		Addresses: make([]*UserAddress, 0),
	}
	testAddress := &UserAddress{
		UserID: 1,
		Value:  "Test addr",
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Create address",
			fields:  fields{},
			args:    args{address: testAddress},
			wantErr: false,
		},
	}
	uStore := &UserStore{
		dbDriver: db,
	}
	err = uStore.CreateUser(testUser1)
	if err != nil {
		t.Fatalf("Error inserting test user: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := uStore.createAddress(tt.args.address); (err != nil) != tt.wantErr {
				t.Errorf("createAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	err = cleanDB(db)
	if err != nil {
		t.Fatalf("Error cleaning DB: %v", err)
	}
	err = db.Close()
	if err != nil {
		t.Fatalf("Error closing DB: %v", err)
	}
}
