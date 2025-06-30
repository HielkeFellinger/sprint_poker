package session

import (
	"errors"
	"hielkefellinger.nl/sprint_poker/app/models"
	"hielkefellinger.nl/sprint_poker/helpers"
	"log"
)

var runningRoomPool = newRoomPool()

type roomPool struct {
	Register   chan *roomSession
	Unregister chan *roomSession
	Sessions   map[*roomSession]bool
}

func newRoomPool() *roomPool {
	rp := &roomPool{
		Register:   make(chan *roomSession),
		Unregister: make(chan *roomSession),
		Sessions:   make(map[*roomSession]bool),
	}
	go rp.Run()
	return rp
}

func (pool *roomPool) Run() {
	for {
		select {
		case session := <-pool.Register:
			pool.Sessions[session] = true
			log.Printf("POOL: New Room Session: '%s'", session.Id)
		case session := <-pool.Unregister:
			if _, ok := pool.Sessions[session]; ok {
				delete(pool.Sessions, session)
				log.Printf("POOL: Removing Room Session: '%s'", session.Id)
			}
		}
	}
}

func CreateNewRoomSession(room models.Room, user models.User) error {
	log.Printf("RoomPool: Attempting to create new room session for room: '%s'", room.Id)
	// Checks
	if _, roomUuidErr := helpers.ParseStringToUuid(room.Id); roomUuidErr != nil {
		return errors.New("room Id is invalid, please refresh")
	}
	if room.LeadId != user.Id {
		return errors.New("room needs to be created with Lead user")
	}
	if _, userUuidErr := helpers.ParseStringToUuid(user.Id); userUuidErr != nil {
		return errors.New("user Id is invalid, please refresh")
	}
	if len(user.Name) == 0 {
		return errors.New("user (display)-name is not set, please set")
	}
	if IsRoomSessionRunning(room.Id) {
		return errors.New("room session already running")
	}

	// Create & Add Lead user
	roomSess := initRoomSession(room)
	roomSess.UpdateUserAsAuthenticatedIfNotAdded(user)
	go roomSess.Run()
	runningRoomPool.Register <- roomSess

	return nil
}

func AddOrUpdateUserToRoomSession(room models.Room, user models.User) bool {
	if rs := runningRoomPool.getRoomSessionById(room.Id); rs != nil {
		rs.UpdateUserAsAuthenticatedIfNotAdded(user)
		return true
	}
	return false
}

func IsUserIdAuthenticatedInRoomSession(userId string, roomId string) bool {
	if rs := runningRoomPool.getRoomSessionById(roomId); rs != nil {
		return rs.IsUserIdAuthenticatedInRoomSession(userId)
	}
	return false
}

func GetPublicRoomStateByRoomId(Id string) models.PublicRoomState {
	prs := models.PublicRoomState{
		CardsVisible: false,
		UserGuesses:  make([]models.UserGuess, 0),
	}

	if rs := runningRoomPool.getRoomSessionById(Id); rs != nil {
		prs.CardsVisible = rs.RoomSessionState.CardsVisible
		prs.UserGuesses = rs.GetAllUserGuesses()
	}

	return prs
}

func IsRoomSessionRunning(Id string) bool {
	return runningRoomPool.getRoomSessionById(Id) != nil
}

func GetUserAliasFromSessionByUserIdAndRoomId(userId string, roomId string) string {
	if rs := runningRoomPool.getRoomSessionById(roomId); rs != nil {
		return rs.GetUserAlias(userId)
	}
	return ""
}

func GetRoomRunningByRoomId(Id string) (models.Room, error) {
	if rs := runningRoomPool.getRoomSessionById(Id); rs != nil {
		return rs.Room, nil
	}
	return models.Room{}, errors.New("room not found")
}

func GetUsersAllowedInRoomByRoomId(Id string) []string {
	if rs := runningRoomPool.getRoomSessionById(Id); rs != nil {
		return rs.GetAllAuthorizedUserIds()
	}
	return make([]string, 0)
}

func (pool *roomPool) getRoomSessionById(Id string) *roomSession {
	for sessionSession, available := range pool.Sessions {
		if sessionSession != nil && available && sessionSession.Id == Id {
			return sessionSession
		}
	}
	return nil
}
