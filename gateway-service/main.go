package main

import (
	"fmt"
	"gateway_service/api/controllers"
	"gateway_service/api/controllers/sockets"
	"gateway_service/config"
	"gateway_service/internal"
	"gateway_service/routes"
	"log"
	"math/rand"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()

	serviceClient, err := internal.GetServiceClient(cfg.UserServiceAddr)
	if err != nil {
		log.Fatal("Failed to create user service client: ", err)
	}

	userController := controllers.NewUserController(serviceClient)
	sockerController := sockets.NewSocketController(serviceClient)
	guildController := controllers.NewGuildController(serviceClient)

	router := http.NewServeMux()

	routes.RegisterControllers(router, userController, sockerController, guildController)

	err = internal.InitKafka("consumer-" + fmt.Sprint(rand.Int()))
	if err != nil {
		log.Println(err)
	}

	log.Print("Server started on address: ", cfg.ServerPort)

	if err := http.ListenAndServe(cfg.ServerPort, router); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
