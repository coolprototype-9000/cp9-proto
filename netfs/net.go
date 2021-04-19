package netfs

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type state byte

const (
	idle state = iota
	listening
	connected
	dead
)

type netInst struct {
	owner    string
	id       uint64
	children []*netPtr

	s  state
	c  *net.Conn
	cs string
	ln *net.Listener
}

func mkEmptyNetInst(as string, id uint64) *netInst {
	n := &netInst{
		owner:    as,
		id:       id,
		s:        idle,
		children: make([]*netPtr, 0),
	}

	tps := []ptrType{dir, ctl, data, listen}
	for _, tp := range tps {
		n.children = append(n.children, &netPtr{
			tp:      tp,
			version: 0,
			ats:     uint32(time.Now().Unix()),
			mts:     uint32(time.Now().Unix()),
			muid:    as,
		})
	}
	return n
}

func (ni *netInst) checkDead() bool {
	return ni.s == dead
}

func (ni *netInst) openCon() error {
	if ni.s != idle {
		return errors.New("incorrect state transition")
	} else if ni.cs == "" {
		return errors.New("need to enter a connection string to ctl")
	}

	nc, err := net.Dial("tcp", ni.cs)
	if err != nil {
		return err
	}
	ni.c = &nc
	ni.s = connected
	ni.cs = ""
	ni.ln = nil
	return nil
}

func (ni *netInst) enterListeningState() error {
	if ni.s != idle {
		return errors.New("incorrect state transition")
	} else if ni.cs == "" {
		return errors.New("need to enter a connection string to ctl")
	}
	ln, err := net.Listen("tcp", ni.cs)
	if err != nil {
		return err
	}

	ni.c = nil
	ni.s = listening
	ni.cs = ""
	ni.ln = &ln
	fmt.Printf("YEEEEEET : Entered listening state%s\n", ni.cs)
	return nil
}

func (ni *netInst) acceptCon() (*net.Conn, error) {
	if ni.s != listening {
		return nil, errors.New("you're not listening")
	}

	l, ok := (*ni.ln).(*net.TCPListener)
	if !ok {
		log.Fatal("WTF")
	}

	l.SetDeadline(time.Now().Add(time.Millisecond))
	c, err := (*ni.ln).Accept()
	fmt.Printf("YEEEEEET : Accepted connection%s\n", ni.cs)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
