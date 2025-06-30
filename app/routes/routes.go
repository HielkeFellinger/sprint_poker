package routes

import (
	"github.com/gin-gonic/gin"
	"hielkefellinger.nl/sprint_poker/app/middleware"
	"hielkefellinger.nl/sprint_poker/app/pages"
	"hielkefellinger.nl/sprint_poker/app/session"
)

func HandlePageRoutes(router *gin.Engine) {
	// Room Routes
	roomRoutes := router.Group("/room")
	{
		// Create
		roomRoutes.GET("/create", middleware.EnsureUserValuesIsSet, pages.RoomCreate())
		roomRoutes.GET("/create/:room_id", middleware.EnsureUserValuesIsSet, pages.RoomCreate())
		roomRoutes.POST("/create", middleware.EnsureUserValuesIsSet, pages.RoomCreatePost())
		roomRoutes.POST("/create/:room_id", middleware.EnsureUserValuesIsSet, pages.RoomCreatePost())
		// Join
		roomRoutes.GET("/join/:room_id", middleware.EnsureUserAndRoomValuesAreSet, pages.RoomJoin())
		roomRoutes.POST("/join/:room_id", middleware.EnsureUserAndRoomValuesAreSet, pages.RoomJoinPost())
		// Session
		roomRoutes.GET("/:room_id/session", middleware.EnsureUserAndRoomValuesAreSetAndUserIsAuthenticated, pages.RoomSession())
		roomRoutes.GET("/:room_id/session/ws", middleware.EnsureUserAndRoomValuesAreSetAndUserIsAuthenticatedWs, session.ServeSessionWS)
	}

	// Pages
	router.GET("/", pages.Homepage())
}
