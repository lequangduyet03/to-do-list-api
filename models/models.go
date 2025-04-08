package models

import "time"

type User struct {
    ID        int        `json:"user_id"`
    Username  string     `json:"username"`
    Email     string     `json:"email"`
    Password  string     `json:"password"` // Sửa tag để ánh xạ từ JSON
    FullName  string     `json:"full_name"`
    CreatedAt time.Time  `json:"created_at"`
    LastLogin *time.Time `json:"last_login"`
}
type Task struct {
	ID          int       `json:"task_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	CategoryID  int       `json:"category_id"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Category struct {
	ID           int    `json:"category_id"`
	CategoryName string `json:"category_name"`
	Color        string `json:"color"`
	UserID       int    `json:"user_id"`
	Description  string `json:"description"` // Thêm trường mới
}

type Reminder struct {
	ID           int       `json:"id"`           // Giữ là "id" để đồng bộ với Flutter
	TaskID       int       `json:"task_id"`
	UserID       int       `json:"user_id"`
	ReminderTime time.Time `json:"reminder_time"`
	IsSent       bool      `json:"is_sent"`      // Đổi thành bool để đồng bộ với Flutter
}


type TaskStatistics struct {
	ID             int       `json:"stat_id"`
	UserID         int       `json:"user_id"`
	Date           time.Time `json:"date"`
	CompletedTasks int       `json:"completed_tasks"`
	PendingTasks   int       `json:"pending_tasks"`
	OverdueTasks   int       `json:"overdue_tasks"`
}

type UserTaskStatistics struct {
    UserID           int                     `json:"user_id"`
    TotalTasks       int                     `json:"total_tasks"`
    CompletedTasks   int                     `json:"completed_tasks"`
    InProgressTasks  int                     `json:"in_progress_tasks"`
    PendingTasks     int                     `json:"pending_tasks"`
    OverdueTasks     int                     `json:"overdue_tasks"`
    TasksByMonth     map[string]int          `json:"tasks_by_month"`
    CompletedByMonth map[string]int          `json:"completed_by_month"`
}

