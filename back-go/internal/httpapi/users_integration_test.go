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
func TestUsersPostgres(t *testing.T) {
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
	router := NewRouter(service.NewUserService(repository.NewUserRepository(db)))

	response := request(router, http.MethodGet, "/user", "")
	if response.Code != 200 || response.Body.String() != "[]" {
		t.Fatalf("test database must start empty: %d %s", response.Code, response.Body.String())
	}
	var created []int64
	defer func() {
		for _, id := range created {
			if _, err := sqlDB.Exec(`DELETE FROM public."user" WHERE id = $1`, id); err != nil {
				t.Errorf("clean up user %d: %v", id, err)
			}
		}
	}()

	body := `{"firstname":"สมชาย ' OR 1=1 --","lastname":"ใจดี","birthday":"1995-01-15","address":"กรุงเทพมหานคร"}`
	response = request(router, http.MethodPost, "/user", body)
	if response.Code != 201 {
		t.Fatalf("create failed: %d %s", response.Code, response.Body.String())
	}
	var first service.UserResponse
	if err := json.Unmarshal(response.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	created = append(created, first.ID)
	if first.ID <= 0 || first.Firstname != "สมชาย ' OR 1=1 --" || first.Birthday == nil || *first.Birthday != "1995-01-15" || first.Address == nil || *first.Address != "กรุงเทพมหานคร" || first.CreatedAt.IsZero() || first.UpdatedAt.IsZero() {
		t.Fatalf("unexpected stored user: %+v", first)
	}
	var count int
	if err := sqlDB.QueryRow(`SELECT count(*) FROM public."user" WHERE id = $1 AND firstname = $2`, first.ID, first.Firstname).Scan(&count); err != nil || count != 1 {
		t.Fatalf("user was not persisted: count=%d err=%v", count, err)
	}
	var createdAt, updatedAt time.Time
	if err := sqlDB.QueryRow(`SELECT created_at, updated_at FROM public."user" WHERE id = $1`, first.ID).Scan(&createdAt, &updatedAt); err != nil {
		t.Fatal(err)
	}
	if !first.CreatedAt.Equal(createdAt) || !first.UpdatedAt.Equal(updatedAt) {
		t.Fatal("response timestamps must match the database-generated values")
	}
	response = request(router, http.MethodGet, fmt.Sprintf("/user/%d", first.ID), "")
	if response.Code != 200 {
		t.Fatalf("get failed: %d %s", response.Code, response.Body.String())
	}
	var fetched service.UserResponse
	if err := json.Unmarshal(response.Body.Bytes(), &fetched); err != nil || fetched.ID != first.ID || fetched.Firstname != first.Firstname || fetched.Birthday == nil || *fetched.Birthday != *first.Birthday || !fetched.CreatedAt.Equal(first.CreatedAt) || !fetched.UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("wrong user: %s, err=%v", response.Body.String(), err)
	}
	response = request(router, http.MethodPost, "/user", `{"firstname":"Optional","lastname":"Fields","birthday":null,"address":null}`)
	if response.Code != 201 {
		t.Fatalf("create nullable fields failed: %d %s", response.Code, response.Body.String())
	}
	var second service.UserResponse
	if err := json.Unmarshal(response.Body.Bytes(), &second); err != nil {
		t.Fatal(err)
	}
	created = append(created, second.ID)
	if second.ID <= first.ID || second.Birthday != nil || second.Address != nil {
		t.Fatalf("invalid generated ID or nullable fields: %+v", second)
	}
	response = request(router, http.MethodGet, "/user", "")
	var values []service.UserResponse
	if response.Code != 200 {
		t.Fatalf("list failed: %d %s", response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &values); err != nil || len(values) != 2 || values[0].ID != first.ID || values[1].ID != second.ID {
		t.Fatalf("unexpected user list: %s, err=%v", response.Body.String(), err)
	}
	response = request(router, http.MethodGet, fmt.Sprintf("/user/%d", second.ID+1), "")
	if response.Code != 404 {
		t.Fatalf("missing user should return 404, got %d", response.Code)
	}
}
