package routers

import (
	"UserManagementVer/controllers"
	"UserManagementVer/middlewares"

	"github.com/gin-gonic/gin"
)

type AuthRouter struct {
	authController *controllers.AuthController
}

func NewAuthRouter(authController *controllers.AuthController) *AuthRouter {
	return &AuthRouter{authController: authController}
}

func (authRouter *AuthRouter) Register(router *gin.RouterGroup, rateLimit *middlewares.RateLimit) {
	authRou := router.Group("/auth")
	{
		authRou.POST("/login", rateLimit.NewRateLimiterMiddleware(), authRouter.authController.Login)
		authRou.GET("/sessions", authRouter.authController.ConfirmLogin)
	}
}
