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

func (accountRouter *AccountRouter) Register(router *gin.RouterGroup) {
	jwtService := services.NewJwtService()
	accountRou := router.Group("/accounts", middlewares.AuthorizeJWT(jwtService))
	{
		accountRou.GET("/:id/detail", accountRouter.accountController.FindAccountById)
		accountRou.POST("/add", accountRouter.accountController.CreateAccount)
		accountRou.PATCH("/:id", accountRouter.accountController.UpdateAccount)
		accountRou.PATCH("/:id/restore", accountRouter.accountController.RestoreAccount)
		accountRou.PATCH("/:id/soft-delete", accountRouter.accountController.SoftDelete)
		accountRou.GET("/search", accountRouter.accountController.SearchAccount)
		accountRou.POST("/:id/update-avatar", accountRouter.accountController.UploadImageS3)
		accountRou.GET("/:id/avatar", accountRouter.accountController.GetAvatar)
		accountRou.PATCH("/time-to-live", accountRouter.accountController.UpdateTimeToLiveHardDelete)
		accountRou.GET("/export/excel", accountRouter.accountController.DownloadAccountsExcel)
		accountRou.POST("/:id/forgot-password", accountRouter.accountController.RestorePassword)
		accountRou.GET("/my-detail", accountRouter.accountController.MyDetails)
	}
}
