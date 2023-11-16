package main

import (
	"laporinchat/models"

	"laporinchat/router"
)

func main() {

	models.ConnectDataBase()

	r := router.SetupRouter()
	router.SetupRoutes(r)

	r.Run(":8888")
}
