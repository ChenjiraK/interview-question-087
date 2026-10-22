package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"back-go/internal/model"
	"back-go/internal/service"

	"github.com/gin-gonic/gin"
)

type stubUserRepository struct {
	create func(context.Context, *model.User) error
	list   func(context.Context) ([]model.User, error)
	get    func(context.Context, int64) (model.User, error)
}

func (s stubUserRepository) Create(ctx context.Context, params *model.User) error {
	return s.create(ctx, params)
}

func (s stubUserRepository) List(ctx context.Context) ([]model.User, error) { return s.list(ctx) }

func (s stubUserRepository) GetByID(ctx context.Context, id int64) (model.User, error) {
	return s.get(ctx, id)
}

func request(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, req)
	return response
}

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, optional := range []bool{true, false} {
		t.Run(fmt.Sprintf("optional_fields_%t", optional), func(t *testing.T) {
			body := `{"firstname":" สมชาย ","lastname":"ใจดี"}`
			if optional {
				body = `{"firstname":" สมชาย ","lastname":"ใจดี","birthday":"1995-01-15","address":"Bangkok"}`
			}
			store := stubUserRepository{create: func(ctx context.Context, params *model.User) error {
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("database operation has no deadline")
				}
				if params.Firstname != "สมชาย" || params.Lastname != "ใจดี" {
					t.Fatalf("unexpected names: %+v", params)
				}
				value := model.User{ID: 42, Firstname: params.Firstname, Lastname: params.Lastname, Address: params.Address,
					CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
				if optional {
					if params.Birthday == nil || params.Birthday.Format(time.DateOnly) != "1995-01-15" || params.Address == nil || *params.Address != "Bangkok" {
						t.Fatalf("unexpected optional fields: %+v", params)
					}
					value.Birthday = params.Birthday
				} else if params.Birthday != nil || params.Address != nil {
					t.Fatal("omitted fields should be NULL")
				}
				*params = value
				return nil
			}}
			response := request(NewRouter(service.NewUserService(store)), http.MethodPost, "/user", body)
			if response.Code != http.StatusCreated || response.Header().Get("Location") != "/user/42" {
				t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
			}
			var value service.UserResponse
			if err := json.Unmarshal(response.Body.Bytes(), &value); err != nil {
				t.Fatal(err)
			}
			if value.ID != 42 || value.CreatedAt.IsZero() || value.UpdatedAt.IsZero() {
				t.Fatalf("missing generated fields: %+v", value)
			}
			if !optional && (!strings.Contains(response.Body.String(), `"birthday":null`) || !strings.Contains(response.Body.String(), `"address":null`)) {
				t.Fatalf("nullable fields must be JSON null: %s", response.Body.String())
			}
		})
	}
}

func TestCreateUserValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		body string
		code int
	}{
		{"empty body", "", 400},
		{"malformed JSON", `{`, 400},
		{"multiple JSON values", `{"firstname":"A","lastname":"B"} {}`, 400},
		{"missing firstname", `{"lastname":"B"}`, 400},
		{"missing lastname", `{"firstname":"A"}`, 400},
		{"blank name", `{"firstname":"  ","lastname":"B"}`, 400},
		{"invalid name type", `{"firstname":1,"lastname":"B"}`, 400},
		{"invalid date", `{"firstname":"A","lastname":"B","birthday":"2025-02-29"}`, 400},
		{"date with time", `{"firstname":"A","lastname":"B","birthday":"2000-01-01T00:00:00Z"}`, 400},
		{"unknown field", `{"firstname":"A","lastname":"B","id":1}`, 400},
		{"large body", `{"firstname":"A","lastname":"B","address":"` + strings.Repeat("x", 1<<20) + `"}`, 413},
	}
	store := stubUserRepository{create: func(context.Context, *model.User) error {
		t.Fatal("invalid input reached the database")
		return nil
	}}
	router := NewRouter(service.NewUserService(store))
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := request(router, http.MethodPost, "/user", test.body)
			if response.Code != test.code {
				t.Fatalf("got %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := stubUserRepository{get: func(_ context.Context, id int64) (model.User, error) {
		if id == 42 {
			return model.User{ID: id, Firstname: "A", Lastname: "B"}, nil
		}
		return model.User{}, model.ErrUserNotFound
	}}
	router := NewRouter(service.NewUserService(store))
	for _, test := range []struct {
		id   string
		code int
	}{{"42", 200}, {"43", 404}, {"0", 400}, {"-1", 400}, {"abc", 400}, {"9223372036854775808", 400}} {
		t.Run(test.id, func(t *testing.T) {
			response := request(router, http.MethodGet, "/user/"+test.id, "")
			if response.Code != test.code {
				t.Fatalf("got %d: %s", response.Code, response.Body.String())
			}
			if test.code == 200 && !strings.Contains(response.Body.String(), `"id":42`) {
				t.Fatal("wrong user returned")
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, empty := range []bool{true, false} {
		store := stubUserRepository{list: func(context.Context) ([]model.User, error) {
			if empty {
				return nil, nil
			}
			return []model.User{{ID: 1}, {ID: 2}}, nil
		}}
		response := request(NewRouter(service.NewUserService(store)), http.MethodGet, "/user", "")
		if response.Code != 200 {
			t.Fatalf("got %d: %s", response.Code, response.Body.String())
		}
		if empty && response.Body.String() != "[]" {
			t.Fatalf("empty response should be [], got %s", response.Body.String())
		}
		var values []service.UserResponse
		if err := json.Unmarshal(response.Body.Bytes(), &values); err != nil {
			t.Fatal(err)
		}
		if !empty && (len(values) != 2 || values[0].ID != 1 || values[1].ID != 2) {
			t.Fatalf("unexpected list: %+v", values)
		}
	}
}

func TestUserDatabaseErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbError := errors.New("private database connection details")
	store := stubUserRepository{
		create: func(context.Context, *model.User) error { return dbError },
		list:   func(context.Context) ([]model.User, error) { return nil, dbError },
		get:    func(context.Context, int64) (model.User, error) { return model.User{}, dbError },
	}
	router := NewRouter(service.NewUserService(store))
	for _, test := range []struct{ method, path, body string }{
		{http.MethodPost, "/user", `{"firstname":"A","lastname":"B"}`},
		{http.MethodGet, "/user", ""},
		{http.MethodGet, "/user/1", ""},
	} {
		response := request(router, test.method, test.path, test.body)
		if response.Code != 500 || response.Body.String() != `{"error":"internal server error"}` {
			t.Fatalf("unexpected error response: %d %s", response.Code, response.Body.String())
		}
	}
}
