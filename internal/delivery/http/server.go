package http

import (
	"github.com/company/hrbot/internal/delivery/http/controllers"
	"github.com/company/hrbot/internal/delivery/http/middleware"
	"github.com/company/hrbot/internal/domain/admin"
	"github.com/company/hrbot/internal/domain/application"
	"github.com/company/hrbot/internal/domain/vacancy"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

func SetupRouter(
	jwtSecret string,
	asynqClient *asynq.Client,
	adminRepo admin.Repository,
	vacancyRepo vacancy.Repository,
	appRepo application.Repository,
) *gin.Engine {
	r := gin.Default()

	authCtrl := controllers.NewAuthController(adminRepo, jwtSecret)
	vacancyCtrl := controllers.NewVacancyController(vacancyRepo)
	appCtrl := controllers.NewApplicationController(appRepo)
	broadcastCtrl := controllers.NewBroadcastController(asynqClient)

	v1 := r.Group("/api/v1")
	{
		// Public (or semi-public) routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authCtrl.Login)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth(jwtSecret))
		{
			vacancies := protected.Group("/vacancies")
			{
				vacancies.GET("", vacancyCtrl.GetAll)
				vacancies.POST("", vacancyCtrl.Create)
				vacancies.PUT("/:id", vacancyCtrl.Update)
				vacancies.DELETE("/:id", vacancyCtrl.Delete)
			}

			applications := protected.Group("/applications")
			{
				applications.GET("", appCtrl.GetAll)
				applications.PATCH("/:id/status", appCtrl.UpdateStatus)
			}

			broadcast := protected.Group("/broadcast")
			{
				broadcast.POST("", broadcastCtrl.SendBroadcast)
			}
		}
	}

	return r
}
