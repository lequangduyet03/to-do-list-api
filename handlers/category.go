package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"backend/database"
	"backend/models"

	"github.com/gorilla/mux"
)

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&category); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	defer r.Body.Close()

	// Kiểm tra trùng lặp danh mục
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE category_name = ? AND user_id = ?)", category.CategoryName, category.UserID).Scan(&exists)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi kiểm tra danh mục")
		return
	}
	if exists {
		RespondWithError(w, http.StatusConflict, "Danh mục đã tồn tại")
		return
	}

	query := "INSERT INTO categories (category_name, color, user_id, description) VALUES (?, ?, ?, ?)"
	result, err := database.DB.Exec(query, category.CategoryName, category.Color, category.UserID, category.Description)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi thêm danh mục")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy ID danh mục")
		return
	}

	category.ID = int(id)
	RespondWithJSON(w, http.StatusCreated, category)
}

func GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	var category models.Category
	query := "SELECT category_id, category_name, color, user_id, description FROM categories WHERE category_id = ?"
	err = database.DB.QueryRow(query, id).Scan(&category.ID, &category.CategoryName, &category.Color, &category.UserID, &category.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			RespondWithError(w, http.StatusNotFound, "Không tìm thấy danh mục")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy thông tin danh mục"+err.Error())
		return
	}

	RespondWithJSON(w, http.StatusOK, category)
}

func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	var category models.Category
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&category); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	defer r.Body.Close()

	query := "UPDATE categories SET category_name = ?, color = ?, description = ? WHERE category_id = ?"
	result, err := database.DB.Exec(query, category.CategoryName, category.Color, category.Description, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi cập nhật danh mục"+err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		RespondWithError(w, http.StatusNotFound, "Không tìm thấy danh mục để cập nhật")
		return
	}

	category.ID = id
	RespondWithJSON(w, http.StatusOK, category)
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	query := "DELETE FROM categories WHERE category_id = ?"
	result, err := database.DB.Exec(query, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi xóa danh mục")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		RespondWithError(w, http.StatusNotFound, "Không tìm thấy danh mục để xóa")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Đã xóa danh mục thành công"})
}
