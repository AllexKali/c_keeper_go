package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Устанавливаем тестовую базу данных (в памяти)
func initTestDB() *gorm.DB {
	dsn := "user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("ошибка при подключении к базе данных для теста")
	}
	// Создаем таблицу для заказов
	db.AutoMigrate(&Order{})
	return db
}

// Тестирование создания заказа
func TestCreateOrder(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.POST("/order", createOrder)

	order := Order{
		OrderNumber: 1,
		MenuID:      1,
		Quantity:    2,
		TableID:     1,
	}

	// Отправляем POST-запрос для создания заказа
	orderJSON, _ := json.Marshal(order)
	req, _ := http.NewRequest("POST", "/order", bytes.NewBuffer(orderJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Проверяем код ответа
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Проверяем, что заказ был создан
	assert.Equal(t, "Заказ успешно создан", response["message"])
	assert.NotNil(t, response["order"])
}

// Тестирование получения всех заказов
func TestGetOrders(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.GET("/orders", getOrders)

	// Отправляем GET-запрос для получения всех заказов
	req, _ := http.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Проверяем код ответа
	assert.Equal(t, http.StatusOK, w.Code)

	var orders []Order
	json.Unmarshal(w.Body.Bytes(), &orders)

	// Проверяем, что заказы получены
	assert.Greater(t, len(orders), 0)
}

// Тестирование получения заказа по ID
func TestGetOrder(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.GET("/order/:id", getOrder)

	// Создаем заказ для теста
	order := Order{
		OrderNumber: 1,
		MenuID:      1,
		Quantity:    2,
		TableID:     1,
		Status:      "В процессе",
	}
	db.Create(&order)

	// Отправляем GET-запрос для получения заказа по ID
	req, _ := http.NewRequest("GET", fmt.Sprintf("/order/%d", order.ID), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Проверяем код ответа
	assert.Equal(t, http.StatusOK, w.Code)

	var response Order
	json.Unmarshal(w.Body.Bytes(), &response)

	// Проверяем, что заказ получен
	assert.Equal(t, order.ID, response.ID)
}

// Тестирование обновления статуса заказа
func TestUpdateOrderStatus(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.PUT("/order/:id/status", UpdateOrderStatus)

	// Создаем заказ для теста
	order := Order{
		OrderNumber: 1,
		MenuID:      1,
		Quantity:    2,
		TableID:     1,
		Status:      "В процессе",
	}
	db.Create(&order)

	// Отправляем PUT-запрос для обновления статуса
	status := map[string]string{"status": "Завершен"}
	statusJSON, _ := json.Marshal(status)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/order/%d/status", order.ID), bytes.NewBuffer(statusJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Проверяем код ответа
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Проверяем, что статус обновлен
	assert.Equal(t, "Статус заказа обновлен", response["message"])
	assert.Equal(t, "Завершен", response["order"].(map[string]interface{})["status"])
}

// Тестирование удаления заказа
func TestDeleteOrder(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.DELETE("/order/:id", deleteOrder)

	// Создаем заказ для теста
	order := Order{
		OrderNumber: 1,
		MenuID:      1,
		Quantity:    2,
		TableID:     1,
		Status:      "В процессе",
	}
	db.Create(&order)

	// Отправляем DELETE-запрос для удаления заказа
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/order/%d", order.ID), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Проверяем код ответа
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Проверяем, что заказ удален
	assert.Equal(t, "Заказ удален", response["message"])

	// Проверяем, что заказ действительно удален
	var checkOrder Order
	err := db.First(&checkOrder, order.ID).Error
	assert.Error(t, err) // Ожидаем ошибку, так как заказ должен быть удален
}
