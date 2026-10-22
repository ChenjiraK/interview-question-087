package model_test

import (
	"sync"
	"testing"

	"back-go/internal/model"

	"gorm.io/gorm/schema"
)

func TestPersonSchema(t *testing.T) {
	parsed, err := schema.Parse(&model.Person{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Table != "public.person" {
		t.Errorf("table = %q, want public.person", parsed.Table)
	}
	if len(parsed.Fields) != 7 {
		t.Errorf("field count = %d, want 7", len(parsed.Fields))
	}
	for _, test := range []struct {
		field, column string
		notNull       bool
	}{
		{"ID", "id", false},
		{"Firstname", "firstname", true},
		{"Lastname", "lastname", true},
		{"Birthday", "birthday", false},
		{"Address", "address", false},
		{"CreatedAt", "created_at", true},
		{"UpdatedAt", "updated_at", true},
	} {
		t.Run(test.field, func(t *testing.T) {
			field := parsed.LookUpField(test.field)
			if field == nil {
				t.Fatalf("missing schema field %s", test.field)
			}
			if field.DBName != test.column || field.NotNull != test.notNull {
				t.Errorf("column=%q notNull=%t, want column=%q notNull=%t", field.DBName, field.NotNull, test.column, test.notNull)
			}
		})
	}
	if len(parsed.PrimaryFields) != 1 || parsed.PrioritizedPrimaryField == nil || parsed.PrioritizedPrimaryField.Name != "ID" || !parsed.PrioritizedPrimaryField.AutoIncrement {
		t.Error("ID must be the sole auto-increment primary key")
	}
	if field := parsed.LookUpField("Birthday"); field.DataType != "date" {
		t.Errorf("birthday data type = %q, want date", field.DataType)
	}
	for _, name := range []string{"CreatedAt", "UpdatedAt"} {
		field := parsed.LookUpField(name)
		if field.DataType != "timestamptz" || !field.HasDefaultValue || field.DefaultValue != "CURRENT_TIMESTAMP" {
			t.Errorf("%s must use database-generated TIMESTAMPTZ default CURRENT_TIMESTAMP", name)
		}
		if field.AutoCreateTime != 0 || field.AutoUpdateTime != 0 {
			t.Errorf("%s must not be overwritten by GORM timestamp callbacks", name)
		}
	}
}
