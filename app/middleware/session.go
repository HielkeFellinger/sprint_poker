package middleware

import (
	"github.com/gin-gonic/gin"
	"hielkefellinger.nl/sprint_poker/app/models"
	"hielkefellinger.nl/sprint_poker/app/session"
	"hielkefellinger.nl/sprint_poker/helpers"
	"log"
	"net/http"
)

func EnsureUserAndRoomValuesAreSetAndUserIsAuthenticated(c *gin.Context) {
	user := ensureSessionCookieAndGetUpToDateUser(c)

	// Check if Room exists; if not, redirect to home
	roomId := c.Param("room_id")

	log.Println(c.Request.RequestURI)
	log.Printf("User ID: '%s'", user.Id)
	log.Printf("Room ID: '%s'", roomId)

	room, roomRetrievalErr := session.GetRoomRunningByRoomId(roomId)
	if roomRetrievalErr != nil {
		c.Set("notification", models.NewNotification(models.Error, "404 - Room does not exist"))
		c.Redirect(http.StatusFound, "/")
		c.Abort()
		return
	}

	// Check if user is allowed to access room; if not, ensure user joins fist
	if !session.IsUserIdAuthenticatedInRoomSession(user.Id, room.Id) {
		c.Set("notification", models.NewNotification(models.Error, "401 - Not Authenticated"))
		c.Redirect(http.StatusFound, "/room/join/"+roomId)
		c.Abort()
		return
	} else {
		user.Name = session.GetUserAliasFromSessionByUserIdAndRoomId(user.Id, room.Id)
	}

	c.Set("room", room)
	c.Set("user", user)
	c.Next()
}

func EnsureUserAndRoomValuesAreSet(c *gin.Context) {
	user := ensureSessionCookieAndGetUpToDateUser(c)

	// Check if Room exists; if not, redirect to home
	roomId := c.Param("room_id")
	room, roomRetrievalErr := session.GetRoomRunningByRoomId(roomId)
	if roomRetrievalErr != nil {
		c.Set("notification", models.NewNotification(models.Error, "404 - Room does not exist"))
		c.Redirect(http.StatusFound, "/")
		c.Abort()
		return
	} else {
		user.Name = session.GetUserAliasFromSessionByUserIdAndRoomId(user.Id, room.Id)
	}
	c.Set("user", user)
	c.Set("room", room)
	c.Next()
}

func EnsureUserValuesIsSet(c *gin.Context) {
	c.Set("user", ensureSessionCookieAndGetUpToDateUser(c))
	c.Next()
}

func ensureSessionCookieAndGetUpToDateUser(c *gin.Context) models.User {
	// Retrieve or recover SessionContent
	sessionContent, sessionError := helpers.ParseSessionCookie(c)
	if sessionError != nil {
		sessionContent = helpers.NewSessionCookieContent()
	}

	// Update cookie
	setCookieErr := helpers.SetSessionJWTCookie(sessionContent, c)
	if setCookieErr != nil {
		log.Printf("Could not set/update session cookie: '%v'", setCookieErr.Error())
	}

	// Return data
	return models.User{
		Id: sessionContent.ID,
	}
}
