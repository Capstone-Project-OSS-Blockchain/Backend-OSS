package connections

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
)

var DB *gorm.DB

func Connect() {
	cek := godotenv.Load()
	if cek != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	conn := os.Getenv("DB_CONNECTION")
	db := os.Getenv("DB_DATABASE")
	username := os.Getenv("DB_USERNAME")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")

	var err error
	DB, err = gorm.Open(conn, username+":"+pass+"@tcp("+host+":"+port+")/"+db+"?parseTime=True")
	if err != nil {
		panic("Failed to connect to MySQL database")
	}
}
