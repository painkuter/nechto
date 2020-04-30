package app

import (
	"sync"

	"github.com/gorilla/websocket"
)

const (
	roomUrlLength = 8
	maxPlayers    = 10
)

type room struct {
	name         string
	deck         []string
	trash        []string
	players      []*player
	mutex        sync.Mutex
	message      chan string
	positions    []*position
	nullPosition *position // используется для заполнения по циклу
}

func newRoom(name string) *room {
	return &room{
		name:         name,
		mutex:        sync.Mutex{},
		message:      make(chan string),
		nullPosition: &position{},
	}
}

func (r *room) addPlayer(name string, ws *websocket.Conn) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	newPlayer := &player{
		name:  name,
		mutex: sync.Mutex{},
		ws:    ws,
	}
	r.takePosition(newPlayer)
	r.players = append(r.players, newPlayer)
	if len(r.players) >= maxPlayers {
		r.start()
	}
}

func (r *room) takePosition(newPlayer *player) {
	p := r.lastPosition()
	newPosition := &position{right: p, player: newPlayer, number: len(r.positions)}
	p.left = newPosition
	r.positions = append(r.positions, newPosition)
	newPlayer.position = newPosition
}

func (r *room) lastPosition() *position {
	currentPosition := r.nullPosition
	for {
		if currentPosition.left != nil {
			currentPosition = currentPosition.left
		} else {
			break
		}
	}
	return currentPosition
}

func (r *room) start() {
	last := r.lastPosition()
	last.left = r.nullPosition.left
	r.nullPosition.left.right = last

	r.sendForPlayers("test")
}

func (r *room) sendForPlayers(message interface{}) { // TODO use wait group
	wg := sync.WaitGroup{}
	for _, p := range r.players {
		wg.Add(1)
		go func(player2 *player) {
			defer wg.Done()
			player2.wsMessage(message)
		}(p)
	}
	wg.Wait()
}
