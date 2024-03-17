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
	PackCount uint
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
	r.GET("/increment/:username", func(ctx *gin.Context) {
		updateUserScore(1, ctx, db)
	})

	r.GET("/decrement/:username", func(ctx *gin.Context) {
		updateUserScore(-1, ctx, db)
	})

	r.GET("/:username", func(ctx *gin.Context) {
		get(ctx, db)
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

func get(c *gin.Context, db *sql.DB) {
	username := c.Param("username")
	user, err := getUserPackByUsername(username, db)
	if err != nil {
		if err == sql.ErrNoRows {
			createdUser, creationError := createUser(username, db)

			if creationError != nil {
				log.Fatalf("Error creating user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error while creating new user"})
				return
			}
			log.Printf("Created new user: %s", username)
			user = createdUser
		} else {
			log.Fatalf("Error scanning row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error"})
			return
		}
	}

	c.JSON(http.StatusOK, user)
}

func updateUserScore(change int, c *gin.Context, db *sql.DB) {
	username := c.Param("username")
	user, err := getUserPackByUsername(username, db)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Fatalf("Error increasing row: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found try creating it first by getting it's score before increasing."})
		}

		log.Fatalf("Error scanning row: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error"})
		return
	}

	user, err = updateScore(username, (uint)(max(0, (int)(user.PackCount)+change)), db)
	if err != nil {
		log.Fatalf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error"})
		return
	}

	c.JSON(http.StatusOK, user)

}

func createUser(username string, db *sql.DB) (UserPacks, error) {
	_, err := db.Exec("INSERT INTO user_packs (username, pack_count) VALUES ($1, $2)", username, 0)
	if err != nil {
		return UserPacks{}, err
	}

	return getUserPackByUsername(username, db)
}

func updateScore(username string, newCount uint, db *sql.DB) (UserPacks, error) {
	_, err := db.Exec("UPDATE user_packs SET pack_count=$1 WHERE username=$2", newCount, username)

	if err != nil {
		return UserPacks{}, err
	}
	return getUserPackByUsername(username, db)
}

func getUserPackByUsername(username string, db *sql.DB) (UserPacks, error) {
	row := db.QueryRow("SELECT * FROM user_packs WHERE username=$1", username)

	var res UserPacks
	err := row.Scan(&res.id, &res.Username, &res.PackCount)

	if err != nil {
		return UserPacks{}, err
	}

	return res, nil
}
