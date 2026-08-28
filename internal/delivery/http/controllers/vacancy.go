package controllers

import (
	"net/http"
	"strconv"

	"github.com/company/hrbot/internal/domain/vacancy"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VacancyController struct {
	repo vacancy.Repository
}

func NewVacancyController(repo vacancy.Repository) *VacancyController {
	return &VacancyController{repo: repo}
}

func (vc *VacancyController) GetAll(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if err != nil || limit < 1 {
		limit = 100
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	vacancies, err := vc.repo.GetAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vacancies"})
		return
	}
	c.JSON(http.StatusOK, vacancies)
}

func (vc *VacancyController) Create(c *gin.Context) {
	var v vacancy.Vacancy
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := vc.repo.Create(c.Request.Context(), &v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vacancy"})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (vc *VacancyController) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var v vacancy.Vacancy
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v.ID = id

	if err := vc.repo.Update(c.Request.Context(), &v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vacancy"})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (vc *VacancyController) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := vc.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vacancy"})
		return
	}
	c.Status(http.StatusNoContent)
}
