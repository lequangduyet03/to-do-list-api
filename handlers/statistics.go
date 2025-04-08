package handlers

import (
    "database/sql"
    "net/http"
    "strconv"

    "backend/database"
    "backend/models"

    "github.com/gorilla/mux"
)

func GetUserTaskStatistics(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    userIDStr, ok := vars["user_id"]
    if !ok {
        RespondWithError(w, http.StatusBadRequest, "Thiếu ID người dùng")
        return
    }

    userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        RespondWithError(w, http.StatusBadRequest, "ID người dùng không hợp lệ")
        return
    }

    // Khởi tạo struct thống kê
    stats := models.UserTaskStatistics{
        UserID:           userID,
        TasksByMonth:     make(map[string]int),
        CompletedByMonth: make(map[string]int),
    }

    // Truy vấn tổng quan thống kê
    queryOverview := `
        SELECT 
            COUNT(*) as total_tasks,
            SUM(CASE WHEN status = 'Completed' THEN 1 ELSE 0 END) as completed_tasks,
            SUM(CASE WHEN status = 'In Progress' THEN 1 ELSE 0 END) as in_progress_tasks,
            SUM(CASE WHEN status = 'Pending' THEN 1 ELSE 0 END) as pending_tasks,
            SUM(CASE WHEN deadline < NOW() AND status != 'Completed' THEN 1 ELSE 0 END) as overdue_tasks
        FROM tasks
        WHERE user_id = ?`
    err = database.DB.QueryRow(queryOverview, userID).Scan(
        &stats.TotalTasks,
        &stats.CompletedTasks,
        &stats.InProgressTasks,
        &stats.PendingTasks,
        &stats.OverdueTasks,
    )
    if err != nil {
        if err != sql.ErrNoRows {
            RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy thống kê tổng quan: "+err.Error())
            return
        }
        // Nếu không có dữ liệu, giữ nguyên giá trị mặc định (0)
    }

    // Truy vấn thống kê theo tháng
    queryByMonth := `
        SELECT 
            DATE_FORMAT(created_at, '%b') as month,
            COUNT(*) as total_tasks,
            SUM(CASE WHEN status = 'Completed' THEN 1 ELSE 0 END) as completed_tasks
        FROM tasks
        WHERE user_id = ?
        GROUP BY DATE_FORMAT(created_at, '%b')`
    rows, err := database.DB.Query(queryByMonth, userID)
    if err != nil {
        RespondWithError(w, http.StatusInternalServerError, "Lỗi khi lấy thống kê theo tháng: "+err.Error())
        return
    }
    defer rows.Close()

    for rows.Next() {
        var month string
        var totalTasks, completedTasks int
        if err := rows.Scan(&month, &totalTasks, &completedTasks); err != nil {
            RespondWithError(w, http.StatusInternalServerError, "Lỗi khi đọc dữ liệu theo tháng: "+err.Error())
            return
        }
        stats.TasksByMonth[month] = totalTasks
        stats.CompletedByMonth[month] = completedTasks
    }

    if err = rows.Err(); err != nil {
        RespondWithError(w, http.StatusInternalServerError, "Lỗi khi truy vấn dữ liệu theo tháng: "+err.Error())
        return
    }

    // Trả về JSON
    RespondWithJSON(w, http.StatusOK, stats)
}