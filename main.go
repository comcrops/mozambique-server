package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type user struct {
	id        string
	username  string
	packCount string
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	connectionString := loadDatabaseUrl()

	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		log.Fatalf("Error while connecting to db")
	}

	rows, err := db.Query("SELECT * FROM user_packs")

	if err != nil {
		log.Fatalf("Error reading the db")
	}
	defer rows.Close()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		var res string
		for rows.Next() {
			rows.Scan(&res)
			log.Println(res)

		}
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run()
}

func loadDatabaseUrl() string {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	connectionString := os.Getenv("DATABASE_URL")

	if connectionString == "" {
		log.Fatalf("Connection string wasn't found or is empty!")
	}
	return connectionString
}

func increase() {

}

