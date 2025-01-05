package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// Модель для заказов
type Order struct {
	ID          uint   `gorm:"primaryKey"`
	OrderNumber uint   `json:"order_number"` // Это поле будет извлекаться из JSON
	MenuID      uint   `json:"menu_id"`
	Quantity    int    `json:"quantity"`
	TableID     uint   `json:"table_id"`
	Status      string `json:"status"`
}

// Инициализация базы данных
func initDatabase() {
	var err error
	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Автоматическая миграция схемы базы данных
	err = db.AutoMigrate(&Order{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("Database connected and migrated successfully")
}

func main() {
	initDatabase()

	r := gin.Default()

	// Роуты для обработки заказов
	r.GET("/orders", getOrders)
	r.POST("/orders", createOrder)
	r.GET("/orders/:id", getOrder)
	r.PUT("/orders/:id", updateOrder)
	r.DELETE("/orders/:id", deleteOrder)

	if err := r.Run(":5002"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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

// Создание нового заказа
func createOrder(c *gin.Context) {
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем статус по умолчанию
	order.Status = "В ожидании"

	// Добавляем новый заказ в базу данных
	if err := db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully", "order": order})
}

// Получение одного заказа по ID
func getOrder(c *gin.Context) {
	id := c.Param("id")
	var order Order
	if err := db.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// Обновление заказа
func updateOrder(c *gin.Context) {
	id := c.Param("id")
	var order Order
	if err := db.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order updated successfully", "order": order})
}

// Удаление заказа
func deleteOrder(c *gin.Context) {
	id := c.Param("id")
	if err := db.Delete(&Order{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}
