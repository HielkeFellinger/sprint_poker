package session

import (
	"hielkefellinger.nl/sprint_poker/app/models"
	"log"
	"slices"
)

type RoomSession interface {
	GetId() string
	GetLeaderId() string
	GetRoom() models.Room
	GetAllAuthorizedUserIds() []string
	GetAllUserIds(filterOut ...string) []string
	IsUserIdAuthenticatedInRoomSession(userId string) bool
	UpdateUserAsAuthenticatedIfNotAdded(user models.User)
	SetUserAlias(userId string, alias string)
	GetUserAlias(userId string) string
}
type roomSession struct {
	Id                   string
	LeaderId             string
	Register             chan *roomUser
	Unregister           chan *roomUser
	Events               chan string
	users                map[*roomUser]bool
	AuthenticatedUserIds []string
	UserIdToAlias        map[string]string
	Room                 models.Room
}

func initRoomSession(room models.Room) *roomSession {
	rs := &roomSession{
		Id:                   room.Id,
		LeaderId:             room.LeadId,
		Register:             make(chan *roomUser),
		Unregister:           make(chan *roomUser),
		Events:               make(chan string),
		users:                make(map[*roomUser]bool),
		AuthenticatedUserIds: make([]string, 0),
		UserIdToAlias:        make(map[string]string),
		Room:                 room,
	}
	return rs
}

func (rs *roomSession) Run() {
	for {
		select {
		case user := <-rs.Register:
			rs.users[user] = true
		case user := <-rs.Unregister:
			delete(rs.users, user)
			delete(rs.UserIdToAlias, user.Id)
			// TODO recover lead leaving; should jou always re-authenticate?
		case eventString := <-rs.Events:
			log.Println(eventString)
		}
	}
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
	return rs.AuthenticatedUserIds
}

func (rs *roomSession) UpdateUserAsAuthenticatedIfNotAdded(user models.User) {
	if !slices.Contains(rs.AuthenticatedUserIds, user.Id) {
		rs.AuthenticatedUserIds = append(rs.AuthenticatedUserIds, user.Id)
	}
	rs.SetUserAlias(user.Id, user.Name)
}

func (rs *roomSession) IsUserIdAuthenticatedInRoomSession(userId string) bool {
	return slices.Contains(rs.AuthenticatedUserIds, userId)
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
