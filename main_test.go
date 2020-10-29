package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"goMinimalAPI/store"
	"log"
	"testing"
	"time"

	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"goMinimalAPI/api"
)

type Case struct {
	Name    string
	Method  string
	Path    string
	Headers map[string]string
	Status  int
	Result  interface{}
	Body    interface{}
}

func prepareDB(DSN string) error {
	db, err := sqlx.Connect("postgres", DSN)
	if err != nil {
		return fmt.Errorf("Error connecting to DB: %v ", err)
	}
	_, err = db.Exec("TRUNCATE addresses, users;")
	if err != nil {
		return err
	}
	err = db.Close()
	if err != nil {
		return err
	}
	return nil
}

func TestUserCreation(t *testing.T) {
	err := prepareDB(DSN)
	if err != nil {
		t.Fatalf("Error preparing DB: %v", err)
	}
	usersStore, err := store.NewUserStore(DSN)
	if err != nil {
		log.Printf("Error creating UsersStore: %v", err)
		return
	}

	apiMux, err := api.NewUserAPIMux(usersStore)
	if err != nil {
		panic(err)
	}
	ts := httptest.NewServer(apiMux)

	testUser1 := &store.User{
		UID:       1,
		FirstName: "TestName",
		LastName:  "LastName",
		Addresses: make([]*store.UserAddress, 2),
	}
	testUser1.Addresses[0] = &store.UserAddress{
		Value: "Address1",
	}
	testUser1.Addresses[1] = &store.UserAddress{
		Value: "User Address2 value",
	}
	testUser2 := &store.User{
		UID:       2,
		FirstName: "TestName2",
		LastName:  "LastName2",
		Addresses: make([]*store.UserAddress, 0),
	}

	cases := []Case{
		{
			Name:   "Test connection: not found root",
			Method: "GET",
			Path:   "/",
			Status: http.StatusNotFound,
		},
		{
			Name:   "Creating user",
			Path:   "/api/user/create",
			Method: "POST",
			Body:   testUser1,
			Status: http.StatusOK,
		},
		{
			Name:   "Creating user with existing id",
			Path:   "/api/user/create",
			Method: "POST",
			Body:   testUser1,
			Status: http.StatusBadRequest,
		},
		{
			Name:   "Creating user no addresses",
			Path:   "/api/user/create",
			Method: "POST",
			Body:   testUser2,
			Status: http.StatusOK,
		},

		{
			Name:    "Get user1",
			Path:    "/api/user/addresses",
			Result:  testUser1,
			Headers: map[string]string{"USER_ID": "1"},
			Status:  http.StatusOK,
		},
		{
			Name:    "Get user2",
			Path:    "/api/user/addresses",
			Result:  testUser2,
			Headers: map[string]string{"USER_ID": "2"},
			Status:  http.StatusOK,
		},
		{
			Name:    "Get not existed user",
			Path:    "/api/user/addresses",
			Result:  testUser1,
			Headers: map[string]string{"USER_ID": "3"},
			Status:  http.StatusNotFound,
		},
	}

	runCases(t, ts, cases)
	err = usersStore.CloseDB()
	if err != nil {
		log.Printf("Error closing DB: %v\n", err)
	}
}

func runCases(t *testing.T, ts *httptest.Server, cases []Case) {
	var (
		client = &http.Client{Timeout: 100 * time.Second}
	)
	for idx, item := range cases {
		var (
			err error
			req *http.Request
		)

		caseName := fmt.Sprintf("case %d: [%s] %s %s", idx, item.Method, item.Path, item.Headers)

		if item.Method == "" || item.Method == http.MethodGet {
			req, err = http.NewRequest(item.Method, ts.URL+item.Path, nil)
			if err == nil {
				t.Fatalf("case %d: Error preparing request: %v", idx, err)
			}
			if req == nil {
				t.Fatalf("case %d: prepared requast is nill", idx)
			} else {
				if item.Headers != nil {
					for hName, hValue := range item.Headers {
						req.Header.Set(hName, hValue)
					}
				}
			}

		} else {
			data, err := json.Marshal(item.Body)
			if err != nil {
				panic(err)
			}
			reqBody := bytes.NewReader(data)
			req, err = http.NewRequest(item.Method, ts.URL+item.Path, reqBody)
			if err == nil {
				t.Fatalf("case %d: Error preparing request: %v", idx, err)
			}
			if req == nil {
				t.Fatalf("case %d: prepared requast is nill", idx)
			} else {
				if item.Headers != nil {
					req.Header.Add("Content-Type", "application/json")
				}
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("[%s] request error: %v", caseName, err)

		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		if item.Status == 0 {
			item.Status = http.StatusOK
		}

		if resp.StatusCode != item.Status {
			t.Fatalf("[%s] expected http status %v, got %v", caseName, item.Status, resp.StatusCode)

		}

		if (item.Status != http.StatusOK) || (len(body) == 0 && item.Result == nil) {
			continue
		}

		fmt.Println(string(body))
		resultUser := &store.User{}
		err = json.Unmarshal(body, &resultUser)
		if err != nil {
			fmt.Println(string(body))
			t.Fatalf("[%s] cant unpack json: %v", caseName, err)
		}
		expected := item.Result
		expectedUser := expected.(*store.User)
		if resultUser.FirstName != expectedUser.FirstName {
			t.Fatalf("[%s] expected first name %s, got %s", caseName, expectedUser.FirstName, resultUser.FirstName)
		}
	}

}
