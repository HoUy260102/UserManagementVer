package app

import (
	"UserManagementVer/collections"
	"UserManagementVer/controllers"
	"UserManagementVer/routers"
	"UserManagementVer/services"

	"go.mongodb.org/mongo-driver/mongo"
)

type AccountModule struct {
	routers routers.Router
}

func NewAccountModule(db *mongo.Database) *AccountModule {
	accountCollection := collections.NewAccountCollection(db.Collection("accounts"))
	jwtService := services.NewJwtService()
	accountController := controllers.NewAccountController(accountCollection, jwtService)
	accountRouter := routers.NewAccountRouter(accountController)
	return &AccountModule{
		routers: accountRouter,
	}
}

func (m *AccountModule) Routers() routers.Router {
	return m.routers
}
