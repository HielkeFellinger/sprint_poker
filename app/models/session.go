package models

type PublicRoomState struct {
	CardsVisible bool        `json:"cardsvisible"`
	UserGuesses  []UserGuess `json:"user_guesses"`
}

type UserGuess struct {
	User *User `json:"user"`
	Card *Card `json:"card"`
}
