package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"hielkefellinger.nl/sprint_poker/app/initializers"
	"hielkefellinger.nl/sprint_poker/app/routes"
	"log"
	"os"
)

var engine *gin.Engine

func init() {
	log.Println("INIT: Starting Initialisation of Sprint-Poker")
	initializers.LoadEnvVariables()
	log.Println("INIT: Done. Initialisation Finished")
}

func main() {
	loadGinEngine()

	// Serve Content
	log.Println("MAIN: Starting Gin.Engine")
	log.Fatal(engine.Run(fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))))
}

func loadGinEngine() {
	log.Println("MAIN: Creation of Gin.Engine")
	engine = gin.Default()
	// Load Routes and (static) content
	log.Println("MAIN: Loading (Static) Content, Templates and Routes")
	routes.HandlePageRoutes(engine)
}
