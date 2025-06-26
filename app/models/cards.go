package models

type Card struct {
	Order    uint   `json:"order"`
	Id       string `json:"id"`
	Value    string `json:"value"`
	Selected bool
}

func defaultCardSet() []Card {
	return []Card{
		{Order: 0, Id: "question mark", Value: "❓"},
		{Order: 1, Id: "drink", Value: "☕"},
		{Order: 2, Id: "0", Value: "0"},
		{Order: 3, Id: "0.5", Value: "0.5"},
		{Order: 4, Id: "1", Value: "1"},
		{Order: 5, Id: "2", Value: "2"},
		{Order: 6, Id: "3", Value: "3"},
		{Order: 7, Id: "5", Value: "5"},
		{Order: 8, Id: "8", Value: "8"},
		{Order: 9, Id: "13", Value: "13"},
		{Order: 10, Id: "20", Value: "20"},
		{Order: 11, Id: "40", Value: "40"},
		{Order: 12, Id: "infinity", Value: "∞"},
	}
}
