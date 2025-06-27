package models

type RoomSessionState struct {
	CardsVisible  bool                  `json:"cards_visible"`
	UserIdToGuess map[string]*UserGuess `json:"-"`
	Guesses       []UserGuess           `json:"guesses"`
}

func NewRoomSessionState() RoomSessionState {
	return RoomSessionState{
		CardsVisible:  false,
		UserIdToGuess: make(map[string]*UserGuess),
		Guesses:       make([]UserGuess, 0),
	}
}

type UserGuess struct {
	User *User `json:"user"`
	Card *Card `json:"card"`
}
