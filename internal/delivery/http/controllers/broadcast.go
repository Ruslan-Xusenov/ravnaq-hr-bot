package controllers

import (
	"net/http"

	"github.com/company/hrbot/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type BroadcastController struct {
	client *asynq.Client
}

func NewBroadcastController(client *asynq.Client) *BroadcastController {
	return &BroadcastController{client: client}
}

type broadcastRequest struct {
	Text     string `json:"text" binding:"required"`
	LangCode string `json:"lang_code"`
}

func (bc *BroadcastController) SendBroadcast(c *gin.Context) {
	var req broadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := worker.NewBroadcastMessageTask(req.Text, req.LangCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	info, err := bc.client.Enqueue(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue task"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Broadcast task enqueued successfully",
		"task_id": info.ID,
		"queue":   info.Queue,
	})
}
