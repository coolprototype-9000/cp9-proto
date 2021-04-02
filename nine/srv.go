package nine

import (
	"fmt"
	"log"
	"math/rand"
	"net"
)

type session struct {
	conn *net.Conn
}

func sessionMain(c *net.Conn, f FileSys, req chan<- *csFCall, resp <-chan *FCall) {
	s := session{conn: c}
	var id uint64 = 0

	for {
		tf, err := Read9P(s.conn)
		if err != nil {
			go func() {
				req <- &csFCall{
					f:  FCall{MsgType: TGoodbye},
					id: id,
				}
				(*c).Close()
			}()
		}

		// If the message is a version, we can handle it
		if tf.MsgType == TVersion {
			// Regenerate ID
			go func() {
				req <- &csFCall{
					f:  FCall{MsgType: TGoodbye},
					id: id,
				}
			}()

			for id == 0 {
				id = rand.Uint64()
			}
			resp := &FCall{
				MsgType: RVersion,
				Tag:     tf.Tag,
				Version: NineVersion,
			}

			if tf.Version != NineVersion {
				resp.Version = "unknown"
				id = 0
			}
			go Write9P(s.conn, resp)
			continue
		} else if id == 0 {
			resp := mkError(tf.Tag, "Uninitialized connection")
			go Write9P(s.conn, resp)
			continue
		}

		// Send the request, and wait for the resp
		// to complete. Once it is done, write back.
		go func() {
			req <- &csFCall{
				f:  tf,
				id: id,
			}
			out := <-resp
			Write9P(s.conn, out)
		}()
	}
}

func singleThreadedProcessor(req <-chan *csFCall, resp chan<- *FCall, f FileSys) {
	for r := range req {
		switch r.f.MsgType {
		case TGoodbye:
			f.Goodbye(r.id)
		case TAttach:
			q, err := f.Attach(r.id, r.f.F, r.f.Uname)
			if err != nil {
				resp <- mkError(r.f.Tag, err.Error())
			} else {
				rf := FCall{
					MsgType: RAttach,
					Tag:     r.f.Tag,
					Q:       q,
				}
				resp <- &rf
			}

		case TWalk:
			wq, err := f.Walk(r.id, r.f.F, r.f.Newf, r.f.Wname)
			if err != nil {
				resp <- mkError(r.f.Tag, err.Error())
			} else {
				rf := FCall{
					MsgType: RWalk,
					Tag:     r.f.Tag,
					Wqid:    wq,
				}
				resp <- &rf
			}

		case TCreate:
			q, err := f.Create(r.id, r.f.F, r.f.Name, r.f.Perm, r.f.Mode)
			if err != nil {
				resp <- mkError(r.f.Tag, err.Error())
			} else {
				rf := FCall{
					MsgType: RCreate,
					Tag:     r.f.Tag,
					Q:       q,
				}
				resp <- &rf
			}

		case TOpen:
			q, err := f.Open(r.id, r.f.F, r.f.Mode)
			if err != nil {
				resp <- mkError(r.f.Tag, err.Error())
			} else {
				rf := FCall{
					MsgType: ROpen,
					Tag:     r.f.Tag,
					Q:       q,
				}
				resp <- &rf
			}

		case TRead:
			data, err := f.Read(r.id, r.f.F, r.f.Offset, r.f.Count)
			if err != nil {
				resp <- mkError(r.f.Tag, err.Error())
			} else {
				rf := FCall{
					MsgType: RRead,
					Tag:     r.f.Tag,
					Data:    data,
				}
				resp <- &rf
			}

		case TWrite:
			count, err := f.Write(r.id, r.f.F, r.f.Offset, r.f.Data)
			if err != nil {
				resp <- mkError(r.f.Tag, err.Error())
			} else {
				rf := FCall{
					MsgType: RWrite,
					Tag:     r.f.Tag,
					Count:   count,
				}
				resp <- &rf
			}

		case TClunk:
			err := f.Clunk(r.id, r.f.F)
			if err != nil {
				resp <- mkError(r.f.Tag, err.Error())
			} else {
				rf := FCall{
					MsgType: RClunk,
					Tag:     r.f.Tag,
				}
				resp <- &rf
			}

		case TRemove:
			err := f.Remove(r.id, r.f.F)
			if err != nil {
				resp <- mkError(r.f.Tag, err.Error())
			} else {
				rf := FCall{
					MsgType: RRemove,
					Tag:     r.f.Tag,
				}
				resp <- &rf
			}

		case TStat:
			s, err := f.Stat(r.id, r.f.F)
			if err != nil {
				resp <- mkError(r.f.Tag, err.Error())
			} else {
				rf := FCall{
					MsgType: RStat,
					Tag:     r.f.Tag,
					St:      s,
				}
				resp <- &rf
			}

		case TWStat:
			err := f.Wstat(r.id, r.f.F, r.f.St)
			if err != nil {
				resp <- mkError(r.f.Tag, err.Error())
			} else {
				rf := FCall{
					MsgType: RWStat,
					Tag:     r.f.Tag,
				}
				resp <- &rf
			}

		default:
			resp <- mkError(r.f.Tag, "Unsupported operation!")

		}

		// Debugging!
	}
}

// ServeForever spins a 9P server using the passed
// configuration, and runs it until this process is killed
// i.e. indefinitely.
func ServeForever(c *Conf) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		log.Fatal("Failed to start backing TCP server (probably perm denied):", err)
	}

	// Set up the filesystem
	devNo := rand.Uint32()
	err = c.Fs.Register(devNo)
	if err != nil {
		log.Fatal("Failed to register filesystem: ", err)
	}

	// Spin the single threaded processor
	req := make(chan *csFCall)
	resp := make(chan *FCall)
	go singleThreadedProcessor(req, resp, c.Fs)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("Failed to accept client TCP con:", err)
		}

		// Pass each session the request/response channels
		go sessionMain(&conn, c.Fs, req, resp)
	}
}

// Pure utility....no better place to put this
func mkError(tag uint16, ename string) *FCall {
	rf := FCall{
		MsgType: RError,
		Tag:     tag,
		Ename:   ename,
	}
	return &rf
}
