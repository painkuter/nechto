package app

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"

	"nechto/internal/pkg/apperr"
	"nechto/internal/pkg/log"
)

type player struct {
	name     string
	position *position
	mutex    sync.Mutex
	ws       *websocket.Conn
	room     *room
	cards    map[string]struct{}
}

func (p *player) wsMessage(message interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	text, err := json.Marshal(message)
	if err != nil {
		apperr.Check(err)
	}
	log.Infof("Sending message %s to %s\n", text, p.name)
	err = p.ws.WriteMessage(websocket.TextMessage, text)
	apperr.Check(err)
}

type position struct {
	number int
	left   *position
	right  *position
	player *player
}
