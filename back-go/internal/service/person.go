package service

import (
	"context"
	"strings"
	"time"

	"back-go/internal/model"
)

type PersonRepository interface {
	Create(context.Context, *model.Person) error
	List(context.Context, int, int) ([]model.Person, int64, error)
	GetByID(context.Context, int64) (model.Person, error)
}

type CreatePersonInput struct {
	Firstname string  `json:"firstname"`
	Lastname  string  `json:"lastname"`
	Birthday  *string `json:"birthday"`
	Address   *string `json:"address"`
}

type PersonResponse struct {
	ID        int64     `json:"id"`
	Firstname string    `json:"firstname"`
	Lastname  string    `json:"lastname"`
	Birthday  *string   `json:"birthday"`
	Address   *string   `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

type PersonListResponse struct {
	Data       []PersonResponse `json:"data"`
	Pagination Pagination       `json:"pagination"`
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

type PersonService struct {
	repository PersonRepository
}

func NewPersonService(repository PersonRepository) *PersonService {
	return &PersonService{repository: repository}
}

func (s *PersonService) Create(ctx context.Context, input CreatePersonInput) (PersonResponse, error) {
	value := model.Person{
		Firstname: strings.TrimSpace(input.Firstname),
		Lastname:  strings.TrimSpace(input.Lastname),
		Address:   input.Address,
	}
	if value.Firstname == "" || value.Lastname == "" {
		return PersonResponse{}, &ValidationError{Message: "firstname and lastname are required"}
	}
	if input.Birthday != nil {
		birthday, err := time.Parse(time.DateOnly, *input.Birthday)
		if err != nil {
			return PersonResponse{}, &ValidationError{Message: "birthday must be a valid date in YYYY-MM-DD format"}
		}
		value.Birthday = &birthday
	}
	if err := s.repository.Create(ctx, &value); err != nil {
		return PersonResponse{}, err
	}
	return toPersonResponse(value), nil
}

func (s *PersonService) List(ctx context.Context, page, limit int) (PersonListResponse, error) {
	if page <= 0 {
		return PersonListResponse{}, &ValidationError{Message: "page must be a positive integer"}
	}
	if limit < 1 || limit > 100 {
		return PersonListResponse{}, &ValidationError{Message: "limit must be an integer between 1 and 100"}
	}
	// GORM accepts an int offset; reject pages whose offset would overflow.
	if page-1 > int(^uint(0)>>1)/limit {
		return PersonListResponse{}, &ValidationError{Message: "page is too large"}
	}
	values, total, err := s.repository.List(ctx, limit, (page-1)*limit)
	if err != nil {
		return PersonListResponse{}, err
	}
	responses := make([]PersonResponse, 0, len(values))
	for _, value := range values {
		responses = append(responses, toPersonResponse(value))
	}
	totalPages := total / int64(limit)
	if total%int64(limit) != 0 {
		totalPages++
	}
	return PersonListResponse{
		Data:       responses,
		Pagination: Pagination{Page: page, Limit: limit, Total: total, TotalPages: totalPages},
	}, nil
}

func (s *PersonService) GetByID(ctx context.Context, id int64) (PersonResponse, error) {
	if id <= 0 {
		return PersonResponse{}, &ValidationError{Message: "id must be a positive integer"}
	}
	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return PersonResponse{}, err
	}
	return toPersonResponse(value), nil
}

func toPersonResponse(value model.Person) PersonResponse {
	response := PersonResponse{
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
