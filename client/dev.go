package client

import (
	"errors"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

func parseDevMnt(sp string) (uint16, error) {
	switch sp {
	case "#r":
		return nine.DevRamFs, nil
	case "#c":
		return nine.DevCons, nil
	case "#n":
		return nine.DevNet, nil
	default:
		return 0, errors.New("illegal special mount")
	}
}
