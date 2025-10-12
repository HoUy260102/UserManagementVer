package app

import (
	"UserManagementVer/collections"
	"UserManagementVer/configs"
	"UserManagementVer/controllers"
	"UserManagementVer/routers"
	"UserManagementVer/services"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthModule struct {
	routers routers.Router
}

func NewAuthModule(db *mongo.Database, rdb *redis.Client) *AuthModule {
	accountCollection := collections.NewAccountCollection(db.Collection("accounts"))
	sessionCollection := collections.NewSessionCollection(db.Collection("sessions"))
	emailService := services.NewEmailService(configs.AppConfig.Email.Host, configs.AppConfig.Email.User, configs.AppConfig.Email.Pass, configs.AppConfig.Email.Port)
	jwtService := services.NewJwtService()
	authController := controllers.NewAuthController(sessionCollection, accountCollection, emailService, jwtService)
	authRouter := routers.NewAuthRouter(authController, rdb)
	return &AuthModule{
		routers: authRouter,
	}
}

func (m *AuthModule) Routers() routers.Router {
	return m.routers
}
