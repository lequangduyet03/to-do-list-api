package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"backend/database"
	"backend/models"

	"github.com/gorilla/mux"
)

func CreateReminder(w http.ResponseWriter, r *http.Request) {
	var reminder models.Reminder
	if err := json.NewDecoder(r.Body).Decode(&reminder); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Dữ liệu không đúng định dạng JSON: "+err.Error())
		return
	}
	defer r.Body.Close()

	if reminder.TaskID == 0 || reminder.UserID == 0 || reminder.ReminderTime.IsZero() {
		RespondWithError(w, http.StatusBadRequest, "Thiếu thông tin cần thiết")
		return
	}

	// Kiểm tra xem có nhắc nhở nào cho cùng task_id và user_id trong khoảng thời gian gần reminder_time không
	var exists int
	// Khoảng thời gian kiểm tra: 1 phút (60 giây) trước và sau reminder_time
	lowerBound := reminder.ReminderTime.Add(-1 * time.Minute)
	upperBound := reminder.ReminderTime.Add(1 * time.Minute)
	queryCheck := `
		SELECT COUNT(*) 
		FROM reminders 
		WHERE task_id = ? 
		AND user_id = ? 
		AND reminder_time BETWEEN ? AND ?`
	err := database.DB.QueryRow(queryCheck, reminder.TaskID, reminder.UserID, lowerBound, upperBound).Scan(&exists)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi kiểm tra nhắc nhở trùng lặp: "+err.Error())
		return
	}
	if exists > 0 {
		RespondWithError(w, http.StatusConflict, "Đã có nhắc nhở cho công việc này trong khoảng thời gian gần thời điểm bạn chọn")
		return
	}

	// Nếu không có nhắc nhở nào trong khoảng thời gian, tiến hành tạo mới
	query := "INSERT INTO reminders (task_id, user_id, reminder_time, is_sent, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())"
	result, err := database.DB.Exec(query, reminder.TaskID, reminder.UserID, reminder.ReminderTime, reminder.IsSent)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi tạo nhắc nhở: "+err.Error())
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy ID nhắc nhở")
		return
	}

	reminder.ID = int(id)
	RespondWithJSON(w, http.StatusCreated, reminder)
}

func GetTaskReminders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["task_id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID công việc không hợp lệ")
		return
	}

	query := "SELECT reminder_id, task_id, user_id, reminder_time, is_sent FROM reminders WHERE task_id = ?"
	rows, err := database.DB.Query(query, taskID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy danh sách nhắc nhở")
		return
	}
	defer rows.Close()

	reminders := []models.Reminder{}
	for rows.Next() {
		var reminder models.Reminder
		if err := rows.Scan(&reminder.ID, &reminder.TaskID, &reminder.UserID, &reminder.ReminderTime, &reminder.IsSent); err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Lỗi khi đọc dữ liệu nhắc nhở")
			return
		}
		reminders = append(reminders, reminder)
	}

	if err = rows.Err(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi truy vấn danh sách nhắc nhở")
		return
	}

	RespondWithJSON(w, http.StatusOK, reminders)
}

func UpdateReminder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	var reminder models.Reminder
	if err := json.NewDecoder(r.Body).Decode(&reminder); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Dữ liệu không đúng định dạng JSON")
		return
	}
	defer r.Body.Close()

	query := "UPDATE reminders SET reminder_time = ?, is_sent = ?, user_id = ?, updated_at = NOW() WHERE reminder_id = ?"
	result, err := database.DB.Exec(query, reminder.ReminderTime, reminder.IsSent, reminder.UserID, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi cập nhật nhắc nhở: "+err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		RespondWithError(w, http.StatusNotFound, "Không tìm thấy nhắc nhở để cập nhật")
		return
	}

	reminder.ID = id
	RespondWithJSON(w, http.StatusOK, reminder)
}

func DeleteReminder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID không hợp lệ")
		return
	}

	query := "DELETE FROM reminders WHERE reminder_id = ?"
	result, err := database.DB.Exec(query, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi xóa nhắc nhở")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		RespondWithError(w, http.StatusNotFound, "Không tìm thấy nhắc nhở để xóa")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Đã xóa nhắc nhở thành công"})
}

// GetTasksWithReminders trả về danh sách các task có reminder của một user
func GetTasksWithReminders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "ID người dùng không hợp lệ")
		return
	}

	// Truy vấn SQL để lấy các task có reminder
	query := `
		SELECT DISTINCT t.task_id, t.user_id, t.title, t.description, t.priority, t.status, t.deadline, t.created_at, t.updated_at
		FROM tasks t
		INNER JOIN reminders r ON t.task_id = r.task_id
		WHERE t.user_id = ?`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy danh sách công việc: "+err.Error())
		return
	}
	defer rows.Close()

	// Khai báo slice để lưu danh sách task
	tasks := []models.Task{}

	// Duyệt qua các dòng kết quả
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Priority,
			&task.Status,
			&task.Deadline,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Lỗi khi đọc dữ liệu công việc: "+err.Error())
			return
		}
		tasks = append(tasks, task)
	}

	// Kiểm tra lỗi sau khi duyệt rows
	if err = rows.Err(); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Lỗi khi truy vấn danh sách công việc: "+err.Error())
		return
	}

	// Trả về danh sách task dưới dạng JSON
	RespondWithJSON(w, http.StatusOK, tasks)
}
