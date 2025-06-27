package pages

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"hielkefellinger.nl/sprint_poker/app/models"
	"hielkefellinger.nl/sprint_poker/app/session"
	"hielkefellinger.nl/sprint_poker/app/views"
	"hielkefellinger.nl/sprint_poker/helpers"
	"html"
	"log"
	"net/http"
)

func RoomCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var notifications = make([]models.Notification, 0)
		rawNotification, hasNotification := c.Get("notification")
		if hasNotification {
			notification := rawNotification.(models.Notification)
			notifications = append(notifications, notification)
		}

		err := render(c, http.StatusOK, views.RoomCreate(notifications))
		if err != nil {
			return
		}
	}
}

func RoomCreatePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		notifications := make([]models.Notification, 0)
		user := c.MustGet("user").(models.User)
		room := models.NewRoom()

		// Validate User
		if _, uuidErr := helpers.ParseStringToUuid(user.Id); uuidErr != nil {
			noUserErr := models.NewNotification(models.Error, "Current user has no valid ID; please refresh session")
			notifications = append(notifications, noUserErr)
		} else {
			room.LeadId = user.Id
		}

		// Validate Request
		var createRoomReq createRoomRequest
		err := c.Bind(&createRoomReq)
		if err == nil {
			// - Check required fields
			if len(createRoomReq.RoomName) == 0 || len(createRoomReq.DisplayName) == 0 {
				missingDataErr := models.NewNotification(models.Error, "Missing Required Room and/or User name")
				notifications = append(notifications, missingDataErr)
			} else {
				user.Name = html.EscapeString(createRoomReq.DisplayName)
				room.Name = html.EscapeString(createRoomReq.RoomName)
			}
			// Check passwords
			room.Private = false
			if len(createRoomReq.Password) > 0 || len(createRoomReq.PasswordCheck) > 0 {
				if createRoomReq.Password != createRoomReq.PasswordCheck {
					passwordMismatchErr := models.NewNotification(models.Error, "Passwords do not match")
					notifications = append(notifications, passwordMismatchErr)
				} else {
					hashedPass, hashErr := helpers.HashPassword(createRoomReq.Password)
					if hashErr != nil {
						hashAttemptErr := models.NewNotification(models.Error, "Passwords do not match")
						notifications = append(notifications, hashAttemptErr)
					} else {
						room.Private = true
						room.Password = string(hashedPass)
					}
				}
			}
		} else {
			noBindErr := models.NewNotification(models.Error, "Could not parse request, please retry")
			notifications = append(notifications, noBindErr)
		}

		// Create the room and its session if no errors have been found
		if len(notifications) == 0 {
			if sessionCreationErr := session.CreateNewRoomSession(room, user); sessionCreationErr != nil {
				sessionCreationErrNotification := models.NewNotification(models.Error, "Failure while creating session: '"+
					sessionCreationErr.Error()+"'")
				notifications = append(notifications, sessionCreationErrNotification)
			}
		}

		if len(notifications) > 0 {
			renderErr := render(c, http.StatusOK, views.RoomCreate(notifications))
			if renderErr != nil {
				log.Println(renderErr)
			}
			return
		} else {
			// Redirect (After creating a successful user)
			c.Redirect(http.StatusFound, "/room/"+room.Id+"/session")
		}
	}
}

type createRoomRequest struct {
	DisplayName   string `form:"displayName"`
	RoomName      string `form:"roomName"`
	Password      string `form:"password"`
	PasswordCheck string `form:"passwordCheck"`
}

func RoomJoin() gin.HandlerFunc {
	return func(c *gin.Context) {
		notifications := make([]models.Notification, 0)
		rawRoom, hasRoom := c.Get("room")
		rawUser, hasUser := c.Get("user")

		// Only Check for room; allow a user to change its display name
		if !hasRoom {
			c.Set("notification", models.NewNotification(models.Error, "404 - Room does not exist"))
			c.Redirect(http.StatusFound, "/")
			return
		}
		room := rawRoom.(models.Room)
		if !session.IsRoomSessionRunning(room.Id) {
			c.Set("notification", models.NewNotification(models.Error, "404 - Room is currently not running"))
			c.Redirect(http.StatusFound, "/")
			return
		}

		// Update user with project settings
		user := models.User{}
		if !hasUser {
			user = rawUser.(models.User)
			user.Name = session.GetUserAliasFromSessionByUserIdAndRoomId(user.Id, room.Id)
		}

		renderErr := render(c, http.StatusOK, views.RoomJoin(room, user, notifications))
		if renderErr != nil {
			log.Println(renderErr)
		}
		return
	}
}

func RoomJoinPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		notifications := make([]models.Notification, 0)
		room := c.MustGet("room").(models.Room)
		user := c.MustGet("user").(models.User)

		// Validate User
		if _, uuidErr := helpers.ParseStringToUuid(user.Id); uuidErr != nil {
			noUserErr := models.NewNotification(models.Error, "Current user has no valid ID; please refresh session")
			notifications = append(notifications, noUserErr)
		}

		// Validate Request
		var joinRoomReq joinRoomRequest
		err := c.Bind(&joinRoomReq)
		if err != nil {
			noBindErr := models.NewNotification(models.Error, "Could not parse request, please retry")
			notifications = append(notifications, noBindErr)
		}
		if len(joinRoomReq.DisplayName) == 0 {
			missingDataErr := models.NewNotification(models.Error, "Missing Required User Display name")
			notifications = append(notifications, missingDataErr)
		} else {
			user.Name = joinRoomReq.DisplayName
		}

		// Try Authentication (if needed)
		if room.Private {
			// bcrypt.CompareHashAndPassword will only take time if the hash is valid
			if errBcrypt := bcrypt.CompareHashAndPassword([]byte(room.Password), []byte(joinRoomReq.Password)); errBcrypt != nil {
				missingDataErr := models.NewNotification(models.Error, "Authentication to room failed, missing or incorrect password")
				notifications = append(notifications, missingDataErr)
			}
		}

		// Add user if no notifications are set
		if len(notifications) != 0 {
			renderErr := render(c, http.StatusOK, views.RoomJoin(room, user, notifications))
			if renderErr != nil {
				log.Println(renderErr)
			}
			return
		} else {
			session.AddOrUpdateUserToRoomSession(room, user)
			// Redirect (After creating a successful user)
			c.Redirect(http.StatusFound, "/room/"+room.Id+"/session")
		}
	}
}

type joinRoomRequest struct {
	DisplayName string `form:"displayName"`
	Password    string `form:"password"`
}

func RoomSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		notifications := make([]models.Notification, 0)
		user := c.MustGet("user").(models.User)
		room := c.MustGet("room").(models.Room)

		// Validate User
		if _, uuidErr := helpers.ParseStringToUuid(user.Id); uuidErr != nil {
			noUserErr := models.NewNotification(models.Error, "Current user has no valid ID; please refresh session")
			notifications = append(notifications, noUserErr)
		}

		// Test if user is Authenticated
		if !session.IsUserIdAuthenticatedInRoomSession(user.Id, room.Id) {
			c.Set("notification", models.NewNotification(models.Error, "401 - Not Authenticated"))
			c.Redirect(http.StatusFound, "/room/join/"+room.Id)
			return
		}

		roomState := session.GetPublicRoomStateByRoomId(room.Id)
		renderErr := render(c, http.StatusOK, views.RoomSession(user, roomState, room, notifications))
		if renderErr != nil {
			log.Println(renderErr)
			return
		}
	}
}
