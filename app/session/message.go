package session

import "github.com/gin-gonic/gin"

type MessageType string

const (
	Guess                MessageType = "guess"
	ToggleCardVisibility MessageType = "toggle-card-visibility"
	RefreshRound         MessageType = "refresh-round"
)

type requestMessage struct {
	UserId  string           `json:"-"`
	Context *gin.Context     `json:"-"`
	Type    MessageType      `json:"type"`
	Value   string           `json:"value"`
	Headers requestHxHeaders `json:"HEADERS"`
}

type requestHxHeaders struct {
	HxRequest     string `json:"HX-Request"`
	HxTrigger     string `json:"HX-Trigger"`
	HxTriggerName string `json:"HX-Trigger-Name"`
	HxTarget      string `json:"HX-Target"`
	HxCurrentUrl  string `json:"HX-Current-URL"`
}
