package main

import (
	"UserManagementVer/app"
	configs "UserManagementVer/configs"
	"UserManagementVer/db"
	"fmt"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	configs.LoadFileConfig()
	//Connect db
	Db := db.ConnectMongo(configs.AppConfig.Database.URI, configs.AppConfig.Database.Name)
	//Connect redis
	rdb := db.NewRedisClient()

	application := app.NewApplication(configs.AppConfig, Db, rdb)
	if err := application.Run(); err != nil {
		fmt.Println(err)
	}
}
