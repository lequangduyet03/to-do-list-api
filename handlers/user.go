package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"backend/database"
	"backend/models"

	"github.com/go-sql-driver/mysql"

	"github.com/gorilla/mux"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
    var user models.User
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&user); err != nil {
        log.Printf("Lỗi giải mã JSON: %v", err)
        RespondWithError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
        return
    }
    defer r.Body.Close()

    log.Printf("Dữ liệu nhận được: %+v", user)

    if user.Username == "" || user.Email == "" || user.Password == "" || user.FullName == "" {
        log.Printf("Trường rỗng: Username=%s, Email=%s, Password=%s, FullName=%s",
            user.Username, user.Email, user.Password, user.FullName)
        RespondWithError(w, http.StatusBadRequest, "Tất cả các trường là bắt buộc")
        return
    }

    query := `INSERT INTO users (username, email, password, full_name, created_at) 
              VALUES (?, ?, ?, ?, ?)`
    result, err := database.DB.Exec(query, user.Username, user.Email, user.Password, user.FullName, time.Now())
    if err != nil {
        if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
            log.Printf("Trùng lặp username hoặc email: %v", err)
            RespondWithError(w, http.StatusConflict, "Username hoặc email đã tồn tại")
        } else {
            log.Printf("Lỗi SQL: %v", err)
            RespondWithError(w, http.StatusInternalServerError, "Lỗi khi thêm người dùng: "+err.Error())
        }
        return
    }

    id, err := result.LastInsertId()
    if err != nil {
        log.Printf("Lỗi lấy ID: %v", err)
        RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy ID người dùng")
        return
    }

    user.ID = int(id)
    w.Header().Set("Content-Type", "application/json")
    RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
        "user": user,
    })
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		RespondWithError(w, http.StatusBadRequest, "Thiếu ID người dùng")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	var user models.User
	var lastLogin sql.NullTime
	query := "SELECT user_id, username, email, full_name, created_at, last_login FROM users WHERE user_id = ?"
	err = database.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.CreatedAt,
		&lastLogin,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Không tìm thấy người dùng với ID: %d", id))
			return
		}
		log.Println("Lỗi truy vấn SQL:", err)
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy thông tin người dùng")
		return
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	w.Header().Set("Content-Type", "application/json")
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Lấy thông tin người dùng thành công",
		"user":    user,
	})
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	var user models.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	defer r.Body.Close()

	query := "UPDATE users SET username = ?, email = ?, full_name = ? WHERE user_id = ?"
	_, err = database.DB.Exec(query, user.Username, user.Email, user.FullName, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi cập nhật người dùng: "+err.Error())
		return
	}

	user.ID = id
	w.Header().Set("Content-Type", "application/json")
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Cập nhật người dùng thành công",
		"user":    user,
	})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	query := "DELETE FROM users WHERE user_id = ?"
	_, err = database.DB.Exec(query, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi xóa người dùng")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Đã xóa người dùng thành công"})
}

func GetUserTasks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID người dùng không hợp lệ")
		return
	}

	// Lấy danh sách công việc của người dùng
	query := `SELECT task_id, title, description, deadline, priority, status, 
              category_id, user_id, created_at, updated_at 
              FROM tasks WHERE user_id = ?`
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy danh sách công việc")
		return
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
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
			RespondWithError(w, http.StatusInternalServerError, "Lỗi khi đọc dữ liệu công việc")
			return
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi truy vấn danh sách công việc")
		return
	}

	RespondWithJSON(w, http.StatusOK, tasks)
}

func GetUserCategories(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID người dùng không hợp lệ")
		return
	}

	query := "SELECT category_id, category_name, color, user_id FROM categories WHERE user_id = ?"
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy danh sách danh mục")
		return
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.CategoryName, &category.Color, &category.UserID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Lỗi khi đọc dữ liệu danh mục")
			return
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi truy vấn danh sách danh mục")
		return
	}

	RespondWithJSON(w, http.StatusOK, categories)
}

// Cấu trúc dữ liệu đăng nhập
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login xử lý việc đăng nhập người dùng
func Login(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&loginReq); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Dữ liệu đăng nhập không hợp lệ")
		return
	}
	defer r.Body.Close()

	// Truy vấn thông tin người dùng từ database
	var user models.User
	var storedPassword string

	query := "SELECT user_id, username, email, password, full_name, created_at FROM users WHERE username = ?"
	err := database.DB.QueryRow(query, loginReq.Username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&storedPassword,
		&user.FullName,
		&user.CreatedAt,
	)

	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Tên đăng nhập hoặc mật khẩu không đúng")
		return
	}

	// Kiểm tra mật khẩu (so sánh trực tiếp)
	if storedPassword != loginReq.Password {
		RespondWithError(w, http.StatusUnauthorized, "Tên đăng nhập hoặc mật khẩu không đúng")
		return
	}

	// Cập nhật thời gian đăng nhập cuối cùng
	now := time.Now()
	database.DB.Exec("UPDATE users SET last_login = ? WHERE user_id = ?", now, user.ID)

	// Xóa mật khẩu trước khi trả về
	user.Password = ""

	// Trả về thông tin người dùng
	w.Header().Set("Content-Type", "application/json")
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Đăng nhập thành công",
		"user":    user,
	})
}
