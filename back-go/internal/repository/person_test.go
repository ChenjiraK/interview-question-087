package repository_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"back-go/internal/model"
	"back-go/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	insertPersonSQL = `INSERT INTO "public"."person" ("firstname","lastname","birthday","address") VALUES ($1,$2,$3,$4) RETURNING *`
	countPersonsSQL = `SELECT count(*) FROM "public"."person"`
	getPersonSQL    = `SELECT * FROM "public"."person" WHERE "person"."id" = $1 ORDER BY "person"."id" LIMIT $2`
)

var personColumns = []string{"id", "firstname", "lastname", "birthday", "address", "created_at", "updated_at"}

func newRepository(t *testing.T) (*repository.PersonRepository, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		mock.ExpectClose()
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close mock database: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet SQL expectations: %v", err)
		}
	})
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	return repository.NewPersonRepository(db), mock
}

func personFixture() model.Person {
	birthday := time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC)
	address := "Bangkok"
	createdAt := time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	return model.Person{ID: 42, Firstname: "สมชาย ' OR 1=1 --", Lastname: "ใจดี", Birthday: &birthday,
		Address: &address, CreatedAt: createdAt, UpdatedAt: createdAt.Add(time.Hour)}
}

func personRows(value model.Person) *sqlmock.Rows {
	var birthday, address any
	if value.Birthday != nil {
		birthday = *value.Birthday
	}
	if value.Address != nil {
		address = *value.Address
	}
	return sqlmock.NewRows(personColumns).AddRow(value.ID, value.Firstname, value.Lastname, birthday, address, value.CreatedAt, value.UpdatedAt)
}

func TestPersonRepositoryCreate(t *testing.T) {
	for _, nullable := range []bool{false, true} {
		t.Run(fmt.Sprintf("nullable_fields_%t", nullable), func(t *testing.T) {
			repo, mock := newRepository(t)
			want := personFixture()
			var birthday, address any
			if nullable {
				want.Birthday, want.Address = nil, nil
			} else {
				birthday, address = *want.Birthday, *want.Address
			}
			mock.ExpectBegin()
			mock.ExpectQuery(insertPersonSQL).WithArgs(want.Firstname, want.Lastname, birthday, address).
				WillReturnRows(personRows(want))
			mock.ExpectCommit()
			value := model.Person{Firstname: want.Firstname, Lastname: want.Lastname, Birthday: want.Birthday, Address: want.Address}
			if err := repo.Create(context.Background(), &value); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(value, want) {
				t.Fatalf("persisted model = %+v, want %+v including database-generated ID and timestamps", value, want)
			}
		})
	}
}

func TestPersonRepositoryCreateErrors(t *testing.T) {
	for _, stage := range []string{"begin", "insert", "commit"} {
		t.Run(stage, func(t *testing.T) {
			repo, mock := newRepository(t)
			wantErr := errors.New(stage + " failed")
			value := model.Person{Firstname: "A", Lastname: "B"}
			switch stage {
			case "begin":
				mock.ExpectBegin().WillReturnError(wantErr)
			case "insert":
				mock.ExpectBegin()
				mock.ExpectQuery(insertPersonSQL).WithArgs("A", "B", nil, nil).WillReturnError(wantErr)
				mock.ExpectRollback()
			case "commit":
				mock.ExpectBegin()
				mock.ExpectQuery(insertPersonSQL).WithArgs("A", "B", nil, nil).
					WillReturnRows(personRows(model.Person{ID: 42, Firstname: "A", Lastname: "B"}))
				mock.ExpectCommit().WillReturnError(wantErr)
			}
			if err := repo.Create(context.Background(), &value); !errors.Is(err, wantErr) {
				t.Fatalf("error = %v, want %v", err, wantErr)
			}
		})
	}
}

func TestPersonRepositoryList(t *testing.T) {
	person := personFixture()
	nullablePerson := person
	nullablePerson.Birthday, nullablePerson.Address = nil, nil
	for _, test := range []struct {
		name   string
		limit  int
		offset int
		total  int64
		value  *model.Person
	}{
		{"first page with nullable fields", 10, 0, 1, &nullablePerson},
		{"second page with all fields", 2, 2, 5, &person},
		{"empty database", 10, 0, 0, nil},
		{"beyond last page", 2, 10, 5, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := newRepository(t)
			mock.ExpectQuery(countPersonsSQL).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(test.total))
			query := `SELECT * FROM "public"."person" ORDER BY id ASC LIMIT $1`
			rows := sqlmock.NewRows(personColumns)
			if test.value != nil {
				rows = personRows(*test.value)
			}
			if test.offset > 0 {
				mock.ExpectQuery(query+` OFFSET $2`).WithArgs(test.limit, test.offset).WillReturnRows(rows)
			} else {
				mock.ExpectQuery(query).WithArgs(test.limit).WillReturnRows(rows)
			}
			values, total, err := repo.List(context.Background(), test.limit, test.offset)
			if err != nil || total != test.total || values == nil {
				t.Fatalf("List = %+v, %d, %v; want non-nil slice and total %d", values, total, err, test.total)
			}
			want := []model.Person{}
			if test.value != nil {
				want = append(want, *test.value)
			}
			if !reflect.DeepEqual(values, want) {
				t.Fatalf("people = %+v, want %+v", values, want)
			}
		})
	}
}

func TestPersonRepositoryListCountError(t *testing.T) {
	repo, mock := newRepository(t)
	wantErr := errors.New("count failed")
	mock.ExpectQuery(countPersonsSQL).WillReturnError(wantErr)
	values, total, err := repo.List(context.Background(), 10, 0)
	if !errors.Is(err, wantErr) || values != nil || total != 0 {
		t.Fatalf("List = %+v, %d, %v; want nil, 0, %v", values, total, err, wantErr)
	}
}

func TestPersonRepositoryListQueryError(t *testing.T) {
	repo, mock := newRepository(t)
	wantErr := errors.New("select failed")
	mock.ExpectQuery(countPersonsSQL).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT * FROM "public"."person" ORDER BY id ASC LIMIT $1 OFFSET $2`).
		WithArgs(2, 2).WillReturnError(wantErr)
	_, total, err := repo.List(context.Background(), 2, 2)
	if !errors.Is(err, wantErr) || total != 5 {
		t.Fatalf("total=%d error=%v, want total=5 error=%v", total, err, wantErr)
	}
}

func TestPersonRepositoryGetByID(t *testing.T) {
	for _, nullable := range []bool{false, true} {
		t.Run(fmt.Sprintf("nullable_fields_%t", nullable), func(t *testing.T) {
			repo, mock := newRepository(t)
			want := personFixture()
			if nullable {
				want.Birthday, want.Address = nil, nil
			}
			mock.ExpectQuery(getPersonSQL).WithArgs(want.ID, 1).WillReturnRows(personRows(want))
			got, err := repo.GetByID(context.Background(), want.ID)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("GetByID = %+v, %v; want %+v", got, err, want)
			}
		})
	}
}

func TestPersonRepositoryGetByIDErrors(t *testing.T) {
	dbError := errors.New("select failed")
	for _, test := range []struct {
		name    string
		dbError error
		wantErr error
	}{
		{"no rows", nil, model.ErrPersonNotFound},
		{"wrapped record not found", fmt.Errorf("query: %w", gorm.ErrRecordNotFound), model.ErrPersonNotFound},
		{"database error", dbError, dbError},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := newRepository(t)
			query := mock.ExpectQuery(getPersonSQL).WithArgs(42, 1)
			if test.dbError == nil {
				query.WillReturnRows(sqlmock.NewRows(personColumns))
			} else {
				query.WillReturnError(test.dbError)
			}
			value, err := repo.GetByID(context.Background(), 42)
			if !errors.Is(err, test.wantErr) || !reflect.DeepEqual(value, model.Person{}) {
				t.Fatalf("GetByID = %+v, %v; want zero model and %v", value, err, test.wantErr)
			}
		})
	}
}

func TestPersonRepositoryCanceledContext(t *testing.T) {
	for _, operation := range []string{"create", "list", "get"} {
		t.Run(operation, func(t *testing.T) {
			repo, _ := newRepository(t)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			var err error
			switch operation {
			case "create":
				err = repo.Create(ctx, &model.Person{Firstname: "A", Lastname: "B"})
			case "list":
				_, _, err = repo.List(ctx, 10, 0)
			case "get":
				_, err = repo.GetByID(ctx, 42)
			}
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("error = %v, want context.Canceled", err)
			}
		})
	}
}
