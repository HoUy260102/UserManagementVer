package main

import (
	configs "UserManagementVer/configs"
	"UserManagementVer/db"
	"UserManagementVer/routers"
	"fmt"

	"github.com/gin-gonic/gin"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	configs.LoadFileConfig()
	Db := db.ConnectMongo(configs.AppConfig.Database.URI, configs.AppConfig.Database.Name)
	rdb := configs.NewRedisClient()
	r := gin.Default()
	v1 := r.Group("/api/v1")
	routers.RegisterRouters(Db, v1, rdb)
	r.Run(fmt.Sprintf(":%d", configs.AppConfig.Server.Port))
}
