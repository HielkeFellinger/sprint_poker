package session

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

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
		_, rawRequestMessage, readErr := ru.Conn.ReadMessage()
		if readErr != nil {
			log.Printf("Error: could not read request message of user: '%s'", ru.Id)
			return
		}

		// Attempt to parse-message
		var requestMsg requestMessage
		marshalErr := json.Unmarshal(rawRequestMessage, &requestMsg)
		if marshalErr != nil {
			log.Println(string(rawRequestMessage))
			log.Printf("Error: could not parse request message of user: '%s'", ru.Id)
			log.Printf("    - Parse error: '%s'", marshalErr.Error())
			return
		}

		// Enrich Info
		requestMsg.UserId = ru.Id

		if ru.Session != nil {
			ru.Session.Events <- requestMsg
		}
	}
}
