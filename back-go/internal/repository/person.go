package repository

import (
	"context"
	"errors"

	"back-go/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PersonRepository struct {
	db *gorm.DB
}

func NewPersonRepository(db *gorm.DB) *PersonRepository {
	return &PersonRepository{db: db}
}

func (r *PersonRepository) Create(ctx context.Context, value *model.Person) error {
	// Read database-generated identity and timestamps back into the model.
	return r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(value).Error
}

func (r *PersonRepository) List(ctx context.Context, limit, offset int) ([]model.Person, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Person{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	values := make([]model.Person, 0)
	err := r.db.WithContext(ctx).Order("id ASC").Limit(limit).Offset(offset).Find(&values).Error
	return values, total, err
}

func (r *PersonRepository) GetByID(ctx context.Context, id int64) (model.Person, error) {
	var value model.Person
	err := r.db.WithContext(ctx).First(&value, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Person{}, model.ErrPersonNotFound
	}
	return value, err
}
