package pages

import (
	"github.com/gin-gonic/gin"
	"hielkefellinger.nl/sprint_poker/app/views"
	"net/http"
)

func Homepage() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := render(c, http.StatusOK, views.Homepage(getNotifications(c)))
		if err != nil {
			return
		}
	}
}
