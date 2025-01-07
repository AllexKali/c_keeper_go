package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

// Модели для таблиц
type Category struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
}

type Menu struct {
	ID                uint     `gorm:"primaryKey" json:"id"`
	Name              string   `json:"name"`
	Price             float64  `json:"price"`
	Description       string   `json:"description"`
	CategoryID        uint     `json:"category_id"`
	AvailableQuantity int      `json:"available_quantity"`
	Category          Category `gorm:"foreignKey:CategoryID;references:ID" json:"category"` // связь с таблицей categories
}

// Глобальная переменная для работы с базой данных
var db *gorm.DB

func initDatabase() {
	var err error
	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Автоматическая миграция схемы базы данных
	err = db.AutoMigrate(&Category{}, &Menu{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("Database connected and migrated successfully")
}

func main() {
	initDatabase()

	r := gin.Default()
	r.Use(cors.Default())

	// CRUD-операции
	r.GET("/menu", getMenu)
	r.GET("/menu/:id", getDishByID) // Добавлен эндпоинт для получения блюда по ID
	r.POST("/menu", addDish)
	r.DELETE("/menu/:id", deleteDish)
	r.PUT("/menu/:id", updateDish)

	if err := r.Run(":5003"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Получение всего меню
func getMenu(c *gin.Context) {
	var menu []Menu
	if err := db.Preload("Category").Find(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, menu)
}

// Получение блюда по ID
func getDishByID(c *gin.Context) {
	id := c.Param("id")
	var dish Menu
	if err := db.Preload("Category").First(&dish, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dish not found"})
		return
	}
	c.JSON(http.StatusOK, dish)
}

// Добавление блюда в меню
func addDish(c *gin.Context) {
	var dish Menu
	if err := c.ShouldBindJSON(&dish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&dish).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dish)
}

// Удаление блюда из меню
func deleteDish(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := db.Delete(&Menu{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Dish deleted successfully"})
}

// Обновление блюда в меню
func updateDish(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedDish Menu
	if err := c.ShouldBindJSON(&updatedDish); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingDish Menu
	if err := db.First(&existingDish, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dish not found"})
		return
	}

	existingDish.Name = updatedDish.Name
	existingDish.Price = updatedDish.Price
	existingDish.Description = updatedDish.Description
	existingDish.CategoryID = updatedDish.CategoryID
	existingDish.AvailableQuantity = updatedDish.AvailableQuantity

	if err := db.Save(&existingDish).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existingDish)
}
