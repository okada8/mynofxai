package api

import (
	"net/http"
	"nofx/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handleListTasks List all tasks
func (s *Server) handleListTasks(c *gin.Context) {
	tasks, err := s.store.Task().List()
	if err != nil {
		SafeInternalError(c, "Failed to list tasks", err)
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// handleCreateTask Create a new task
func (s *Server) handleCreateTask(c *gin.Context) {
	var req struct {
		Name           string `json:"name" binding:"required"`
		Type           string `json:"type" binding:"required"`
		TraderID       string `json:"trader_id"`
		CronExpression string `json:"cron_expression" binding:"required"`
		Enabled        bool   `json:"enabled"`
		Params         string `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters")
		return
	}

	task := &store.Task{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Type:           req.Type,
		TraderID:       req.TraderID,
		CronExpression: req.CronExpression,
		Enabled:        req.Enabled,
		Params:         req.Params,
	}

	if err := s.taskManager.AddTask(task); err != nil {
		SafeInternalError(c, "Failed to create task", err)
		return
	}

	c.JSON(http.StatusCreated, task)
}

// handleUpdateTask Update a task
func (s *Server) handleUpdateTask(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name           string  `json:"name"`
		Type           string  `json:"type"`
		TraderID       *string `json:"trader_id"` // Use pointer to distinguish between missing and empty
		CronExpression string  `json:"cron_expression"`
		Enabled        *bool   `json:"enabled"`
		Params         *string `json:"params"` // Use pointer to distinguish between missing and empty
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request parameters")
		return
	}

	task, err := s.store.Task().Get(id)
	if err != nil {
		SafeNotFound(c, "Task")
		return
	}

	if req.Name != "" {
		task.Name = req.Name
	}
	if req.Type != "" {
		task.Type = req.Type
	}
	
	// Update TraderID if provided (including empty string)
	if req.TraderID != nil {
		task.TraderID = *req.TraderID
	}
	
	if req.CronExpression != "" {
		task.CronExpression = req.CronExpression
	}
	if req.Enabled != nil {
		task.Enabled = *req.Enabled
	}
	
	// Update Params if provided (including empty string)
	if req.Params != nil {
		task.Params = *req.Params
	}

	if err := s.taskManager.UpdateTask(task); err != nil {
		SafeInternalError(c, "Failed to update task", err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// handleDeleteTask Delete a task
func (s *Server) handleDeleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := s.taskManager.DeleteTask(id); err != nil {
		SafeInternalError(c, "Failed to delete task", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

// handleRunTask Run a task manually
func (s *Server) handleRunTask(c *gin.Context) {
	id := c.Param("id")
	if err := s.taskManager.RunTask(id); err != nil {
		SafeInternalError(c, "Failed to run task", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task triggered"})
}
