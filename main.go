package main

import (
	"backend/database"
	"backend/handlers"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {

	var err error // Khai báo 1 lần duy nhất

	// Gọi đúng hàm InitDB từ package database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Không thể khởi tạo cơ sở dữ liệu: %v", err)
	}
	defer db.Close()

	fmt.Println("Kết nối cơ sở dữ liệu thành công!")

	if err != nil {
		log.Fatalf("Không thể khởi tạo cơ sở dữ liệu: %v", err)
	}
	defer db.Close()

	// Thiết lập router
	router := mux.NewRouter()

	// Router API cho User
	router.HandleFunc("/api/users", handlers.CreateUser).Methods("POST")
	router.HandleFunc("/api/users/{id}", handlers.GetUser).Methods("GET")
	router.HandleFunc("/api/users/{id}", handlers.UpdateUser).Methods("PUT")
	router.HandleFunc("/api/users/{id}", handlers.DeleteUser).Methods("DELETE")
	router.HandleFunc("/api/users/{user_id}/tasks", handlers.GetUserTasks).Methods("GET")
	router.HandleFunc("/api/users/{user_id}/categories", handlers.GetUserCategories).Methods("GET")
	router.HandleFunc("/api/users/login", handlers.Login).Methods("POST")

	// Router API cho Task
	router.HandleFunc("/api/tasks", handlers.CreateTask).Methods("POST")
	router.HandleFunc("/api/tasks/{id}", handlers.GetTask).Methods("GET")
	router.HandleFunc("/api/tasks/{id}", handlers.UpdateTask).Methods("PUT")
	router.HandleFunc("/api/tasks/{id}", handlers.DeleteTask).Methods("DELETE")

	// Router API cho Category
	router.HandleFunc("/api/categories", handlers.CreateCategory).Methods("POST")
	router.HandleFunc("/api/categories/{id}", handlers.GetCategory).Methods("GET")
	router.HandleFunc("/api/categories/{id}", handlers.UpdateCategory).Methods("PUT")
	router.HandleFunc("/api/categories/{id}", handlers.DeleteCategory).Methods("DELETE")

	// Router API cho Reminder
	router.HandleFunc("/api/reminders", handlers.CreateReminder).Methods("POST")
	router.HandleFunc("/api/reminders/{id}", handlers.UpdateReminder).Methods("PUT")
	router.HandleFunc("/api/reminders/{id}", handlers.DeleteReminder).Methods("DELETE")
	router.HandleFunc("/api/tasks/{task_id}/reminders", handlers.GetTaskReminders).Methods("GET")

	// statistics
	router.HandleFunc("/api/users/{user_id}/statistics", handlers.GetUserTaskStatistics).Methods("GET")
	router.HandleFunc("/api/users/{user_id}/tasks-with-reminders", handlers.GetTasksWithReminders).Methods("GET")
	// Khởi động server
	
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("Không tìm thấy PORT, dùng mặc định: 8080")
		port = "8080"
	}

	fmt.Printf("✅ Server đang chạy tại: http://0.0.0.0:%s\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, router); err != nil {
		fmt.Printf("Không thể khởi động server: %v\n", err)
	}
}
