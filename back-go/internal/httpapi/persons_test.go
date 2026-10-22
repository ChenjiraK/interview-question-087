package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"back-go/internal/model"
	"back-go/internal/service"

	"github.com/gin-gonic/gin"
)

type stubPersonRepository struct {
	create func(context.Context, *model.Person) error
	list   func(context.Context, int, int) ([]model.Person, int64, error)
	get    func(context.Context, int64) (model.Person, error)
}

func (s stubPersonRepository) Create(ctx context.Context, params *model.Person) error {
	return s.create(ctx, params)
}

func (s stubPersonRepository) List(ctx context.Context, limit, offset int) ([]model.Person, int64, error) {
	return s.list(ctx, limit, offset)
}

func (s stubPersonRepository) GetByID(ctx context.Context, id int64) (model.Person, error) {
	return s.get(ctx, id)
}

func request(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, req)
	return response
}

func TestCreatePerson(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, optional := range []bool{true, false} {
		t.Run(fmt.Sprintf("optional_fields_%t", optional), func(t *testing.T) {
			body := `{"firstname":" สมชาย ","lastname":"ใจดี"}`
			if optional {
				body = `{"firstname":" สมชาย ","lastname":"ใจดี","birthday":"1995-01-15","address":"Bangkok"}`
			}
			store := stubPersonRepository{create: func(ctx context.Context, params *model.Person) error {
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("database operation has no deadline")
				}
				if params.Firstname != "สมชาย" || params.Lastname != "ใจดี" {
					t.Fatalf("unexpected names: %+v", params)
				}
				value := model.Person{ID: 42, Firstname: params.Firstname, Lastname: params.Lastname, Address: params.Address,
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
			response := request(NewRouter(service.NewPersonService(store)), http.MethodPost, "/person", body)
			if response.Code != http.StatusCreated || response.Header().Get("Location") != "/person/42" {
				t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
			}
			var value service.PersonResponse
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

func TestCreatePersonValidation(t *testing.T) {
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
	store := stubPersonRepository{create: func(context.Context, *model.Person) error {
		t.Fatal("invalid input reached the database")
		return nil
	}}
	router := NewRouter(service.NewPersonService(store))
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := request(router, http.MethodPost, "/person", test.body)
			if response.Code != test.code {
				t.Fatalf("got %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestGetPerson(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := stubPersonRepository{get: func(_ context.Context, id int64) (model.Person, error) {
		if id == 42 {
			return model.Person{ID: id, Firstname: "A", Lastname: "B"}, nil
		}
		return model.Person{}, model.ErrPersonNotFound
	}}
	router := NewRouter(service.NewPersonService(store))
	for _, test := range []struct {
		id   string
		code int
	}{{"42", 200}, {"43", 404}, {"0", 400}, {"-1", 400}, {"abc", 400}, {"9223372036854775808", 400}} {
		t.Run(test.id, func(t *testing.T) {
			response := request(router, http.MethodGet, "/person/"+test.id, "")
			if response.Code != test.code {
				t.Fatalf("got %d: %s", response.Code, response.Body.String())
			}
			if test.code == 200 && !strings.Contains(response.Body.String(), `"id":42`) {
				t.Fatal("wrong person returned")
			}
			if test.code == 404 && response.Body.String() != `{"error":"person not found"}` {
				t.Fatalf("unexpected not-found response: %s", response.Body.String())
			}
		})
	}
}

func TestPersonRouteNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(nil)
	for _, test := range []struct{ method, path string }{
		{http.MethodPost, "/user"},
		{http.MethodGet, "/user"},
		{http.MethodGet, "/user/1"},
		{http.MethodGet, "/person"},
		{http.MethodPost, "/persons"},
		{http.MethodGet, "/persons/1"},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			response := request(router, test.method, test.path, "")
			if response.Code != http.StatusNotFound || response.Body.String() != `{"error":"route not found"}` {
				t.Fatalf("unexpected route match: %d %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestListPersons(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name, query string
		page, limit int
		total       int64
		ids         []int64
		totalPages  int64
	}{
		{"defaults", "", 1, 10, 2, []int64{1, 2}, 1},
		{"empty", "", 1, 10, 0, nil, 0},
		{"second page", "?page=2&limit=2", 2, 2, 5, []int64{3, 4}, 3},
		{"last page", "?page=3&limit=2", 3, 2, 5, []int64{5}, 3},
		{"beyond last page", "?page=4&limit=2", 4, 2, 5, nil, 3},
		{"exact pages", "?limit=2", 1, 2, 4, []int64{1, 2}, 2},
		{"maximum limit", "?limit=100", 1, 100, 1, []int64{1}, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := stubPersonRepository{list: func(ctx context.Context, limit, offset int) ([]model.Person, int64, error) {
				if _, ok := ctx.Deadline(); !ok {
					t.Fatal("database operation has no deadline")
				}
				if limit != test.limit || offset != (test.page-1)*test.limit {
					t.Fatalf("unexpected pagination: limit=%d offset=%d", limit, offset)
				}
				var values []model.Person
				for _, id := range test.ids {
					values = append(values, model.Person{ID: id})
				}
				return values, test.total, nil
			}}
			response := request(NewRouter(service.NewPersonService(store)), http.MethodGet, "/persons"+test.query, "")
			if response.Code != 200 {
				t.Fatalf("got %d: %s", response.Code, response.Body.String())
			}
			var result service.PersonListResponse
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			want := service.Pagination{Page: test.page, Limit: test.limit, Total: test.total, TotalPages: test.totalPages}
			if result.Pagination != want || result.Data == nil || len(result.Data) != len(test.ids) {
				t.Fatalf("unexpected response: %s", response.Body.String())
			}
			for i, id := range test.ids {
				if result.Data[i].ID != id {
					t.Fatalf("unexpected person: %+v", result.Data[i])
				}
			}
		})
	}
}

func TestListPersonsValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := stubPersonRepository{list: func(context.Context, int, int) ([]model.Person, int64, error) {
		t.Fatal("invalid pagination reached the database")
		return nil, 0, nil
	}}
	router := NewRouter(service.NewPersonService(store))
	for _, query := range []string{
		"page=0", "page=-1", "page=abc", "page=1.5", "page=", "page=9223372036854775808",
		"page=" + strconv.Itoa(int(^uint(0)>>1)) + "&limit=100",
		"limit=0", "limit=-1", "limit=101", "limit=abc", "limit=1.5", "limit=", "limit=9223372036854775808",
	} {
		t.Run(query, func(t *testing.T) {
			response := request(router, http.MethodGet, "/persons?"+query, "")
			if response.Code != http.StatusBadRequest {
				t.Fatalf("got %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestPersonDatabaseErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dbError := errors.New("private database connection details")
	store := stubPersonRepository{
		create: func(context.Context, *model.Person) error { return dbError },
		list:   func(context.Context, int, int) ([]model.Person, int64, error) { return nil, 0, dbError },
		get:    func(context.Context, int64) (model.Person, error) { return model.Person{}, dbError },
	}
	router := NewRouter(service.NewPersonService(store))
	for _, test := range []struct{ method, path, body string }{
		{http.MethodPost, "/person", `{"firstname":"A","lastname":"B"}`},
		{http.MethodGet, "/persons", ""},
		{http.MethodGet, "/person/1", ""},
	} {
		response := request(router, test.method, test.path, test.body)
		if response.Code != 500 || response.Body.String() != `{"error":"internal server error"}` {
			t.Fatalf("unexpected error response: %d %s", response.Code, response.Body.String())
		}
	}
}
