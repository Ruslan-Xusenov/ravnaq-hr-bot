package controllers

import (
	"net/http"
	"strconv"

	"github.com/company/hrbot/internal/domain/application"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ApplicationController struct {
	repo application.Repository
}

func NewApplicationController(repo application.Repository) *ApplicationController {
	return &ApplicationController{repo: repo}
}

func (ac *ApplicationController) GetAll(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if err != nil || limit < 1 {
		limit = 100
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	apps, err := ac.repo.GetAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}
	c.JSON(http.StatusOK, apps)
}

type updateStatusReq struct {
	Status string `json:"status" binding:"required"`
}

func (ac *ApplicationController) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req updateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ac.repo.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
