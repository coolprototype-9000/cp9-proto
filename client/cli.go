package client

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

// RunClient spins a very simple
// CLI, useful for debugging the 9P
// server.
func RunCli(c *nine.Conf, r *nine.Conf) {
	cAddress := fmt.Sprintf("localhost:%d", c.Port)
	cc, err := net.Dial("tcp", cAddress)
	if err != nil {
		log.Fatal("Failed to dial local consfs:", err)
	}

	rAddress := fmt.Sprintf("localhost:%d", r.Port)
	rc, err := net.Dial("tcp", rAddress)
	if err != nil {
		log.Fatal("Failed to dial local consfs:", err)
	}

	terminal := setupConsFs(&cc)

	// Run forever
	for {
		fmt.Print("> ")
		call := fRead(&cc, terminal, 0, 9001)
		if call.MsgType != nine.RRead {
			continue
		}
		cmd := string(call.Data)
		if len(cmd) == 0 {
			continue
		}

		strArray := strings.Split(cmd, " ")
		switch strArray[0] {
		case "Tversion":
			if len(strArray) != 3 {
				fWrite(&cc, terminal, 0, "incorrect parameters")
				continue
			}

			msize, err := strconv.Atoi(strArray[1])
			if msize > int(^uint32(0)) || msize < 0 {
				fWrite(&cc, terminal, 0, "msize is out of bounds")
				continue
			} else if err != nil {
				fWrite(&cc, terminal, 0, "syntax error: msize should be a number")
				continue
			}

			version := strArray[2]
			if len(version) > int(^uint16(0)) {
				fWrite(&cc, terminal, 0, "Version is out of bounds")
				continue
			}

			res := fVersion(&rc, uint32(msize), version)
			fWrite(&cc, terminal, 0, fmt.Sprintf("%v", res))

		//size[4] Tattach tag[2] fid[4] afid[4] uname[s] aname[s]
		case "Tattach":
			if len(strArray) != 3 {
				fWrite(&cc, terminal, 0, "incorrect parameters")
				continue
			}

			fid, err := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			} else if err != nil {
				fWrite(&cc, terminal, 0, "syntax error: fid should be a number")
				continue
			}

			uname := strArray[2]
			if len(uname) > int(^uint16(0)) {
				fWrite(&cc, terminal, 0, "name is out of bounds")
				continue
			}

			res := fAttach(&rc, nine.Fid(fid), uname)
			fWrite(&cc, terminal, 0, fmt.Sprintf("%v", res))

		case "Topen":
			if len(strArray) != 3 {
				fWrite(&cc, terminal, 0, "incorrect parameters")
				continue
			}

			fid, err1 := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			}

			mode, err2 := strconv.Atoi(strArray[2])
			if mode > int(^byte(0)) || mode < 0 {
				fWrite(&cc, terminal, 0, "mode is out of bounds")
				continue
			} else if err1 != nil || err2 != nil {
				fWrite(&cc, terminal, 0, "syntax error: somewhere an integer conv failed")
				continue
			}

			res := fOpen(&rc, nine.Fid(fid), byte(mode))
			fWrite(&cc, terminal, 0, fmt.Sprintf("%v", res))

		case "Twalk":
			if len(strArray) < 3 {
				fWrite(&cc, terminal, 0, "incorrect parameters")
				continue
			}

			fid, err1 := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			}

			nfid, err2 := strconv.Atoi(strArray[2])
			if nfid > int(^uint32(0)) || nfid < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			}

			if err1 != nil || err2 != nil {
				fWrite(&cc, terminal, 0, "syntax error somewhere")
				continue
			}

			var res *nine.FCall
			if len(strArray) == 3 {
				res = fWalk(&rc, nine.Fid(fid), nine.Fid(nfid), []string{})
			} else {
				ns := strArray[3:]
				ok := true
				for _, nm := range ns {
					if len(nm) > int(^uint16(0)) {
						fWrite(&cc, terminal, 0, "walk name too long")
						ok = false
						break
					}
				}

				if !ok {
					continue
				}
				res = fWalk(&rc, nine.Fid(fid), nine.Fid(nfid), ns)
			}

			fWrite(&cc, terminal, 0, fmt.Sprintf("%v", res))

		case "Tcreate":
			//size[4] Tcreate tag[2] fid[4] name[s] perm[4] mode[1]

			if len(strArray) != 5 {
				fWrite(&cc, terminal, 0, "incorrect parameters")
				continue
			}

			fid, err1 := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			}

			name := strArray[2]
			if len(name) > int(^uint16(0)) {
				fWrite(&cc, terminal, 0, "name too long")
				continue
			}

			perm, err2 := strconv.Atoi(strArray[3])
			if perm > int(^uint32(0)) || perm < 0 {
				fWrite(&cc, terminal, 0, "perm is out of bounds")
				continue
			}

			mode, err3 := strconv.Atoi(strArray[4])
			if mode > int(^byte(0)) || mode < 0 {
				fWrite(&cc, terminal, 0, "mode is out of bounds")
				continue
			}

			if err1 != nil || err2 != nil || err3 != nil {
				fWrite(&cc, terminal, 0, "syntax error")
				continue
			}

			res := fCreate(&rc, nine.Fid(fid), name, uint32(perm), byte(mode))
			fWrite(&cc, terminal, 0, fmt.Sprintf("%v", res))

		case "Tread":

			if len(strArray) != 4 {
				fWrite(&cc, terminal, 0, "incorrect parameters")
				continue
			}

			fid, err1 := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			}

			offset, err2 := strconv.Atoi(strArray[2])
			if uint64(offset) > ^uint64(0) || offset < 0 {
				fWrite(&cc, terminal, 0, "offset is out of bounds")
				continue
			}

			count, err3 := strconv.Atoi(strArray[3])
			if count > int(^uint32(0)) || count < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			}

			if err1 != nil || err2 != nil || err3 != nil {
				fWrite(&cc, terminal, 0, "syntax error")
				continue
			}

			res := fRead(&rc, nine.Fid(fid), uint64(offset), uint32(count))
			fWrite(&cc, terminal, 0, fmt.Sprintf("%v (%s)", res, string(res.Data)))

		case "Twrite":
			//size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]

			if len(strArray) < 4 {
				fWrite(&cc, terminal, 0, "incorrect parameters")
				continue
			}

			fid, err1 := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			}

			offset, err2 := strconv.Atoi(strArray[2])
			if uint64(offset) > ^uint64(0) || offset < 0 {
				fWrite(&cc, terminal, 0, "offset is out of bounds")
				continue
			}

			if err1 != nil || err2 != nil {
				fWrite(&cc, terminal, 0, "syntax error")
				continue
			}

			data := strings.Join(strArray[3:], " ")
			if len(data) > int(^uint16(0)) {
				fWrite(&cc, terminal, 0, "data too long")
				continue
			}

			res := fWrite(&rc, nine.Fid(fid), uint64(offset), data)
			fWrite(&cc, terminal, 0, fmt.Sprintf("%v", res))

		case "Tclunk":
			//size[4] Tclunk tag[2] fid[4]

			if len(strArray) != 2 {
				fWrite(&cc, terminal, 0, "incorrect parameters")
				continue
			}

			fid, err := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			}

			if err != nil {
				fWrite(&cc, terminal, 0, "syntax error")
				continue
			}

			res := fClunk(&rc, nine.Fid(fid))
			fWrite(&cc, terminal, 0, fmt.Sprintf("%v", res))

		case "Tremove":

			if len(strArray) != 2 {
				fWrite(&cc, terminal, 0, "incorrect parameters")
				continue
			}

			fid, err := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			}

			if err != nil {
				fWrite(&cc, terminal, 0, "syntax error")
				continue
			}

			res := fRemove(&rc, nine.Fid(fid))
			fWrite(&cc, terminal, 0, fmt.Sprintf("%v", res))

		case "Tstat":

			if len(strArray) != 2 {
				fWrite(&cc, terminal, 0, "incorrect parameters")
				continue
			}

			fid, err := strconv.Atoi(strArray[1])
			if fid > int(^uint32(0)) || fid < 0 {
				fWrite(&cc, terminal, 0, "fid is out of bounds")
				continue
			}

			if err != nil {
				fWrite(&cc, terminal, 0, "syntax error")
				continue
			}

			res := fStat(&rc, nine.Fid(fid))
			fWrite(&cc, terminal, 0, fmt.Sprintf("%v", res))

		case "Twstat":
			fWrite(&cc, terminal, 0, "UNSUPPORTED BY THE CLI ):")

		default:
			fWrite(&cc, terminal, 0, "Command not recognized big sad")

		}
	}
}
