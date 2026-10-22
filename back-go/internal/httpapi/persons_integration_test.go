package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"back-go/internal/database"
	"back-go/internal/repository"
	"back-go/internal/service"

	"github.com/gin-gonic/gin"
)

// TEST_DATABASE_URL must point to a separate, migrated PostgreSQL test database.
func TestPersonsPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	t.Setenv("DATABASE_URL", dsn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := database.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	gin.SetMode(gin.TestMode)
	router := NewRouter(service.NewPersonService(repository.NewPersonRepository(db)))

	response := request(router, http.MethodGet, "/persons", "")
	if response.Code != 200 || response.Body.String() != `{"data":[],"pagination":{"page":1,"limit":10,"total":0,"total_pages":0}}` {
		t.Fatalf("test database must start empty: %d %s", response.Code, response.Body.String())
	}
	var created []int64
	defer func() {
		for _, id := range created {
			if _, err := sqlDB.Exec(`DELETE FROM public."person" WHERE id = $1`, id); err != nil {
				t.Errorf("clean up person %d: %v", id, err)
			}
		}
	}()

	body := `{"firstname":"สมชาย ' OR 1=1 --","lastname":"ใจดี","birthday":"1995-01-15","address":"กรุงเทพมหานคร"}`
	response = request(router, http.MethodPost, "/person", body)
	if response.Code != 201 {
		t.Fatalf("create failed: %d %s", response.Code, response.Body.String())
	}
	var first service.PersonResponse
	if err := json.Unmarshal(response.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	created = append(created, first.ID)
	if first.ID <= 0 || first.Firstname != "สมชาย ' OR 1=1 --" || first.Birthday == nil || *first.Birthday != "1995-01-15" || first.Address == nil || *first.Address != "กรุงเทพมหานคร" || first.CreatedAt.IsZero() || first.UpdatedAt.IsZero() {
		t.Fatalf("unexpected stored person: %+v", first)
	}
	var count int
	if err := sqlDB.QueryRow(`SELECT count(*) FROM public."person" WHERE id = $1 AND firstname = $2`, first.ID, first.Firstname).Scan(&count); err != nil || count != 1 {
		t.Fatalf("person was not persisted: count=%d err=%v", count, err)
	}
	var createdAt, updatedAt time.Time
	if err := sqlDB.QueryRow(`SELECT created_at, updated_at FROM public."person" WHERE id = $1`, first.ID).Scan(&createdAt, &updatedAt); err != nil {
		t.Fatal(err)
	}
	if !first.CreatedAt.Equal(createdAt) || !first.UpdatedAt.Equal(updatedAt) {
		t.Fatal("response timestamps must match the database-generated values")
	}
	response = request(router, http.MethodGet, fmt.Sprintf("/person/%d", first.ID), "")
	if response.Code != 200 {
		t.Fatalf("get failed: %d %s", response.Code, response.Body.String())
	}
	var fetched service.PersonResponse
	if err := json.Unmarshal(response.Body.Bytes(), &fetched); err != nil || fetched.ID != first.ID || fetched.Firstname != first.Firstname || fetched.Birthday == nil || *fetched.Birthday != *first.Birthday || !fetched.CreatedAt.Equal(first.CreatedAt) || !fetched.UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("wrong person: %s, err=%v", response.Body.String(), err)
	}
	response = request(router, http.MethodPost, "/person", `{"firstname":"Optional","lastname":"Fields","birthday":null,"address":null}`)
	if response.Code != 201 {
		t.Fatalf("create nullable fields failed: %d %s", response.Code, response.Body.String())
	}
	var second service.PersonResponse
	if err := json.Unmarshal(response.Body.Bytes(), &second); err != nil {
		t.Fatal(err)
	}
	created = append(created, second.ID)
	if second.ID <= first.ID || second.Birthday != nil || second.Address != nil {
		t.Fatalf("invalid generated ID or nullable fields: %+v", second)
	}
	response = request(router, http.MethodGet, "/persons", "")
	var result service.PersonListResponse
	if response.Code != 200 {
		t.Fatalf("list failed: %d %s", response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil || len(result.Data) != 2 || result.Data[0].ID != first.ID || result.Data[1].ID != second.ID || result.Pagination.Total != 2 || result.Pagination.TotalPages != 1 {
		t.Fatalf("unexpected person list: %s, err=%v", response.Body.String(), err)
	}
	for _, test := range []struct {
		page int
		ids  []int64
	}{
		{1, []int64{first.ID}},
		{2, []int64{second.ID}},
		{3, nil},
	} {
		response = request(router, http.MethodGet, fmt.Sprintf("/persons?page=%d&limit=1", test.page), "")
		var page service.PersonListResponse
		if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil || response.Code != 200 {
			t.Fatalf("pagination failed: %d %s, err=%v", response.Code, response.Body.String(), err)
		}
		if page.Data == nil || len(page.Data) != len(test.ids) || page.Pagination != (service.Pagination{Page: test.page, Limit: 1, Total: 2, TotalPages: 2}) {
			t.Fatalf("unexpected page: %s", response.Body.String())
		}
		for i, id := range test.ids {
			if page.Data[i].ID != id {
				t.Fatalf("unexpected person: %+v", page.Data[i])
			}
		}
	}
	response = request(router, http.MethodGet, fmt.Sprintf("/person/%d", second.ID+1), "")
	if response.Code != 404 {
		t.Fatalf("missing person should return 404, got %d", response.Code)
	}
}
