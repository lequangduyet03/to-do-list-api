package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"backend/database"
	"backend/models"

	"database/sql"

	"github.com/gorilla/mux"
)

func CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&task); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	defer r.Body.Close()

	// Kiểm tra deadline, nếu không có thì mặc định là now
	if task.Deadline.IsZero() {
		task.Deadline = time.Now()
	}

	// Kiểm tra dữ liệu đầu vào
	if task.Title == "" || task.Description == "" || task.Deadline.IsZero() {
		RespondWithError(w, http.StatusBadRequest, "Thiếu thông tin cần thiết")
		return
	}

	now := time.Now()
	query := `INSERT INTO tasks 
	          (title, description, deadline, priority, status, category_id, user_id, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := database.DB.Exec(
		query,
		task.Title,
		task.Description,
		task.Deadline,
		task.Priority,
		task.Status,
		task.CategoryID,
		task.UserID,
		now,
		now,
	)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi thêm công việc: "+err.Error())
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy ID công việc")
		return
	}

	task.ID = int(id)
	task.CreatedAt = now
	task.UpdatedAt = now
	RespondWithJSON(w, http.StatusCreated, task)
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Thiếu ID công việc")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	var task models.Task
	query := `SELECT task_id, title, description, deadline, priority, status, 
              category_id, user_id, created_at, updated_at 
              FROM tasks WHERE task_id = ?`
	err = database.DB.QueryRow(query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Deadline,
		&task.Priority,
		&task.Status,
		&task.CategoryID,
		&task.UserID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Không tìm thấy công việc với ID: %d", id))
			return
		}
		log.Println("Lỗi truy vấn SQL:", err)
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy thông tin công việc")
		return
	}

	RespondWithJSON(w, http.StatusOK, task)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var task models.Task
	body, err := io.ReadAll(r.Body) // Error: undefined: io
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Unable to read request body")
		return
	}
	defer r.Body.Close()

	fmt.Printf("Request Body: %s\n", string(body))

	if err := json.Unmarshal(body, &task); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid data: "+err.Error())
		return
	}

	now := time.Now()
	query := `UPDATE tasks 
              SET title = ?, description = ?, deadline = ?, priority = ?, 
              status = ?, category_id = ?, updated_at = ? 
              WHERE task_id = ?`
	result, err := database.DB.Exec(
		query,
		task.Title,
		task.Description,
		task.Deadline,
		task.Priority,
		task.Status,
		task.CategoryID,
		now,
		id,
	)
	if err != nil {
		fmt.Printf("Database error: %v\n", err)
		RespondWithError(w, http.StatusInternalServerError, "Error updating task: "+err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		RespondWithError(w, http.StatusNotFound, "Task not found for updating")
		return
	}

	task.ID = id
	task.UpdatedAt = now
	RespondWithJSON(w, http.StatusOK, task)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	query := "DELETE FROM tasks WHERE task_id = ?"
	result, err := database.DB.Exec(query, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi xóa công việc")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		RespondWithError(w, http.StatusNotFound, "Không tìm thấy công việc để xóa")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Đã xóa công việc thành công"})
}
