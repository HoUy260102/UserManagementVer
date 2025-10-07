package routers

import (
	"UserManagementVer/collections"
	"UserManagementVer/configs"
	"UserManagementVer/controllers"
	"UserManagementVer/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRouters(db *mongo.Database, v *gin.RouterGroup) {
	accountCollection := collections.NewAccountCollection(db.Collection("accounts"))
	sessionCollection := collections.NewSessionCollection(db.Collection("sessions"))
	emailService := services.NewEmailService(configs.AppConfig.Email.Host, configs.AppConfig.Email.User, configs.AppConfig.Email.Pass, configs.AppConfig.Email.Port)
	jwtService := services.NewJwtService(configs.AppConfig.Jwt.SecretKey, configs.AppConfig.Jwt.Issuer)
	accountController := controllers.NewAccountController(accountCollection, jwtService)
	authController := controllers.NewAuthController(sessionCollection, accountCollection, emailService, jwtService)
	authRouter := NewAuthRouter(authController)
	accountRouter := NewAccountRouter(accountController)
	accountRouter.RegisterRoutes(v, jwtService)
	authRouter.Register(v)
}
