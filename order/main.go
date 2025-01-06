package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
)

// Структура для заказа
type Order struct {
	ID          uint   `gorm:"primaryKey"`
	OrderNumber uint   `json:"order_number"` // Номер заказа
	MenuID      uint   `json:"menu_id"`
	Quantity    int    `json:"quantity"`
	TableID     uint   `json:"table_id"`
	Status      string `json:"status"`
}

var db *gorm.DB

// Инициализация базы данных
func initDB() {
	var err error
	dsn := "user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("ошибка при подключении к базе данных: %v", err)
	}
	db.AutoMigrate(&Order{})
}

func fetchDishDetails(menuID uint) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://localhost:5003/menu/%d", menuID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка соединения с menu: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("menu вернул статус: %d", resp.StatusCode)
	}

	var menuItem map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&menuItem); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	return menuItem, nil
}

// Создание заказа
func createOrder(c *gin.Context) {
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем заказ в базу данных
	order.Status = "В процессе"
	if err := db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Заказ успешно создан",
		"order":   order,
	})
}

// Получение всех заказов
func getOrders(c *gin.Context) {
	var orders []Order
	if err := db.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// Получение конкретного заказа по ID
func getOrder(c *gin.Context) {
	orderID := c.Param("id")
	var order Order
	if err := db.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Заказ не найден"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// Обновление статуса заказа
func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	var order Order
	if err := db.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Заказ не найден"})
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order.Status = input.Status
	if err := db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус заказа обновлен", "order": order})
}

// Удаление заказа
func deleteOrder(c *gin.Context) {
	orderID := c.Param("id")
	if err := db.Delete(&Order{}, orderID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Заказ удален"})
}

// Получение описания блюда по ID заказа
func getDishDescriptionByOrderID(c *gin.Context) {
	// Извлекаем ID заказа из параметров запроса
	orderID := c.Param("id")
	var order Order
	if err := db.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Заказ не найден"})
		return
	}

	// Получаем данные блюда из сервиса menu
	menuItem, err := fetchDishDetails(order.MenuID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось получить данные о блюде: %v", err)})
		return
	}

	// Извлекаем описание блюда
	description, ok := menuItem["description"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось извлечь описание блюда"})
		return
	}

	// Возвращаем описание блюда
	c.JSON(http.StatusOK, gin.H{
		"order_id":    order.ID,
		"menu_id":     order.MenuID,
		"description": description,
	})
}

func main() {
	initDB()

	r := gin.Default()

	// CRUD-операции для заказов
	r.POST("/order", createOrder)

	r.GET("/order/:id/description", getDishDescriptionByOrderID)

	r.GET("/orders", getOrders)
	r.GET("/order/:id", getOrder)
	r.PUT("/order/:id/status", UpdateOrderStatus)
	r.DELETE("/order/:id", deleteOrder)

	r.Run(":5004") // сервис будет доступен на порту 5004
}
