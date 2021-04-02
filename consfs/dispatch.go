package consfs

import (
	"bufio"
	"math"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type dispatch struct {
	l      sync.Mutex
	maxCon uint
	c      chan *service
}

var dispatcher dispatch = dispatch{}

func getService() *service {
	return <-dispatcher.c
}

// HandleWsCon callback, gets set up asynchronously, use
// the channel to funnel new connections
func HandleWsCon(ws *websocket.Conn) {
	dispatcher.l.Lock()

	// Fix me later!
	myCon := dispatcher.maxCon
	dispatcher.maxCon++
	dispatcher.l.Unlock()

	srv := service{
		ws:     ws,
		conNum: myCon,
		rdr:    bufio.NewReader(ws),
	}

	dispatcher.c <- &srv
	time.Sleep(math.MaxInt64)
}

//func Dial(url_, protocol, origin string) (ws *Conn, err error)
// http.Handle("/", websocket.Handler(ourfxn))
// http.ListenAndServe...
