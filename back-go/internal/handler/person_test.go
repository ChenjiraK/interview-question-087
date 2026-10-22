package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"back-go/internal/handler"
	"back-go/internal/model"
	"back-go/internal/service"

	"github.com/gin-gonic/gin"
)

type stubPersonService struct {
	create func(context.Context, service.CreatePersonInput) (service.PersonResponse, error)
	list   func(context.Context, int, int) (service.PersonListResponse, error)
	get    func(context.Context, int64) (service.PersonResponse, error)
}

func (s stubPersonService) Create(ctx context.Context, input service.CreatePersonInput) (service.PersonResponse, error) {
	return s.create(ctx, input)
}

func (s stubPersonService) List(ctx context.Context, page, limit int) (service.PersonListResponse, error) {
	return s.list(ctx, page, limit)
}

func (s stubPersonService) GetByID(ctx context.Context, id int64) (service.PersonResponse, error) {
	return s.get(ctx, id)
}

func newRequest(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, recorder
}

func assertJSON(t *testing.T, response *httptest.ResponseRecorder, status int, want any) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status = %d, want %d: %s", response.Code, status, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want JSON", got)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var gotBody, wantBody any
	if err := json.Unmarshal(response.Body.Bytes(), &gotBody); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if err := json.Unmarshal(wantJSON, &wantBody); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotBody, wantBody) {
		t.Errorf("body = %s, want %s", response.Body.String(), wantJSON)
	}
}

func assertRequestContext(t *testing.T, ctx, parent context.Context) {
	t.Helper()
	deadline, ok := ctx.Deadline()
	if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > 5*time.Second {
		t.Fatalf("service context must have a deadline within five seconds: %v", deadline)
	}
	if ctx.Value(contextKey{}) != parent.Value(contextKey{}) {
		t.Fatal("request context values were not forwarded")
	}
}

type contextKey struct{}

func TestPersonHandlerCreate(t *testing.T) {
	birthday, address := "2000-02-29", "Bangkok"
	createdAt := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name  string
		body  string
		input service.CreatePersonInput
	}{
		{"all fields", `{"firstname":" A ","lastname":"B","birthday":"2000-02-29","address":"Bangkok"}`, service.CreatePersonInput{Firstname: " A ", Lastname: "B", Birthday: &birthday, Address: &address}},
		{"omitted optional fields", `{"firstname":"A","lastname":"B"}`, service.CreatePersonInput{Firstname: "A", Lastname: "B"}},
		{"explicit nulls", `{"firstname":"A","lastname":"B","birthday":null,"address":null}`, service.CreatePersonInput{Firstname: "A", Lastname: "B"}},
		{"trailing whitespace", "{\"firstname\":\"A\",\"lastname\":\"B\"}\n\t ", service.CreatePersonInput{Firstname: "A", Lastname: "B"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			c, response := newRequest(http.MethodPost, "/person", test.body)
			parent := context.WithValue(c.Request.Context(), contextKey{}, "request-id")
			c.Request = c.Request.WithContext(parent)
			want := service.PersonResponse{ID: 42, Firstname: "A", Lastname: "B", Birthday: test.input.Birthday,
				Address: test.input.Address, CreatedAt: createdAt, UpdatedAt: createdAt}
			calls := 0
			var serviceCtx context.Context
			stub := stubPersonService{create: func(ctx context.Context, input service.CreatePersonInput) (service.PersonResponse, error) {
				calls++
				serviceCtx = ctx
				assertRequestContext(t, ctx, parent)
				if !reflect.DeepEqual(input, test.input) {
					t.Fatalf("input = %+v, want %+v", input, test.input)
				}
				return want, nil
			}}
			handler.NewPersonHandler(stub).Create(c)
			assertJSON(t, response, http.StatusCreated, want)
			if got := response.Header().Get("Location"); got != "/person/42" {
				t.Errorf("Location = %q, want /person/42", got)
			}
			if calls != 1 || serviceCtx.Err() != context.Canceled {
				t.Fatalf("calls=%d, context error=%v; want one call and canceled context", calls, serviceCtx.Err())
			}
		})
	}
}

func TestPersonHandlerCreateInvalidJSON(t *testing.T) {
	for _, test := range []struct {
		name, body string
		status     int
	}{
		{"empty", "", http.StatusBadRequest},
		{"malformed", `{`, http.StatusBadRequest},
		{"wrong field type", `{"firstname":42}`, http.StatusBadRequest},
		{"array", `[]`, http.StatusBadRequest},
		{"unknown field", `{"firstname":"A","lastname":"B","id":1}`, http.StatusBadRequest},
		{"multiple values", `{"firstname":"A","lastname":"B"} {}`, http.StatusBadRequest},
		{"trailing garbage", `{"firstname":"A","lastname":"B"} x`, http.StatusBadRequest},
		{"oversize field", `{"address":"` + strings.Repeat("x", 1<<20) + `"}`, http.StatusRequestEntityTooLarge},
		{"oversize trailing whitespace", `{}` + strings.Repeat(" ", 1<<20), http.StatusRequestEntityTooLarge},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub := stubPersonService{create: func(context.Context, service.CreatePersonInput) (service.PersonResponse, error) {
				t.Fatal("invalid JSON reached service")
				return service.PersonResponse{}, nil
			}}
			c, response := newRequest(http.MethodPost, "/person", test.body)
			handler.NewPersonHandler(stub).Create(c)
			message := "invalid JSON body or unknown fields"
			if test.status == http.StatusRequestEntityTooLarge {
				message = "request body must not exceed 1 MiB"
			}
			assertJSON(t, response, test.status, gin.H{"error": message})
			if response.Header().Get("Location") != "" {
				t.Error("failed creation must not set Location")
			}
		})
	}
}

func TestPersonHandlerList(t *testing.T) {
	for _, test := range []struct {
		name, query string
		page, limit int
	}{
		{"defaults", "", 1, 10},
		{"custom pagination", "?page=3&limit=2", 3, 2},
		{"default page", "?limit=100", 1, 100},
		{"default limit", "?page=2", 2, 10},
	} {
		t.Run(test.name, func(t *testing.T) {
			c, response := newRequest(http.MethodGet, "/persons"+test.query, "")
			want := service.PersonListResponse{Data: []service.PersonResponse{},
				Pagination: service.Pagination{Page: test.page, Limit: test.limit}}
			calls := 0
			var serviceCtx context.Context
			stub := stubPersonService{list: func(ctx context.Context, page, limit int) (service.PersonListResponse, error) {
				calls++
				serviceCtx = ctx
				assertRequestContext(t, ctx, c.Request.Context())
				if page != test.page || limit != test.limit {
					t.Fatalf("page=%d limit=%d, want page=%d limit=%d", page, limit, test.page, test.limit)
				}
				return want, nil
			}}
			handler.NewPersonHandler(stub).List(c)
			assertJSON(t, response, http.StatusOK, want)
			if calls != 1 || serviceCtx.Err() != context.Canceled {
				t.Fatalf("calls=%d, context error=%v", calls, serviceCtx.Err())
			}
		})
	}
}

func TestPersonHandlerListInvalidQuery(t *testing.T) {
	for _, query := range []string{"page=abc", "page=1.5", "page=", "page=9223372036854775808", "limit=abc", "limit=1.5", "limit=", "limit=9223372036854775808"} {
		t.Run(query, func(t *testing.T) {
			stub := stubPersonService{list: func(context.Context, int, int) (service.PersonListResponse, error) {
				t.Fatal("invalid query reached service")
				return service.PersonListResponse{}, nil
			}}
			c, response := newRequest(http.MethodGet, "/persons?"+query, "")
			handler.NewPersonHandler(stub).List(c)
			message := "page must be a positive integer"
			if strings.HasPrefix(query, "limit=") {
				message = "limit must be an integer between 1 and 100"
			}
			assertJSON(t, response, http.StatusBadRequest, gin.H{"error": message})
		})
	}
}

func TestPersonHandlerGetByID(t *testing.T) {
	c, response := newRequest(http.MethodGet, "/person/42", "")
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	want := service.PersonResponse{ID: 42, Firstname: "A", Lastname: "B"}
	calls := 0
	var serviceCtx context.Context
	stub := stubPersonService{get: func(ctx context.Context, id int64) (service.PersonResponse, error) {
		calls++
		serviceCtx = ctx
		assertRequestContext(t, ctx, c.Request.Context())
		if id != 42 {
			t.Fatalf("id=%d, want 42", id)
		}
		return want, nil
	}}
	handler.NewPersonHandler(stub).GetByID(c)
	assertJSON(t, response, http.StatusOK, want)
	if calls != 1 || serviceCtx.Err() != context.Canceled {
		t.Fatalf("calls=%d, context error=%v", calls, serviceCtx.Err())
	}
}

func TestPersonHandlerGetByIDInvalidParameter(t *testing.T) {
	for _, id := range []string{"", "abc", "1.5", "9223372036854775808"} {
		t.Run(id, func(t *testing.T) {
			stub := stubPersonService{get: func(context.Context, int64) (service.PersonResponse, error) {
				t.Fatal("invalid ID reached service")
				return service.PersonResponse{}, nil
			}}
			c, response := newRequest(http.MethodGet, "/person/"+id, "")
			c.Params = gin.Params{{Key: "id", Value: id}}
			handler.NewPersonHandler(stub).GetByID(c)
			assertJSON(t, response, http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		})
	}
}

func TestPersonHandlerServiceErrors(t *testing.T) {
	for _, test := range []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{"validation", &service.ValidationError{Message: "invalid input"}, http.StatusBadRequest, "invalid input"},
		{"wrapped validation", fmt.Errorf("create: %w", &service.ValidationError{Message: "invalid input"}), http.StatusBadRequest, "invalid input"},
		{"not found", model.ErrPersonNotFound, http.StatusNotFound, "person not found"},
		{"wrapped not found", fmt.Errorf("lookup: %w", model.ErrPersonNotFound), http.StatusNotFound, "person not found"},
		{"internal error", errors.New("private database details"), http.StatusInternalServerError, "internal server error"},
		{"deadline", context.DeadlineExceeded, http.StatusInternalServerError, "internal server error"},
		{"cancellation", context.Canceled, http.StatusInternalServerError, "internal server error"},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, operation := range []string{"create", "list", "get"} {
				t.Run(operation, func(t *testing.T) {
					calls := 0
					stub := stubPersonService{
						create: func(context.Context, service.CreatePersonInput) (service.PersonResponse, error) {
							calls++
							return service.PersonResponse{}, test.err
						},
						list: func(context.Context, int, int) (service.PersonListResponse, error) {
							calls++
							return service.PersonListResponse{}, test.err
						},
						get: func(context.Context, int64) (service.PersonResponse, error) {
							calls++
							return service.PersonResponse{}, test.err
						},
					}
					c, response := newRequest(http.MethodPost, "/person", `{"firstname":"A","lastname":"B"}`)
					c.Params = gin.Params{{Key: "id", Value: "42"}}
					h := handler.NewPersonHandler(stub)
					switch operation {
					case "create":
						h.Create(c)
					case "list":
						h.List(c)
					case "get":
						h.GetByID(c)
					}
					assertJSON(t, response, test.status, gin.H{"error": test.message})
					if calls != 1 || response.Header().Get("Location") != "" {
						t.Fatalf("calls=%d, Location=%q", calls, response.Header().Get("Location"))
					}
				})
			}
		})
	}
}
