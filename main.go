package main

import (
	"log"
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	database.ConnectDB()


	app := fiber.New()

	app.Use(logger.New()) 
    
   
    // app.Use(logger.New(logger.Config{
    //     Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
    // }))
	
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	routes.SetupRoutes(app)

	log.Fatal(app.Listen(":3001"))
}



