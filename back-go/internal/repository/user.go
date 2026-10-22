package repository

import (
	"context"
	"errors"

	"back-go/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, value *model.User) error {
	// Read database-generated identity and timestamps back into the model.
	return r.db.WithContext(ctx).Clauses(clause.Returning{}).Create(value).Error
}

func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	values := make([]model.User, 0)
	err := r.db.WithContext(ctx).Order("id ASC").Find(&values).Error
	return values, err
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (model.User, error) {
	var value model.User
	err := r.db.WithContext(ctx).First(&value, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, model.ErrUserNotFound
	}
	return value, err
}
