package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func InitDB() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Lỗi khi tải file .env, sẽ sử dụng biến môi trường hệ thống")
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")

	
	if dbUser == "" || dbPassword == "" || dbHost == "" || dbName == "" {
		return nil, fmt.Errorf("thiếu biến môi trường bắt buộc: DB_USER, DB_PASSWORD, DB_HOST hoặc DB_NAME")
	}

	cfg := mysql.Config{
		User:   dbUser,
		Passwd: dbPassword,
		Net:    "tcp",
		Addr:   dbHost,
		DBName: dbName,
		Params: map[string]string{
			"parseTime": "true",
			"charset":   "utf8mb4",
			"loc":       "Local",
		},
		AllowNativePasswords: true,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối đến cơ sở dữ liệu: %v", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("lỗi kết nối cơ sở dữ liệu: %v", err)
	}

	log.Println("Đã kết nối thành công đến cơ sở dữ liệu!")
	DB = db
	return db, nil
}
