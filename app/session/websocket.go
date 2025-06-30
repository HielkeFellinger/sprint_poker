package session

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"hielkefellinger.nl/sprint_poker/app/models"
	"log"
	"net/http"
)

const (
	maxMessageSize = 512
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

	// Room should be active and user should have joined
	runningRoomSession := runningRoomPool.getRoomSessionById(room.Id)
	if runningRoomSession == nil {
		c.Redirect(http.StatusFound, "/")
		c.Abort()
		return
	}

	// Upgrade Connection
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Could not upgrade connection of user: '%s' , message: '%s'", user.Id, err.Error())
		return
	}

	log.Printf("Websocket upgraded; Registering WS user: '%s' to room: '%s'", user.Id, room.Id)
	roomUsr := &roomUser{
		Id:      user.Id,
		Name:    runningRoomSession.GetUserAlias(user.Id),
		Conn:    ws,
		Session: runningRoomSession,
	}
	roomUsr.Conn.SetReadLimit(maxMessageSize)
	runningRoomSession.Register <- roomUsr
	// TODO Read / Write Pump in own routine + Ping Pong?
	go roomUsr.Read()
	log.Printf("Finished registering WS user: '%s' to room: '%s'", user.Id, room.Id)
}
