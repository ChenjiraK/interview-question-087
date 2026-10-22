package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"back-go/internal/model"
	"back-go/internal/service"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	Create(context.Context, service.CreateUserInput) (service.UserResponse, error)
	List(context.Context) ([]service.UserResponse, error)
	GetByID(context.Context, int64) (service.UserResponse, error)
}

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(c *gin.Context) {
	var input service.CreateUserInput
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		invalidJSON(c, err)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		invalidJSON(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	value, err := h.service.Create(ctx, input)
	if err != nil {
		userError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/user/%d", value.ID))
	c.JSON(http.StatusCreated, value)
}

func (h *UserHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	values, err := h.service.List(ctx)
	if err != nil {
		userError(c, err)
		return
	}
	c.JSON(http.StatusOK, values)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a positive integer"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	value, err := h.service.GetByID(ctx, id)
	if err != nil {
		userError(c, err)
		return
	}
	c.JSON(http.StatusOK, value)
}

func invalidJSON(c *gin.Context, err error) {
	var sizeError *http.MaxBytesError
	if errors.As(err, &sizeError) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body must not exceed 1 MiB"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body or unknown fields"})
}

func userError(c *gin.Context, err error) {
	var validationError *service.ValidationError
	if errors.As(err, &validationError) {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
		return
	}
	if errors.Is(err, model.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	log.Printf("user request failed: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
