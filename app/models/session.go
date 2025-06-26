package models

type RoomState struct {
	CardsVisible bool        `json:"cards_visible"`
	Guesses      []UserGuess `json:"guesses"`
}

type UserGuess struct {
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
	CardId   string `json:"card_id"`
}
