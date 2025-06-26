package session

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"hielkefellinger.nl/sprint_poker/app/models"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// @TODO: SEC FAIL/DANGER THIS DOES BYPASS ORIGIN CHECK!!
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeSessionWS(c *gin.Context) {
	user := c.MustGet("user").(models.User)
	room := c.MustGet("room").(models.Room)

	log.Printf("Attempting to add WS user: '%s' to room: '%s'", user.Name, room.Name)

	// Room should be active and user should have joined
	runningRoomSession := runningRoomPool.getRoomSessionById(room.Id)
	if runningRoomSession == nil {
		// TODO no creation allowed
		return
	}

	// Upgrade Connection
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// TODO connection upgrade failed!
		log.Println(err)
		return
	}

	roomUsr := &roomUser{
		Id:      user.Id,
		Name:    runningRoomSession.GetUserAlias(user.Id),
		Conn:    ws,
		Session: runningRoomSession,
	}
	log.Printf("Registering WS user: '%s' to room: '%s'", user.Name, room.Name)
	runningRoomSession.Register <- roomUsr
	roomUsr.Read()
	log.Printf("Finished registering WS user: '%s' to room: '%s'", user.Name, room.Name)
}
