package nine

import (
	"errors"
	"fmt"
	"io"
	"net"
)

// Read9P gets a 9P message off the wire, retrying
// until it is successful then de-marshaling it prior
// to return
func Read9P(c *net.Conn) (FCall, error) {
	szBuf := make([]byte, 4)
again:
	cnt, err := (*c).Read(szBuf)
	if err != nil {
		return FCall{}, err
	} else if cnt != 4 {
		fmt.Println("[DEBUG]: read9P: Erroneous size read")
		goto again
	}

	sz, _ := unmarshalUint32(szBuf)
	msgBuf := make([]byte, sz-4)
	cnt, err = io.ReadFull(*c, msgBuf)
	if err != nil {
		return FCall{}, err
	} else if cnt != len(msgBuf) {
		fmt.Printf("[DEBUG]: read9P: Erroneous body read. Got %d vs. %d\n", cnt, len(msgBuf))
		goto again
	}

	f := unmarshalFCall(append(szBuf, msgBuf...))
	return f, nil
}

// Write9P puts a 9P message on the wire, returning
// errors if something goes wrong along the way
func Write9P(c *net.Conn, f *FCall) error {
	bytes, err := marshalFCall(*f)
	if err != nil {
		return err
	}

	cnt, err := (*c).Write(bytes)
	if err != nil {
		return err
	} else if cnt != len(bytes) {
		return errors.New("incorrect count written to wire")
	}

	return nil
}
