package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"nofx/logger"
	"nofx/store"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// TaskManager manages scheduled tasks
type TaskManager struct {
	store         *store.Store
	traderManager *TraderManager
	cron          *cron.Cron
	entryMap      map[string]cron.EntryID
	mu            sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewTaskManager creates a new task manager
func NewTaskManager(s *store.Store, tm *TraderManager) *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskManager{
		store:         s,
		traderManager: tm,
		// Support both standard (5 fields) and seconds-extended (6 fields) cron expressions
		// Also support descriptors like @daily, @hourly
		cron:          cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		))),
		entryMap:      make(map[string]cron.EntryID),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the task manager
func (m *TaskManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Info("🚀 Starting Task Manager...")

	// Load tasks from database
	if err := m.LoadTasks(); err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	m.cron.Start()
	logger.Info("✅ Task Manager started")
	return nil
}

// Stop stops the task manager
func (m *TaskManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Info("⏹ Stopping Task Manager...")
	m.cron.Stop()
	m.cancel()
}

// LoadTasks loads all tasks from database and schedules them
func (m *TaskManager) LoadTasks() error {
	tasks, err := m.store.Task().List()
	if err != nil {
		return err
	}

	logger.Infof("📋 Loading %d scheduled tasks...", len(tasks))

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		if err := m.scheduleTask(&task); err != nil {
			logger.Errorf("❌ Failed to schedule task %s (%s): %v", task.Name, task.ID, err)
		}
	}

	return nil
}

// AddTask adds a new task
func (m *TaskManager) AddTask(task *store.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Save to DB
	if err := m.store.Task().Create(task); err != nil {
		return err
	}

	// Schedule if enabled
	if task.Enabled {
		return m.scheduleTask(task)
	}
	return nil
}

// UpdateTask updates an existing task
func (m *TaskManager) UpdateTask(task *store.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update in DB
	if err := m.store.Task().Update(task); err != nil {
		return err
	}

	// Remove existing schedule
	if entryID, exists := m.entryMap[task.ID]; exists {
		m.cron.Remove(entryID)
		delete(m.entryMap, task.ID)
	}

	// Schedule if enabled
	if task.Enabled {
		if err := m.scheduleTask(task); err != nil {
			logger.Errorf("Failed to reschedule task %s: %v", task.Name, err)
			return err
		}
	}
	return nil
}

// DeleteTask deletes a task
func (m *TaskManager) DeleteTask(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Delete from DB
	if err := m.store.Task().Delete(id); err != nil {
		return err
	}

	// Remove from schedule
	if entryID, exists := m.entryMap[id]; exists {
		m.cron.Remove(entryID)
		delete(m.entryMap, id)
	}

	return nil
}

// RunTask manually runs a task immediately
func (m *TaskManager) RunTask(id string) error {
	task, err := m.store.Task().Get(id)
	if err != nil {
		return err
	}

	go m.executeTask(task)
	return nil
}

// scheduleTask internal method to schedule a task
func (m *TaskManager) scheduleTask(task *store.Task) error {
	// Parse cron expression
	// If cron expression is empty, skip
	if task.CronExpression == "" {
		return fmt.Errorf("cron expression is empty")
	}

	// Wrap execution logic
	job := cron.FuncJob(func() {
		m.executeTask(task)
	})

	entryID, err := m.cron.AddJob(task.CronExpression, job)
	if err != nil {
		return fmt.Errorf("invalid cron expression '%s': %w", task.CronExpression, err)
	}

	m.entryMap[task.ID] = entryID
	logger.Infof("📅 Scheduled task: %s (%s) -> %s", task.Name, task.Type, task.CronExpression)
	return nil
}

// executeTask executes the task logic
func (m *TaskManager) executeTask(task *store.Task) {
	startTime := time.Now().UnixMilli()
	logger.Infof("▶️ Executing task: %s (%s)", task.Name, task.ID)

	var err error

	switch task.Type {
	case "report":
		err = m.executeReportTask(task)
	case "sync":
		// err = m.executeSyncTask(task) // Not implemented yet, handled by AutoTrader internally for now
		logger.Warnf("Task type 'sync' not fully implemented yet")
	case "custom":
		err = m.executeCustomTask(task)
	default:
		logger.Warnf("Unknown task type: %s", task.Type)
		err = fmt.Errorf("unknown task type")
	}

	// Update task status
	endTime := time.Now().UnixMilli()
	nextTime := int64(0)

	// Calculate next run time
	if entryID, exists := m.entryMap[task.ID]; exists {
		entry := m.cron.Entry(entryID)
		nextTime = entry.Next.UnixMilli()
	}

	if updateErr := m.store.Task().UpdateLastRunTime(task.ID, startTime, nextTime); updateErr != nil {
		logger.Errorf("Failed to update task execution time: %v", updateErr)
	}

	if err != nil {
		logger.Errorf("❌ Task %s failed: %v", task.Name, err)
	} else {
		logger.Infof("✅ Task %s completed in %d ms", task.Name, endTime-startTime)
	}
}

// executeReportTask executes a report task
func (m *TaskManager) executeReportTask(task *store.Task) error {
	if task.TraderID == "" {
		return fmt.Errorf("trader_id is required for report task")
	}

	// Get Trader
	at, err := m.traderManager.GetTrader(task.TraderID)
	if err != nil {
		return fmt.Errorf("trader %s not found: %w", task.TraderID, err)
	}

	// Parse params if needed (e.g., report format)
	var params map[string]interface{}
	if task.Params != "" {
		_ = json.Unmarshal([]byte(task.Params), &params)
	}

	// Execute report
	at.SendHourlyReport() // Currently hardcoded format, can be extended later
	return nil
}

// executeCustomTask executes a custom shell script task
func (m *TaskManager) executeCustomTask(task *store.Task) error {
	if task.Params == "" {
		return fmt.Errorf("script path is required in params for custom task")
	}

	// Params stores the script path directly or as JSON
	scriptPath := task.Params
	
	// Basic check if it looks like a path
	if strings.HasPrefix(scriptPath, "{") {
		// Try to parse as JSON if it looks like JSON
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(task.Params), &params); err == nil {
			if path, ok := params["script_path"].(string); ok {
				scriptPath = path
			}
		}
	}

	logger.Infof("Running custom script: %s", scriptPath)
	
	// Create command
	// Note: We use fields[0] as command and fields[1:] as args to support args in script path
	fields := strings.Fields(scriptPath)
	if len(fields) == 0 {
		return fmt.Errorf("empty script path")
	}
	
	cmdName := fields[0]
	cmdArgs := fields[1:]
	
	cmd := exec.Command(cmdName, cmdArgs...)
	
	// Inherit environment variables from parent process
	cmd.Env = os.Environ()
	
	// If trader is associated, pass trader ID as env var
	if task.TraderID != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TRADER_ID=%s", task.TraderID))
	}

	// Run command
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		logger.Infof("Script output: %s", string(output))
	}
	
	if err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}
