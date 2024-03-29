package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type UserPacks struct {
	id        string
	Username  string
	PackCount uint
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	connectionString := loadDatabaseUrl()

	db, err := sql.Open("postgres", connectionString)

	if err != nil {
		log.Printf("Error while connecting to db")
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

	r.GET("/reset/:username", func(ctx *gin.Context) {
		resetCount(ctx, db)
	})

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"https://comcrops.at", "https://mozambique.comcrops.at"}
	r.Use(cors.New(config))

	r.RunTLS("comcrops.at:8181", "/etc/letsencrypt/live/comcrops.at/cert.pem", "/etc/letsencrypt/live/comcrops.at/privkey.pem")
}

func loadDatabaseUrl() string {
	err := godotenv.Load()

	if err != nil {
		log.Printf("Error loading .env file")
	}

	connectionString := os.Getenv("DATABASE_URL")

	if connectionString == "" {
		log.Printf("Connection string wasn't found or is empty!")
	}
	return connectionString
}

func resetCount(c *gin.Context, db *sql.DB) {
	username := c.Param("username")
	user, err := getUserPackByUsername(username, db)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Error increasing row: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found try creating it first by getting it's score before increasing."})
		}

		log.Printf("Error scanning row: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error"})
		return
	}

	updateUserScore(-int(user.PackCount), c, db)
}

func get(c *gin.Context, db *sql.DB) {
	username := c.Param("username")
	user, err := getUserPackByUsername(username, db)
	if err != nil {
		if err == sql.ErrNoRows {
			createdUser, creationError := createUser(username, db)

			if creationError != nil {
				log.Printf("Error creating user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error while creating new user"})
				return
			}
			log.Printf("Created new user: %s", username)
			user = createdUser
		} else {
			log.Printf("Error scanning row: %v", err)
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
			log.Printf("Error changing row: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found try creating it first by getting it's score before changing it."})
			return
		}

		log.Printf("Error scanning row: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server error"})
		return
	}

	user, err = updateScore(username, (uint)(max(0, (int)(user.PackCount)+change)), db)
	if err != nil {
		log.Printf("Error creating user: %v", err)
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
