package routers

import (
	"UserManagementVer/controllers"
	"UserManagementVer/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type AuthRouter struct {
	authController *controllers.AuthController
	rdb            *redis.Client
}

func NewAuthRouter(authController *controllers.AuthController, rdb *redis.Client) *AuthRouter {
	return &AuthRouter{authController: authController, rdb: rdb}
}
func (authRouter *AuthRouter) Register(router *gin.RouterGroup) {
	rateLimitMiddleWare := middlewares.NewRateLimitService(authRouter.rdb)
	authRou := router.Group("/auth")
	{
		authRou.POST("/login", rateLimitMiddleWare.NewRateLimiterMiddleware(), authRouter.authController.Login)
		authRou.GET("/sessions", authRouter.authController.ConfirmLogin)
		authRou.POST("/logout", authRouter.authController.Logout)
		authRou.POST("/renew", authRouter.authController.RenewAccessToken)
	}
}
