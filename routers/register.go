package routers

import (
	"UserManagementVer/collections"
	"UserManagementVer/configs"
	"UserManagementVer/controllers"
	"UserManagementVer/middlewares"
	"UserManagementVer/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRouters(db *mongo.Database, v *gin.RouterGroup, rdb *redis.Client) {
	accountCollection := collections.NewAccountCollection(db.Collection("accounts"))
	sessionCollection := collections.NewSessionCollection(db.Collection("sessions"))
	emailService := services.NewEmailService(configs.AppConfig.Email.Host, configs.AppConfig.Email.User, configs.AppConfig.Email.Pass, configs.AppConfig.Email.Port)
	jwtService := services.NewJwtService()
	accountController := controllers.NewAccountController(accountCollection, jwtService)
	authController := controllers.NewAuthController(sessionCollection, accountCollection, emailService, jwtService)
	rateLimit := middlewares.NewRateLimitService(rdb)
	authRouter := NewAuthRouter(authController)
	accountRouter := NewAccountRouter(accountController)
	accountRouter.RegisterRoutes(v, jwtService)
	authRouter.Register(v, rateLimit)
}
