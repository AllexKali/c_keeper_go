package main

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Роуты из main.go
	r.GET("/menu", getMenu)
	r.POST("/menu", addDish)
	r.DELETE("/menu/:id", deleteDish)
	r.PUT("/menu/:id", updateDish)

	return r
}

func TestGetMenu(t *testing.T) {
	initDatabase()
	router := setupRouter()

	// Создание mock-данных
	db.Create(&Menu{Name: "Test Dish", Price: 10.5, Description: "Delicious", AvailableQuantity: 5, CategoryID: 1})

	req, _ := http.NewRequest("GET", "/menu", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var menu []Menu
	err := json.Unmarshal(w.Body.Bytes(), &menu)
	assert.NoError(t, err)
	assert.NotEmpty(t, menu)
}

func TestAddDish(t *testing.T) {
	initDatabase()
	router := setupRouter()

	newDish := Menu{
		Name:              "New Dish",
		Price:             20.5,
		Description:       "Test Description",
		CategoryID:        1,
		AvailableQuantity: 10,
	}
	body, _ := json.Marshal(newDish)

	req, _ := http.NewRequest("POST", "/menu", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var dish Menu
	err := json.Unmarshal(w.Body.Bytes(), &dish)
	assert.NoError(t, err)
	assert.Equal(t, newDish.Name, dish.Name)
	assert.Equal(t, newDish.Price, dish.Price)
	assert.Equal(t, newDish.Description, dish.Description)
}

func TestDeleteDish(t *testing.T) {
	initDatabase()
	router := setupRouter()

	// Создание mock-данных
	dish := Menu{Name: "To Delete", Price: 15.0, Description: "To be deleted", AvailableQuantity: 5, CategoryID: 1}
	db.Create(&dish)

	req, _ := http.NewRequest("DELETE", "/menu/"+strconv.Itoa(int(dish.ID)), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateDish(t *testing.T) {
	initDatabase()
	router := setupRouter()

	// Создание mock-данных
	dish := Menu{Name: "To Update", Price: 12.0, Description: "Before update", AvailableQuantity: 5, CategoryID: 1}
	db.Create(&dish)

	updatedDish := Menu{
		Name:              "Updated Dish",
		Price:             18.0,
		Description:       "Updated description",
		CategoryID:        1,
		AvailableQuantity: 8,
	}
	body, _ := json.Marshal(updatedDish)

	req, _ := http.NewRequest("PUT", "/menu/"+strconv.Itoa(int(dish.ID)), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var dishResponse Menu
	err := json.Unmarshal(w.Body.Bytes(), &dishResponse)
	assert.NoError(t, err)
	assert.Equal(t, updatedDish.Name, dishResponse.Name)
	assert.Equal(t, updatedDish.Price, dishResponse.Price)
	assert.Equal(t, updatedDish.Description, dishResponse.Description)
	assert.Equal(t, updatedDish.AvailableQuantity, dishResponse.AvailableQuantity)
}
