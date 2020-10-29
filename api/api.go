package api

import (
	"encoding/json"
	"fmt"
	"goMinimalAPI/store"
	"log"
	"net/http"
	"strconv"
)

type UserAPIError struct {
	Error string `json:"error"`
}

type UserAPI struct {
	usersStore *store.UserStore
}

type AddressesResponse struct {
	UID       int
	Addresses []string
}

func NewUserAPIMux(usersStore *store.UserStore) (*http.ServeMux, error) {
	userAPI := &UserAPI{usersStore: usersStore}
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/user/create", userAPI.CreateUser)
	apiMux.HandleFunc("/api/user/addresses", userAPI.GetUserAddresses)
	return apiMux, nil
}

func (api *UserAPI) CreateUser(w http.ResponseWriter, r *http.Request) {
	userRecord := &store.User{}
	err := json.NewDecoder(r.Body).Decode(userRecord)
	log.Printf("Error decoding input JSON: %v ", err)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		apiError := &UserAPIError{Error: "Error parsing request input"}
		err = json.NewEncoder(w).Encode(apiError)
		if err != nil {
			log.Printf("Error encoding respinse error on user creation (%s): %v", apiError.Error, err)
		}
		return
	}

	err = api.usersStore.CreateUser(userRecord)

	if err != nil {
		log.Printf("Error creating user record: %s ", err)
		apiError := &UserAPIError{}
		switch err.(type) {
		case store.UserExistError:
			w.WriteHeader(http.StatusBadRequest)
			apiError.Error = fmt.Sprintf("User with id %d already exist", userRecord.UID)
			break
		default:
			w.WriteHeader(http.StatusInternalServerError)
			apiError.Error = "Error creating user"
		}

		err = json.NewEncoder(w).Encode(apiError)
		if err != nil {
			log.Printf("Error encoding respinse error on user creation after DB (%s): %v", apiError.Error, err)
		}
		return
	}
}

func (api *UserAPI) GetUserAddresses(w http.ResponseWriter, r *http.Request) {
	uidStr := r.Header.Get("USER_ID")
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		apiError := &UserAPIError{Error: "Wrong USER_ID value"}
		err = json.NewEncoder(w).Encode(apiError)
		if err != nil {
			log.Printf("[GetUserAddresses] Error encoding respinse error on parsing user id (%s): %v",
				apiError.Error, err)
		}
		return
	}

	user, err := api.usersStore.GetUserByID(uid)
	if err != nil {
		log.Printf("Error getting user record: %v ", err)
		apiError := &UserAPIError{}
		switch err.(type) {
		case store.UserDoesNotExistError:
			w.WriteHeader(http.StatusNotFound)
			apiError.Error = fmt.Sprintf("User with id %d does not exist", uid)
		default:
			w.WriteHeader(http.StatusInternalServerError)
			apiError.Error = "Error getting user data"
		}
		err = json.NewEncoder(w).Encode(apiError)
		if err != nil {
			log.Printf("[GetUserAddresses] Error encoding response error on user get after DB (%s): %v",
				apiError.Error, err)
		}
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Printf("[GetUserAddresses] Error encoding user response (%v): %v",
			user, err)
	}
}
