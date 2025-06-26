package models

import "github.com/google/uuid"

type Room struct {
	Id       string `json:"id"`
	LeadId   string `json:"lead_id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Cards    []Card `json:"cards"`
	Private  bool
}

func (room Room) DetermineAccess() {
	room.Private = len(room.Password) > 0
}

func NewRoom() Room {
	return Room{
		Id:      uuid.NewString(),
		Cards:   defaultCardSet(),
		Private: true,
	}
}
