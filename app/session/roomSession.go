package session

import (
	"context"
	"github.com/gorilla/websocket"
	"hielkefellinger.nl/sprint_poker/app/models"
	"hielkefellinger.nl/sprint_poker/app/views/components"
	"log"
	"slices"
)

type roomSessionState struct {
	CardsVisible  bool                         `json:"cards_visible"`
	UserIdToGuess map[string]*models.UserGuess `json:"-"`
}

func newRoomSessionState() roomSessionState {
	return roomSessionState{
		CardsVisible:  false,
		UserIdToGuess: make(map[string]*models.UserGuess),
	}
}

type roomSession struct {
	Id                   string
	LeaderId             string
	Register             chan *roomUser
	Unregister           chan *roomUser
	Events               chan requestMessage
	users                map[*roomUser]bool
	authenticatedUserIds []string
	UserIdToAlias        map[string]string
	Room                 models.Room
	RoomSessionState     roomSessionState
}

func initRoomSession(room models.Room) *roomSession {
	rs := &roomSession{
		Id:                   room.Id,
		LeaderId:             room.LeadId,
		Register:             make(chan *roomUser),
		Unregister:           make(chan *roomUser),
		Events:               make(chan requestMessage),
		users:                make(map[*roomUser]bool),
		authenticatedUserIds: make([]string, 0),
		UserIdToAlias:        make(map[string]string),
		Room:                 room,
		RoomSessionState:     newRoomSessionState(),
	}
	return rs
}

func (rs *roomSession) Run() {
	for {
		// TODO; add way to remove empty sessions
		select {
		case user := <-rs.Register:
			rs.users[user] = true

			// Update Room Session State
			rs.addUserGuessOfRoomUser(*user)

		case user := <-rs.Unregister:
			// Recover lead leaving; allow for reconnection
			delete(rs.users, user)
			delete(rs.UserIdToAlias, user.Id)
			delete(rs.RoomSessionState.UserIdToGuess, user.Id)
		case requestMsg := <-rs.Events:

			switch requestMsg.Type {
			case Guess:
				//
			case ToggleCardVisibility:
				//
			case RefreshRound:
				//
			}

			// Send the new User Guess Card Box to the users
			for user, connected := range rs.users {
				writer, err := user.Conn.NextWriter(websocket.TextMessage)
				if err != nil {
					log.Println(err)
					return
				}
				if connected {
					// TODO Cleanup (Proof of concept)

					ctx := context.Background()
					log.Printf("Rendering Guess Box for user %s", user.Id)
					renderErr := components.UsersGuessBox(rs.RoomSessionState.CardsVisible, models.User{Id: user.Id, Name: user.Name},
						rs.GetAllUserGuesses()).Render(ctx, writer)

					if renderErr != nil {
						log.Println(renderErr)
					}

					writer.Close()
				}
			}

		}
	}
}

func (rs *roomSession) addUserGuessOfRoomUser(user roomUser) {
	userGuess := &models.UserGuess{
		User: &models.User{
			Id:   user.Id,
			Name: user.Name,
		},
		Card: nil,
	}
	rs.RoomSessionState.UserIdToGuess[user.Id] = userGuess
}

func (rs *roomSession) GetAllUserGuesses() []models.UserGuess {
	guesses := make([]models.UserGuess, 0)
	for _, userGuess := range rs.RoomSessionState.UserIdToGuess {
		if userGuess != nil {
			guesses = append(guesses, *userGuess)
		}
	}
	return guesses
}

func (rs *roomSession) GetId() string {
	return rs.Id
}

func (rs *roomSession) GetLeaderId() string {
	return rs.LeaderId
}

func (rs *roomSession) GetRoom() models.Room {
	return rs.Room
}

func (rs *roomSession) SetUserAlias(userId string, alias string) {
	rs.UserIdToAlias[userId] = alias
}

func (rs *roomSession) GetAllAuthorizedUserIds() []string {
	return rs.authenticatedUserIds
}

func (rs *roomSession) UpdateUserAsAuthenticatedIfNotAdded(user models.User) {
	if !slices.Contains(rs.authenticatedUserIds, user.Id) {
		rs.authenticatedUserIds = append(rs.authenticatedUserIds, user.Id)
	}
	rs.SetUserAlias(user.Id, user.Name)
}

func (rs *roomSession) IsUserIdAuthenticatedInRoomSession(userId string) bool {
	return slices.Contains(rs.authenticatedUserIds, userId)
}

func (rs *roomSession) GetUserAlias(userId string) string {
	if alias, ok := rs.UserIdToAlias[userId]; ok {
		return alias
	}
	return ""
}

func (rs *roomSession) GetAllUserIds(filterOut ...string) []string {
	userIds := make([]string, 0)
	// Always add Lead
	if !slices.Contains(filterOut, rs.LeaderId) {
		userIds = append(userIds, rs.LeaderId)
	}
	for user := range rs.users {

		// Filter out if applicable
		if len(filterOut) > 0 && slices.Contains(filterOut, user.Id) {
			continue
		}
		userIds = append(userIds, user.Id)
	}
	return userIds
}
