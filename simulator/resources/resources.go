package resources

import (
	"errors"
	"sync"

	socketio "github.com/zishang520/socket.io/socket"
)

type Resources struct {
	ExitGroup       sync.WaitGroup             `json:"-"`
	WebSocket       *socketio.Socket           `json:"-"`
	LinkedDevSocket map[int](*socketio.Socket) `json:"-"`
	ConnDevSocket   map[string]int             `json:"-"`
	//	LinkedDevEUItoId map[string]int        `json:"-"`
}

func (r *Resources) AddWebSocket(WebSocket *socketio.Socket) {
	r.WebSocket = WebSocket
}

func (r *Resources) DevAddSocket(devSocket *socketio.Socket, Id int) {
	r.LinkedDevSocket[Id] = devSocket
	SId := (*devSocket).Id()
	r.ConnDevSocket[string(SId)] = Id
}

func (r *Resources) DevDeleteSocket(SId string) {
	if _, ok := r.ConnDevSocket[SId]; ok {
		delete(r.LinkedDevSocket, r.ConnDevSocket[SId])
		delete(r.ConnDevSocket, SId)
	}
}

func (r *Resources) DevGetIdFromSocketID(SId string) (int, error) {
	if id, ok := r.ConnDevSocket[SId]; ok {
		return id, nil
	} else {
		return -1, errors.New("no device associated with socketio SId")
	}
}
