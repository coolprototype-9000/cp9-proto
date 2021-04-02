package consfs

import (
	"bufio"
	"fmt"
	"strings"

	"golang.org/x/net/websocket"
)

const servicePort = 6969 // nice

type service struct {
	ws     *websocket.Conn
	conNum uint
	rdr    *bufio.Reader
}

func (s *service) readWS() (string, error) {
	line, err := s.rdr.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimRight(line, "\n")
	return line, nil
}

func (s *service) writeWS(str string) error {
	_, err := fmt.Fprintf(s.ws, "%s\n", str)
	return err
}
