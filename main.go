package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type UserPacks struct {
	id        string
	Username  string
	PackCount int
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
	defer db.Close()

	r := gin.Default()
	r.GET("/:username", func(ctx *gin.Context) {
		increase(ctx, db)
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

func increase(c *gin.Context, db *sql.DB) {
	username := c.Param("username")
	row := db.QueryRow("SELECT * FROM user_packs WHERE username=$1", username)

	var res UserPacks
	err := row.Scan(&res.id, &res.Username, &res.PackCount)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No rows were returned for username: %s", username)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		log.Fatalf("Error scanning row: %v", err)
	}

	log.Println(res)
	c.JSON(http.StatusOK, res)
}
