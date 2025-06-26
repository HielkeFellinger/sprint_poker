package pages

import (
	"github.com/gin-gonic/gin"
	"hielkefellinger.nl/sprint_poker/app/models"
	"hielkefellinger.nl/sprint_poker/app/views"
	"log"
	"net/http"
)

func Homepage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var notifications = make([]models.Notification, 0)
		rawNotification, hasNotification := c.Get("notification")
		if hasNotification {
			notification := rawNotification.(models.Notification)
			notifications = append(notifications, notification)
		}

		log.Printf("rawNotification: %v", rawNotification)

		err := render(c, http.StatusOK, views.Homepage(notifications))
		if err != nil {
			return
		}
	}
}
