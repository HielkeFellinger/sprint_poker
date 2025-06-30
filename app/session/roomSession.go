package session

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"hielkefellinger.nl/sprint_poker/app/models"
	"hielkefellinger.nl/sprint_poker/app/views/components"
	"log"
	"slices"
	"sort"
	"strings"
	"time"
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
	userIdToRoomUser     map[string]*roomUser
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
		userIdToRoomUser:     make(map[string]*roomUser),
		Room:                 room,
		RoomSessionState:     newRoomSessionState(),
	}
	return rs
}

func (rs *roomSession) Run() {
	checkIfEmptyTimer := time.NewTicker(180 * time.Second)
	defer checkIfEmptyTimer.Stop()
	for {
		select {
		case user := <-rs.Register:
			rs.users[user] = true

			// Update Room Session State
			rs.addUserGuessOfRoomUser(*user)
			rs.userIdToRoomUser[user.Id] = user

			rs.sendUpdatedUserGuessBoxToAllUsers()
		case user := <-rs.Unregister:
			delete(rs.users, user)
			delete(rs.UserIdToAlias, user.Id)
			delete(rs.userIdToRoomUser, user.Id)
			delete(rs.RoomSessionState.UserIdToGuess, user.Id)
			// Optionally, add Recover lead leaving; allow for reconnection

			// Close Room if empty
			if len(rs.users) == 0 {
				runningRoomPool.Unregister <- rs
				return
			}
		case requestMsg := <-rs.Events:
			switch requestMsg.Type {
			case Guess:
				log.Println("New Guess Message")
				err := rs.updateGuessOfUser(requestMsg)
				if err != nil {
					log.Printf("Error updating Guess: '%s'", err.Error())
					break
				}
				// Send the new User Guess Card Box to the users over websocket
				rs.sendUpdatedUserGuessBoxToAllUsers()
			case ToggleCardVisibility:
				log.Println("New ToggleCardVisibility Message")
				rs.RoomSessionState.CardsVisible = !rs.RoomSessionState.CardsVisible
				// Send the new User Guess Card Box to the users over websocket
				rs.sendUpdatedUserGuessBoxToAllUsers()
			case RefreshRound:
				log.Println("New RefreshRound Message")

				// Clear & Update Visibility
				rs.RoomSessionState.CardsVisible = false
				for id, _ := range rs.RoomSessionState.UserIdToGuess {
					if rs.RoomSessionState.UserIdToGuess[id] != nil {
						rs.RoomSessionState.UserIdToGuess[id].Card = nil
					}
				}
				// Send the new User Guess Card Box to the users over websocket
				rs.sendUpdatedUserGuessBoxToAllUsers()
			default:
				log.Printf("Unknown Message Type '%s'", requestMsg.Type)
			}
		case <-checkIfEmptyTimer.C:
			// Close Room if empty after timeout
			if len(rs.users) == 0 {
				log.Printf("Removing Room Session (empty after timeout): '%s'", rs.Id)
				runningRoomPool.Unregister <- rs
				return
			}
		}
	}
}

func (rs *roomSession) sendUpdatedUserGuessBoxToAllUsers() {
	for user, connected := range rs.users {
		if !connected || user.Conn == nil {
			log.Printf("Could not use connection of user: '%s'", user.Id)
			continue
		}
		rs.sendUpdatedUserGuessOfRoomUser(user)
	}
}

func (rs *roomSession) sendUpdatedUserGuessOfRoomUser(user *roomUser) {
	// Writer and Context Setup
	writer, nextWriterErr := user.Conn.NextWriter(websocket.TextMessage)
	if nextWriterErr != nil {
		log.Printf("Could not get connection writer for user: '%s' With Error : '%s'", user.Id, nextWriterErr.Error())
		return
	}
	ctx, timeOutCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer timeOutCancel()

	// Render
	renderErr := components.UsersGuessBox(rs.RoomSessionState.CardsVisible,
		models.User{Id: user.Id, Name: user.Name}, rs.GetAllUserGuesses()).Render(ctx, writer)
	if renderErr != nil {
		log.Printf("Could not render sendUpdatedUserGuessBoxToAllUsers Write for user: '%s' With Error : '%s'", user.Id, renderErr.Error())
	}

	// Close
	if closeErr := writer.Close(); closeErr != nil {
		log.Printf("Could not close sendUpdatedUserGuessBoxToAllUsers Write for user: '%s' With Error : '%s'", user.Id, closeErr.Error())
	}
}

func (rs *roomSession) updateGuessOfUser(requestMsg requestMessage) error {
	user, userAvail := rs.userIdToRoomUser[requestMsg.UserId]
	if !userAvail {
		return fmt.Errorf("could not find user: '%s' to update Guess", requestMsg.UserId)
	}

	userGuess, userGuessAvail := rs.RoomSessionState.UserIdToGuess[user.Id]
	if !userGuessAvail {
		return fmt.Errorf("no session info availible for user : '%s' to update Guess", requestMsg.UserId)
	}

	// Find matching Card of Room
	triggerCard := strings.Replace(requestMsg.Headers.HxTrigger, "card_", "", 1)
	matchUpdated := false
	for _, card := range rs.Room.Cards {
		if card.Id == triggerCard {
			userGuess.Card = &card
			rs.RoomSessionState.UserIdToGuess[user.Id] = userGuess
			matchUpdated = true
			break
		}
	}

	if !matchUpdated {
		return fmt.Errorf("invalid card: '%s' of user: '%s'", user.Id, requestMsg.UserId)
	}
	return nil
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

	// Sort to keep consistency Based on name
	sort.Slice(guesses, func(i, j int) bool {
		if guesses[i].User == nil {
			return true
		} else if guesses[j].User == nil {
			return false
		}
		return guesses[i].User.Name < guesses[j].User.Name
	})
	// Sort by card value (if cards are visible). Name-first sorting is still nice for sorting with same value
	if rs.RoomSessionState.CardsVisible {
		sort.Slice(guesses, func(i, j int) bool {
			if guesses[i].Card == nil {
				return false
			} else if guesses[j].Card == nil {
				return true
			}
			return guesses[i].Card.Order < guesses[j].Card.Order
		})
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
