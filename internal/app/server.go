package app

import (
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/codemodus/parth"
	"github.com/gorilla/websocket"

	"nechto/internal/config"
	"nechto/internal/pkg/apperr"
	"nechto/internal/pkg/log"
)

type Server struct {
	rooms map[string]*room
}

func NewServer() *Server {
	return &Server{
		rooms: map[string]*room{"new": newRoom("new")},
	}
}

func (s *Server) Start() error {
	// rh := http.RedirectHandler("/room", 301) // TODO make redirect
	// http.Handle("/", rh)                     // Path to redirect to connect default room
	http.HandleFunc("/room", s.roomHandler) // Path to connect default room
	// http.HandleFunc("/room/", s.roomHandler) // Path to connect existed room
	// http.HandleFunc("/new-room", newRoomHandler)     // Path to create new room -> redirecting to /room/[URL]
	http.HandleFunc("/ws", s.WsHandler) // WebSocket handler
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/view/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	log.Infof("Handlers initialized. Serve listening on: %s", config.ADDR)
	if err := http.ListenAndServe(config.ADDR, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
	return nil
}

func (s *Server) ActiveRoom() *room {
	if len(s.rooms) == 0 {
		panic("no rooms")
	}
	for _, r := range s.rooms {
		return r
	}
	return nil
}

// просто отображает страничку
func (s *Server) roomHandler(w http.ResponseWriter, r *http.Request) {
	roomUrl, err := parth.SegmentToString(r.URL.Path, -1)
	apperr.Check(err)
	room := s.ActiveRoom()

	if roomUrl == "room" {
		roomUrl = s.ActiveRoom().name
	} else {
		if len(roomUrl) != roomUrlLength {
			log.Info("Wrong room-Url")
			http.Error(w, "Room not found", 404)
			return
		}
	}

	playerName, err := parth.SegmentToString(r.URL.Path, 0)
	apperr.Check(err)
	log.Infof(playerName)

	var homeTempl = template.Must(template.ParseFiles("./view/index.html"))
	data := roomResponse{r.Host, roomUrl, len(room.players) + 1}
	err = homeTempl.Execute(w, data)
	apperr.Check(err)
}

type roomResponse struct {
	Host     string `json:"host"`
	RoomName string `json:"room_name"`
	Players  int    `json:"players"`
}

func (s *Server) WsHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		HandshakeTimeout: time.Second,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin:      websocket.IsWebSocketUpgrade,
	}
	ws, err := upgrader.Upgrade(w, r, http.Header{"Set-Cookie": {"sessionID=1234"}}) // fixme
	if e, ok := err.(websocket.HandshakeError); ok {
		log.Infof("Websocket handshake error: %s", e.Error())
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Error(err)
		http.Error(w, "Runtime error", 500)
		return
	}
	// TODO get url from request

	playerName := getPlayerName(r, len(s.ActiveRoom().players))
	log.Infof("Player %s has joined to room %s", playerName, s.ActiveRoom().name)

	s.ActiveRoom().addPlayer(playerName, ws)
}

func getPlayerName(r *http.Request, playersCount int) string {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Error("Error getting player name: ", err)
	}
	if len(params["name"]) > 0 {
		return params["name"][0]
	}
	return "Anonymous Player #" + strconv.Itoa(playersCount+1)
}
