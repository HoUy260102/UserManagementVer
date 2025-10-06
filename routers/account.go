package routers

import (
	"UserManagementVer/controllers"
	"UserManagementVer/middlewares"
	"UserManagementVer/services"

	"github.com/gin-gonic/gin"
)

type AccountRouter struct {
	accountController *controllers.AccountController
}

func NewAccountRouter(accountController *controllers.AccountController) *AccountRouter {
	return &AccountRouter{accountController: accountController}
}

func (accountRouter *AccountRouter) RegisterRoutes(router *gin.RouterGroup, jwtService *services.JwtService) {
	accountRou := router.Group("/accounts")
	{
		accountRou.GET("/:id/detail", middlewares.AuthorizeJWT(jwtService), accountRouter.accountController.FindAccountById)
		accountRou.POST("/add", middlewares.AuthorizeJWT(jwtService), accountRouter.accountController.CreateAccount)
		accountRou.PATCH("/:id", middlewares.AuthorizeJWT(jwtService), accountRouter.accountController.UpdateAccount)
		accountRou.PATCH("/:id/restore", middlewares.AuthorizeJWT(jwtService), accountRouter.accountController.RestoreAccount)
		accountRou.PATCH("/:id/soft-delete", middlewares.AuthorizeJWT(jwtService), accountRouter.accountController.SoftDelete)
		accountRou.GET("/update", middlewares.AuthorizeJWT(jwtService), accountRouter.accountController.SearchAccount)
		accountRou.POST("/:id/update-avatar", middlewares.AuthorizeJWT(jwtService), accountRouter.accountController.UploadImage)
		accountRou.GET("/:id/avatar", middlewares.AuthorizeJWT(jwtService), accountRouter.accountController.GetAvatar)
		accountRou.PATCH("/time-to-live", middlewares.AuthorizeJWT(jwtService), accountRouter.accountController.UpdateTimeToLiveHardDelete)
		accountRou.GET("/export/excel", middlewares.AuthorizeJWT(jwtService), accountRouter.accountController.DownloadAccountsExcel)
	}
}
