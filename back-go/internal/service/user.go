package service

import (
	"context"
	"strings"
	"time"

	"back-go/internal/model"
)

type UserRepository interface {
	Create(context.Context, *model.User) error
	List(context.Context) ([]model.User, error)
	GetByID(context.Context, int64) (model.User, error)
}

type CreateUserInput struct {
	Firstname string  `json:"firstname"`
	Lastname  string  `json:"lastname"`
	Birthday  *string `json:"birthday"`
	Address   *string `json:"address"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Birthday  *string   `json:"birthday"`
	Address   *string   `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (UserResponse, error) {
	value := model.User{
		Firstname: strings.TrimSpace(input.Firstname),
		Lastname:  strings.TrimSpace(input.Lastname),
		Address:   input.Address,
	}
	if value.Firstname == "" || value.Lastname == "" {
		return UserResponse{}, &ValidationError{Message: "firstname and lastname are required"}
	}
	if input.Birthday != nil {
		birthday, err := time.Parse(time.DateOnly, *input.Birthday)
		if err != nil {
			return UserResponse{}, &ValidationError{Message: "birthday must be a valid date in YYYY-MM-DD format"}
		}
		value.Birthday = &birthday
	}
	if err := s.repository.Create(ctx, &value); err != nil {
		return UserResponse{}, err
	}
	return toUserResponse(value), nil
}

func (s *UserService) List(ctx context.Context) ([]UserResponse, error) {
	values, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}
	responses := make([]UserResponse, 0, len(values))
	for _, value := range values {
		responses = append(responses, toUserResponse(value))
	}
	return responses, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (UserResponse, error) {
	if id <= 0 {
		return UserResponse{}, &ValidationError{Message: "id must be a positive integer"}
	}
	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return UserResponse{}, err
	}
	return toUserResponse(value), nil
}

func toUserResponse(value model.User) UserResponse {
	response := UserResponse{
		ID:        value.ID,
		Firstname: value.Firstname,
		Lastname:  value.Lastname,
		Address:   value.Address,
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
	}
	if value.Birthday != nil {
		date := value.Birthday.Format(time.DateOnly)
		response.Birthday = &date
	}
	return response
}
