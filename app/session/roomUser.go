package session

import (
	"github.com/gorilla/websocket"
	"log"
	"math"
)

type RoomUser interface {
	GetId() string
	GetName() string
}

type roomUser struct {
	Id      string
	Name    string
	Session *roomSession
	Conn    *websocket.Conn
}

func (ru *roomUser) GetId() string {
	return ru.Id
}

func (ru *roomUser) GetName() string {
	return ru.Name
}

func (ru *roomUser) Read() {
	defer func() {
		// Disconnect from Room Session
		if ru.Session != nil {
			ru.Session.Unregister <- ru
			ru.Session = nil
		}
		// Close Connection
		err := ru.Conn.Close()
		if err != nil {
			log.Println(err)
			return
		}
	}()
	for {
		// Test availability Room Session and Connection
		if ru.Session == nil || ru.Conn == nil {
			return
		}

		// User input unsafe!
		_, rawEvent, err := ru.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// TODO remove logging
		maxLength := int(math.Min(float64(len(rawEvent)), 1024))
		log.Println(string(rawEvent[0:maxLength]))

		if ru.Session != nil {
			ru.Session.Events <- string(rawEvent)
		}
	}
}
