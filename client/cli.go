package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

// RunClient spins a very simple
// CLI, useful for debugging the 9P
// server.
func RunCli(c *nine.Conf) {
	address := fmt.Sprintf("localhost:%d", c.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Failed to dial local server:", err)
	}
	// Run forever
	rdr := bufio.NewReader(os.Stdin)
	for {
		var cmd string
		fmt.Print("> ")
		cmd, err := rdr.ReadString('\n')
		cmd = strings.TrimRight(cmd, "\n")
		if err != nil || len(cmd) == 0 {
			continue
		}

		strArray := strings.Split(cmd, " ")
		switch strArray[0] {
		case "Tversion":
			if len(strArray) != 3 {
				fmt.Println("incorrect parameters")
				continue
			}

			msize, err := strconv.Atoi(strArray[1])
			if msize > int(^uint32(0)) || msize < 0 {
				fmt.Println("msize is out of bounds")
				continue
			} else if err != nil {
				fmt.Println("syntax error: msize should be a number")
				continue
			}

			version := strArray[2]
			if len(version) > int(^uint16(0)) {
				fmt.Println("Version is out of bounds")
				continue
			}

			fVersion(&conn, uint32(msize), version)

		//size[4] Tattach tag[2] fid[4] afid[4] uname[s] aname[s]
		case "Tattach":
			if len(strArray) != 3 {
				fmt.Println("incorrect parameters")
				continue
			}

			fid, err := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fmt.Println("fid is out of bounds")
				continue
			} else if err != nil {
				fmt.Println("syntax error: fid should be a number")
				continue
			}

			uname := strArray[2]
			if len(uname) > int(^uint16(0)) {
				fmt.Println("name is out of bounds")
				continue
			}

			fAttach(&conn, nine.Fid(fid), uname)

		case "Topen":
			if len(strArray) != 4 {
				fmt.Println("incorrect parameters")
				continue
			}

			fid, err1 := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			mode, err2 := strconv.Atoi(strArray[2])
			if mode > int(^byte(0)) || mode < 0 {
				fmt.Println("mode is out of bounds")
				continue
			} else if err1 != nil || err2 != nil {
				fmt.Println("syntax error: somewhere an integer conv failed")
				continue
			}

			fOpen(&conn, nine.Fid(fid), byte(mode))

		case "Twalk":
			if len(strArray) < 3 {
				fmt.Println("incorrect parameters")
				continue
			}

			fid, err1 := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			nfid, err2 := strconv.Atoi(strArray[2])
			if nfid > int(^uint32(0)) || nfid < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			if err1 != nil || err2 != nil {
				fmt.Println("syntax error somewhere")
				continue
			}

			if len(strArray) == 3 {
				fWalk(&conn, nine.Fid(fid), nine.Fid(nfid), []string{})
			} else {
				ns := strArray[3:]
				ok := true
				for _, nm := range ns {
					if len(nm) > int(^uint16(0)) {
						fmt.Println("walk name too long")
						ok = false
						break
					}
				}

				if !ok {
					continue
				}
				fWalk(&conn, nine.Fid(fid), nine.Fid(nfid), ns)
			}

		case "Tcreate":
			//size[4] Tcreate tag[2] fid[4] name[s] perm[4] mode[1]

			if len(strArray) != 5 {
				fmt.Println("incorrect parameters")
				continue
			}

			fid, err1 := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			name := strArray[2]
			if len(name) > int(^uint16(0)) {
				fmt.Println("name too long")
				continue
			}

			perm, err2 := strconv.Atoi(strArray[3])
			if perm > int(^uint32(0)) || perm < 0 {
				fmt.Println("perm is out of bounds")
				continue
			}

			mode, err3 := strconv.Atoi(strArray[4])
			if mode > int(^byte(0)) || mode < 0 {
				fmt.Println("mode is out of bounds")
				continue
			}

			if err1 != nil || err2 != nil || err3 != nil {
				fmt.Println("syntax error")
				continue
			}

			fCreate(&conn, nine.Fid(fid), name, uint32(perm), byte(mode))

		case "Tread":

			if len(strArray) != 4 {
				fmt.Println("incorrect parameters")
				continue
			}

			fid, err1 := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			offset, err2 := strconv.Atoi(strArray[2])
			if uint64(offset) > ^uint64(0) || offset < 0 {
				fmt.Println("offset is out of bounds")
				continue
			}

			count, err3 := strconv.Atoi(strArray[3])
			if count > int(^uint32(0)) || count < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			if err1 != nil || err2 != nil || err3 != nil {
				fmt.Println("syntax error")
				continue
			}

			fRead(&conn, nine.Fid(fid), uint64(offset), uint32(count))

		case "Twrite":
			//size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]

			if len(strArray) < 5 {
				fmt.Println("incorrect parameters")
				continue
			}

			fid, err1 := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			offset, err2 := strconv.Atoi(strArray[2])
			if uint64(offset) > ^uint64(0) || offset < 0 {
				fmt.Println("offset is out of bounds")
				continue
			}

			count, err3 := strconv.Atoi(strArray[3])
			if count > int(^uint32(0)) || count < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			if err1 != nil || err2 != nil || err3 != nil {
				fmt.Println("syntax error")
				continue
			}

			data := strings.Join(strArray[4:], " ")
			if len(data) > int(^uint16(0)) {
				fmt.Println("data too long")
				continue
			}

			fWrite(&conn, nine.Fid(fid), uint64(offset), data)

		case "Tclunk":
			//size[4] Tclunk tag[2] fid[4]

			if len(strArray) != 3 {
				fmt.Println("incorrect parameters")
				continue
			}

			fid, err := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			if err != nil {
				fmt.Println("syntax error")
				continue
			}

			fClunk(&conn, nine.Fid(fid))

		case "Tremove":

			if len(strArray) != 2 {
				fmt.Println("incorrect parameters")
				continue
			}

			fid, err := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			if err != nil {
				fmt.Println("syntax error")
				continue
			}

			fRemove(&conn, nine.Fid(fid))

		case "Tstat":

			if len(strArray) != 2 {
				fmt.Println("incorrect parameters")
				continue
			}

			fid, err := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fmt.Println("fid is out of bounds")
				continue
			}

			if err != nil {
				fmt.Println("syntax error")
				continue
			}

			fStat(&conn, nine.Fid(fid))

		case "Twstat":
			fmt.Println("UNSUPPORTED BY THE CLI :(")

		default:
			fmt.Println("Command not recognized big sad")

		}
	}
}
