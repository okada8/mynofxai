package store

import (
	"time"

	"gorm.io/gorm"
)

// Task scheduled task model
type Task struct {
	ID             string `gorm:"primaryKey" json:"id"`
	Type           string `gorm:"index" json:"type"`      // Task type: "report", "sync", "custom"
	Name           string `json:"name"`                   // Task name
	Description    string `json:"description"`            // Task description
	TraderID       string `gorm:"index" json:"trader_id"` // Associated trader ID (optional)
	CronExpression string `json:"cron_expression"`        // Cron expression (e.g., "@hourly", "0 30 * * * *")
	Enabled        bool   `gorm:"default:true" json:"enabled"`
	Params         string `json:"params"` // JSON parameters
	LastRunTime    int64  `json:"last_run_time"`
	NextRunTime    int64  `json:"next_run_time"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

// TaskStore task storage
type TaskStore struct {
	db *gorm.DB
}

// NewTaskStore creates new task store
func NewTaskStore(db *gorm.DB) *TaskStore {
	return &TaskStore{db: db}
}

// InitTables initializes task tables
func (s *TaskStore) InitTables() error {
	return s.db.AutoMigrate(&Task{})
}

// Create creates a new task
func (s *TaskStore) Create(task *Task) error {
	now := time.Now().UnixMilli()
	task.CreatedAt = now
	task.UpdatedAt = now
	return s.db.Create(task).Error
}

// Update updates a task
func (s *TaskStore) Update(task *Task) error {
	task.UpdatedAt = time.Now().UnixMilli()
	// Use Save to update all fields including zero values
	return s.db.Save(task).Error
}

// Delete deletes a task
func (s *TaskStore) Delete(id string) error {
	return s.db.Delete(&Task{}, "id = ?", id).Error
}

// Get gets a task by ID
func (s *TaskStore) Get(id string) (*Task, error) {
	var task Task
	err := s.db.First(&task, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// List gets all tasks
func (s *TaskStore) List() ([]Task, error) {
	var tasks []Task
	err := s.db.Find(&tasks).Error
	return tasks, err
}

// ListByTrader gets tasks by trader ID
func (s *TaskStore) ListByTrader(traderID string) ([]Task, error) {
	var tasks []Task
	err := s.db.Where("trader_id = ?", traderID).Find(&tasks).Error
	return tasks, err
}

// UpdateLastRunTime updates last run time
func (s *TaskStore) UpdateLastRunTime(id string, lastRunTime int64, nextRunTime int64) error {
	return s.db.Model(&Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_run_time": lastRunTime,
		"next_run_time": nextRunTime,
		"updated_at":    time.Now().UnixMilli(),
	}).Error
}
