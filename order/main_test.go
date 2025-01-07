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

	orderJSON, _ := json.Marshal(order)
	req, _ := http.NewRequest("POST", "/order", bytes.NewBuffer(orderJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Заказ успешно создан", response["message"])
	assert.NotNil(t, response["order"])
}

// Тестирование получения всех заказов
func TestGetOrders(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.GET("/orders", getOrders)

	req, _ := http.NewRequest("GET", "/orders", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var orders []Order
	json.Unmarshal(w.Body.Bytes(), &orders)

	assert.Greater(t, len(orders), 0)
}

// Тестирование получения заказа по ID
func TestGetOrder(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.GET("/order/:id", getOrder)

	order := Order{
		OrderNumber: 1,
		MenuID:      1,
		Quantity:    2,
		TableID:     1,
		Status:      "В процессе",
	}
	db.Create(&order)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/order/%d", order.ID), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Order
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, order.ID, response.ID)
}

// Тестирование получения заказа по несуществующему ID
func TestGetOrderNotFound(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.GET("/order/:id", getOrder)

	req, _ := http.NewRequest("GET", "/order/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Заказ не найден", response["error"])
}

// Тестирование обновления статуса заказа
func TestUpdateOrderStatus(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.PUT("/order/:id/status", UpdateOrderStatus)

	order := Order{
		OrderNumber: 1,
		MenuID:      1,
		Quantity:    2,
		TableID:     1,
		Status:      "В процессе",
	}
	db.Create(&order)

	status := map[string]string{"status": "Завершен"}
	statusJSON, _ := json.Marshal(status)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/order/%d/status", order.ID), bytes.NewBuffer(statusJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Equal(t, "Статус заказа обновлен", response["message"])
	assert.Equal(t, "Завершен", response["order"].(map[string]interface{})["status"])
}

func TestUpdateOrderStatusInvalid(t *testing.T) {
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

	// Отправляем PUT-запрос с некорректным статусом
	invalidStatus := map[string]string{"status": "Неизвестный"}
	invalidStatusJSON, _ := json.Marshal(invalidStatus)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/order/%d/status", order.ID), bytes.NewBuffer(invalidStatusJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Проверяем код ответа
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Проверяем сообщение об ошибке
	assert.Equal(t, "Некорректный статус", response["error"])
}

func TestDeleteOrder(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.DELETE("/order/:id", deleteOrder)

	// Создаем новый заказ
	order := Order{ /* заполните тестовые данные */ }
	db.Create(&order)

	// Удаляем заказ
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/order/%d", order.ID), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Проверяем код ответа
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Проверяем сообщение об успешном удалении
	assert.Equal(t, "Заказ успешно удалён", response["message"])
}

func TestDeleteOrderNotFound(t *testing.T) {
	db = initTestDB()
	r := gin.Default()
	r.DELETE("/order/:id", deleteOrder)

	// Отправляем DELETE-запрос для несуществующего заказа
	req, _ := http.NewRequest("DELETE", "/order/9999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Проверяем код ответа
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Проверяем сообщение об ошибке
	assert.Equal(t, "Заказ не найден", response["error"])
}

// Тест на случай, когда заказ не найден
func TestGetDishDescriptionByOrderID_OrderNotFound(t *testing.T) {
	// Инициализация тестовой базы данных
	_ = initTestDB()
	r := gin.Default()
	r.GET("/order/:id/description", getDishDescriptionByOrderID)

	// Отправка запроса для несуществующего заказа
	req, _ := http.NewRequest("GET", "/order/999/description", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Проверка кода ответа и сообщения
	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Заказ не найден", response["error"])
}

func TestGetDishDescriptionByOrderID_MenuServiceError(t *testing.T) {
	// Имитируем ошибку получения данных меню (например, заказ не найден)
	req, _ := http.NewRequest("GET", "/order/999/description", nil)
	rec := httptest.NewRecorder()
	handler := gin.New()
	handler.GET("/order/:id/description", getDishDescriptionByOrderID)

	handler.ServeHTTP(rec, req)

	// Ожидаем ошибку 404, так как заказ не найден
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "Заказ не найден")
}

func TestGetDishDescriptionByOrderID_InvalidMenuData(t *testing.T) {
	// Имитируем ошибку с пустыми данными блюда (например, заказ не найден)
	req, _ := http.NewRequest("GET", "/order/999/description", nil)
	rec := httptest.NewRecorder()
	handler := gin.New()
	handler.GET("/order/:id/description", getDishDescriptionByOrderID)

	// Обрабатываем запрос
	handler.ServeHTTP(rec, req)

	// Ожидаем ошибку 404, так как заказ не найден
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "Заказ не найден")
}

func TestGetDishDescriptionByOrderID_Success(t *testing.T) {
	// Создаем mock-запрос, который будет возвращать корректные данные
	req, _ := http.NewRequest("GET", "/order/214/description", nil)
	rec := httptest.NewRecorder()
	handler := gin.New()
	handler.GET("/order/:id/description", getDishDescriptionByOrderID)

	handler.ServeHTTP(rec, req)

	// Проверяем статус ответа и содержимое
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Классический борщ с мясом и сметаной")
	assert.Contains(t, rec.Body.String(), "120.5")
}

func TestGetDishDescriptionByOrderID_NotFound(t *testing.T) {
	// Отправляем запрос с несуществующим order_id
	req, _ := http.NewRequest("GET", "/order/999/description", nil)
	rec := httptest.NewRecorder()
	handler := gin.New()
	handler.GET("/order/:id/description", getDishDescriptionByOrderID)

	handler.ServeHTTP(rec, req)

	// Проверяем, что вернулся статус 404 и корректное сообщение
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "Заказ не найден")
}
