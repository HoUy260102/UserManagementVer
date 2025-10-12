package app

import (
	"UserManagementVer/configs"
	"UserManagementVer/routers"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type Module interface {
	Routers() routers.Router
}

type Application struct {
	config *configs.Config
	router *gin.Engine
}

func NewApplication(cfg *configs.Config, db *mongo.Database, rdb *redis.Client) *Application {
	r := gin.Default()
	modules := []Module{
		NewAccountModule(db), NewAuthModule(db, rdb),
	}
	api := r.Group("/api/v1")
	routers.RegisterRouters(api, getModuleRouters(modules)...)
	return &Application{
		config: cfg,
		router: r,
	}
}

func (app *Application) Run() error {
	return app.router.Run(fmt.Sprintf(":%d", app.config.Server.Port))
}

func getModuleRouters(modules []Module) []routers.Router {
	routers := []routers.Router{}
	for _, m := range modules {
		routers = append(routers, m.Routers())
	}
	return routers
}
