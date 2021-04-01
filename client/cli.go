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
// server. It assumes a port of 5640, and
// that files are being served from localhost.
func RunClient(c *nine.Conf) {
	var strArray []string
	address := fmt.Sprintf("localhost:%d", c.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Failed to dial local server:", err)
	}
	strArray = strings.Split(address, " ") //holy grail

	//size[4] Tversion tag[2] msize[4] version[s]
	//size[4] Rversion tag[2] msize[4] version[s]

	if strArray[0] == "version" || strArray[0] == "Version" || strArray[0] == "TVersion" || strArray[0] == "RVersion" {

		tag, err := strconv.Atoi(strArray[1]) /*Other alternative is to use ParseInt,
		  it's a bit faster but more work confusing imo*/
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		msize, err := strconv.Atoi(strArray[2])
		if msize > 4294967295 || msize < 0 {
			fmt.Printf("msize is out of bounds")
		}

		s, err := strconv.Atoi(strArray[3])
		if s > 65535 || s < 0 {
			fmt.Printf("s is out of bounds")
		}

		version := (strArray[4])
		bytes := []byte(version)
		if bytes[0] > (byte)((2^(s*8))-1) || bytes[0] < 0 { //This is most probably worng, not sure how to check for string ---> bytes
			fmt.Printf("Version is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	//size[4] Tauth tag[2] afid[4] uname[s] aname[s]

	if strArray[0] == "Tauth" || strArray[0] == "auth" || strArray[0] == "Auth" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		afid, err := strconv.Atoi(strArray[2])
		if afid > 4294967295 || afid < 0 {
			fmt.Printf("afid is out of bounds")
		}

		s, err := strconv.Atoi(strArray[3])
		if s > 65535 || s < 0 {
			fmt.Printf("s is out of bounds")
		}

		uname := (strArray[4])
		bytes := []byte(uname)
		if bytes[0] > (byte)((2^(s*8))-1) || bytes[0] < 0 { //This is most probably worng, not sure how to check for string ---> bytes
			fmt.Printf("uname is out of bounds")
		}

		s1, err := strconv.Atoi(strArray[5])
		if s1 > 65535 || s1 < 0 {
			fmt.Printf("s1 is out of bounds")
		}

		aname := (strArray[6])
		bytes1 := []byte(aname)
		if bytes1[0] > (byte)((2^(s1*8))-1) || bytes1[0] < 0 { //FIXME:This is most probably worng, not sure how to check for string ---> bytes
			fmt.Printf("aname is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//size[4] Rauth tag[2] aqid[13]
	else if strArray[0] == "Rauth" { // FIXME: Not sure if "auth" and "Auth" count as Tauth or Rauth

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		aqid, err := strconv.Atoi(strArray[3])
		if aqid > (2^(13*8))-1 || aqid < 0 { //FIXME:Big OOF, needs to be resolved
			fmt.Printf("aqid is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}

	}

	//size[4] Rerror tag[2] ename[s]
	else if strArray[0] == "Rerror" || strArray[0] == "error" || strArray[0] == "Error" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		s, err := strconv.Atoi(strArray[2])
		if s > 65535 || s < 0 {
			fmt.Printf("s is out of bounds")
		}

		ename := strArray[3]
		bytes := []byte(ename)
		if bytes[0] > (byte)((2^(s*8))-1) || bytes[0] < 0 { //FIXMEThis is most probably worng, not sure how to check for string ---> bytes
			fmt.Printf("ename is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//size[4] Tflush tag[2] oldtag[2]
	else if strArray[0] == "Tflush" || strArray[0] == "flush" || strArray[0] == "Flush" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		oldtag, err := strconv.Atoi(strArray[1])
		if oldtag > 65535 || oldtag < 0 {
			fmt.Printf("oldtag is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//size[4] Rflush tag[2]
	else if strArray[0] == "Rflush" { //TODO: Just like auth, not sure if "flush" and "FLush" show be here or in Tflush

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}
		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//size[4] Tattach tag[2] fid[4] afid[4] uname[s] aname[s]
	else if strArray[0] == "Tattach" || strArray[0] == "Attach" || strArray[0] == "attach" {
		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		fid, err := strconv.Atoi(strArray[2])
		if fid > 4294967295 || fid < 0 {
			fmt.Printf("fid is out of bounds")
		}

		afid, err := strconv.Atoi(strArray[3])
		if afid > 4294967295 || afid < 0 {
			fmt.Printf("afid is out of bounds")
		}

		s, err := strconv.Atoi(strArray[4])
		if s > 65535 || s < 0 {
			fmt.Printf("s is out of bounds")
		}

		uname := (strArray[5])
		bytes := []byte(uname)
		if bytes[0] > (byte)((2^(s*8))-1) || bytes[0] < 0 { //This is most probably worng, not sure how to check for string ---> bytes
			fmt.Printf("uname is out of bounds")
		}

		s1, err := strconv.Atoi(strArray[6])
		if s1 > 65535 || s1 < 0 {
			fmt.Printf("s1 is out of bounds")
		}

		aname := (strArray[7])
		bytes1 := []byte(aname)
		if bytes1[0] > (byte)((2^(s1*8))-1) || bytes1[0] < 0 { //FIXME:This is most probably worng, not sure how to check for string ---> bytes
			fmt.Printf("aname is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//size[4] Rattach tag[2] qid[13]
	else if strArray[0] == "Rattach" {
		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		qid, err := strconv.Atoi(strArray[2])
		if qid > (2^(13*8))-1 || qid < 0 { //FIXME:Big OOF, needs to be resolved
			fmt.Printf("qid is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//TODO: size[4] Twalk tag[2] fid[4] newfid[4] nwname[2] nwname*(wname[s])

	//TODO: size[4] Rwalk tag[2] nwqid[2] nwqid*(wqid[13])

	//size[4] Topen tag[2] fid[4] mode[1]
	else if strArray[0] == "Topen" || strArray[0] == "open" || strArray[0] == "Open" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		fid, err := strconv.Atoi(strArray[2])
		if fid > 4294967295 || fid < 0 {
			fmt.Printf("fid is out of bounds")
		}

		mode, err := strconv.Atoi(strArray[3])
		if mode > 255 || mode < 0 {
			fmt.Printf("mode is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//size[4] Ropen tag[2] qid[13] iounit[4]
	else if strArray[0] == "Ropen" {
		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		qid, err := strconv.Atoi(strArray[2])
		if qid > (2^(13*8))-1 || qid < 0 {
			fmt.Printf("qid is out of bounds")
		}

		iounit, err := strconv.Atoi(strArray[3])
		if iounit > 4294967295 || iounit < 0 {
			fmt.Printf("iounit is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}

	}

	//size[4] Tcreate tag[2] fid[4] name[s] perm[4] mode[1]
	else if strArray[0] == "Tcreate" || strArray[0] == "create" || strArray[0] == "Create" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		fid, err := strconv.Atoi(strArray[2])
		if fid > 4294967295 || fid < 0 {
			fmt.Printf("fid is out of bounds")
		}

		s, err := strconv.Atoi(strArray[3])
		if s > 65535 || s < 0 {
			fmt.Printf("s is out of bounds")
		}

		name := (strArray[4])
		bytes := []byte(name)
		if bytes[0] > (byte)((2^(s*8))-1) || bytes[0] < 0 { //This is most probably worng, not sure how to check for string ---> bytes
			fmt.Printf("name is out of bounds")
		}

		perm, err := strconv.Atoi(strArray[5])
		if perm > 4294967295 || perm < 0 {
			fmt.Printf("perm is out of bounds")
		}

		mode, err := strconv.Atoi(strArray[6])
		if mode > 255 || mode < 0 {
			fmt.Printf("mode is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//size[4] Rcreate tag[2] qid[13] iounit[4]
	else if strArray[0] == "Rcreate" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		qid, err := strconv.Atoi(strArray[2])
		if qid > (2^(13*8))-1 || qid < 0 {
			fmt.Printf("qid is out of bounds")
		}

		iounit, err := strconv.Atoi(strArray[3])
		if iounit > 4294967295 || iounit < 0 {
			fmt.Printf("iounit is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//size[4] Tread tag[2] fid[4] offset[8] count[4]

	else if strArray[0] == "Tread" || strArray[0] == "read" || strArray[0] == "Read" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		fid, err := strconv.Atoi(strArray[2])
		if fid > 4294967295 || fid < 0 {
			fmt.Printf("fid is out of bounds")
		}

		offset, err := strconv.Atoi(strArray[3])
		if offset > (2^(8*8))-1 || offset < 0 {
			fmt.Printf("offset is out of bounds")
		}

		count, err := strconv.Atoi(strArray[4])
		if count > 4294967295 || count < 0 {
			fmt.Printf("fid is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//size[4] Rread tag[2] count[4] data[count]
	else if strArray[0] == "Rread" {

		tag, err := strconv.Atoi(strArray[1]) /*Other alternative is to use ParseInt,
		  it's a bit faster but more work confusing imo*/
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		count, err := strconv.Atoi(strArray[2])
		if count > 4294967295 || count < 0 {
			fmt.Printf("fid is out of bounds")
		}

		data, err := strconv.Atoi(strArray[3])
		if data > (2^(count*8))-1 || data < 0 {
			fmt.Printf("data is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //FIXME:Not sure what the error message should be
		}
	}

	//size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]
	else if strArray[0] == "write" || strArray[0] == "Write" || strArray[0] == "TWrite" {
		
		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		fid, err := strconv.Atoi(strArray[2])
		if fid > 4294967295 || fid < 0 {
			fmt.Printf("fid is out of bounds")
		}

		offset, err := strconv.Atoi(strArray[3])
		if offset > (2^(8*8))-1 || offset < 0 { //FIXME:Big OOF, needs to be resolved
			fmt.Printf("offset is out of bounds")
		}

		count, err := strconv.Atoi(strArray[4])
		if count > 4294967295 || count < 0 {
			fmt.Printf("fid is out of bounds")
		}

		data, err := strconv.Atoi(strArray[5])
		if data > (2^(count*8))-1 || data < 0 {
			fmt.Printf("data is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	//size[4] Rwrite tag[2] count[4]
	else if strArray[0] == "Rwrite" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		count, err := strconv.Atoi(strArray[2])
		if count > 4294967295 || count < 0 {
			fmt.Printf("fid is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	//size[4] Tclunk tag[2] fid[4]
	else if strArray[0] == "Tclunk" || strArray[0] == "Clunk" || strArray[0] == "clunk" {
		
		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		fid, err := strconv.Atoi(strArray[2])
		if fid > 4294967295 || fid < 0 {
			fmt.Printf("fid is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	//size[4] Rclunk tag[2]
	else if strArray[0] == "Rclunk" {
		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}
		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	//size[4] Tremove tag[2] fid[4]
	else if strArray[0] == "Tremove" || strArray[0] == "Remove" || strArray[0] == "remove" {
		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		fid, err := strconv.Atoi(strArray[2])
		if fid > 4294967295 || fid < 0 {
			fmt.Printf("fid is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	//size[4] Rremove tag[2]
	else if strArray[0] == "Rremove" {
		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}
		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	//size[4] Tstat tag[2] fid[4]
	else if strArray[0] == "Tstat" || strArray[0] == "Stat" || strArray[0] == "stat" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		fid, err := strconv.Atoi(strArray[2])
		if fid > 4294967295 || fid < 0 {
			fmt.Printf("fid is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	//size[4] Rstat tag[2] stat[n]
	else if strArray[0] == "Rstat" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		n, err := strconv.Atoi(strArray[2]) //FIXME:Same as the string cases, not sure wht the doc is saying
		if n > 65535 || n < 0 {
			fmt.Printf("s is out of bounds")
		}

		stat, err := strconv.Atoi(strArray[3])
		if stat > 2^(n*8)-1 || stat < 0 {
			fmt.Printf("stat is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	//size[4] Twstat tag[2] fid[4] stat[n]
	else if strArray[0] == "Twstat" || strArray[0] == "Wstat" || strArray[0] == "wstat" {

		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		fid, err := strconv.Atoi(strArray[2])
		if fid > 4294967295 || fid < 0 {
			fmt.Printf("fid is out of bounds")
		}

		n, err := strconv.Atoi(strArray[3]) //FIXME:Same as the string cases, not sure wht the doc is saying
		if n > 65535 || n < 0 {
			fmt.Printf("s is out of bounds")
		}

		stat, err := strconv.Atoi(strArray[4])
		if stat > 2^(n*8)-1 || stat < 0 {
			fmt.Printf("stat is out of bounds")
		}
		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	//size[4] Rwstat tag[2]
	else if strArray[0] == "Rwstat" {
		tag, err := strconv.Atoi(strArray[1])
		if tag > 65535 || tag < 0 {
			fmt.Printf("tag is out of bounds")
		}

		if err != nil {
			log.Fatal("Failed to dial local server:", err) //Not sure what the error message should be
		}
	}

	else {
		fmt.Printf("Command not recognized :(")
	}

	tf := nine.FCall{
		MsgType: nine.TVersion,
		Tag:     5,
		Version: (*c).Version,
	}
	rf := writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RVersion)

	// Make me a new fid!
	var rootFid nine.Fid = 10
	tf = nine.FCall{
		MsgType: nine.TAttach,
		Tag:     6,
		F:       rootFid,
		Uname:   "jaytlang",
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RAttach)

	// Dup it
	var anotherRootFid nine.Fid = 11
	tf = nine.FCall{
		MsgType: nine.TWalk,
		Tag:     9,
		F:       rootFid,
		Newf:    anotherRootFid,
		Wname:   make([]string, 0),
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RWalk)

	// Make me another one
	tf = nine.FCall{
		MsgType: nine.TCreate,
		Tag:     7,
		F:       anotherRootFid,
		Name:    "snoop",
		Perm:    (nine.FDir << nine.FStatOffset) | nine.PUR | nine.PUW | nine.PUX | nine.PGR | nine.PGX | nine.POR | nine.POX,
		Mode:    nine.OREAD,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RCreate)

	// Clunk this one
	tf = nine.FCall{
		MsgType: nine.TClunk,
		Tag:     14,
		F:       anotherRootFid,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RClunk)

	// Walk to the new directory
	var dirFid nine.Fid = 420
	tf = nine.FCall{
		MsgType: nine.TWalk,
		Tag:     15,
		F:       rootFid,
		Newf:    dirFid,
		Wname:   []string{"snoop"},
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RWalk)

	// Stat it
	tf = nine.FCall{
		MsgType: nine.TStat,
		Tag:     20,
		F:       dirFid,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RStat)

	// Make a file in it
	tf = nine.FCall{
		MsgType: nine.TCreate,
		Tag:     70,
		F:       dirFid,
		Name:    "snoop.txt",
		Perm:    (nine.FAppend << nine.FStatOffset) | nine.PUR | nine.PUW | nine.PGR | nine.PGW,
		Mode:    nine.ORDWR,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RCreate)

	// Write to it
	tf = nine.FCall{
		MsgType: nine.TWrite,
		Tag:     25,
		F:       dirFid,
		Offset:  5,
		Data:    []byte("SMOKE WEED ERY DAY"),
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RWrite)

	// Stat the new file that has a changed length
	tf = nine.FCall{
		MsgType: nine.TStat,
		Tag:     21,
		F:       dirFid,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RStat)

	// Read 30 bytes from our file
	tf = nine.FCall{
		MsgType: nine.TRead,
		Tag:     230,
		F:       dirFid,
		Offset:  0,
		Count:   30,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RRead)

	conn.Close()
}

/*
	size[4] Tversion tag[2] msize[4] version[s]
	size[4] Rversion tag[2] msize[4] version[s]

	size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]

	size[4] Twalk tag[2] fid[4] newfid[4] nwname[2] nwname*(wname[s])
	size[4] Rwalk tag[2] nwqid[2] nwqid*(qid[13])

	size[4] Twstat tag[2] fid[4] stat[n]

	32 bit number that describes file permissions (called Mode in stat)
	fDir|fAppend|fExcl|fAuth|ftmp|unused|unused...|unused|PUR|PUW|PUX|PGR|PGW|PGX|POR|POW|POX

	write 32434 3 0 2 hi
*/

func writeAndRead(c *net.Conn, f *nine.FCall) *nine.FCall {
	if err := nine.Write9P(c, f); err != nil {
		log.Fatal(err)
	}
	rf, err := nine.Read9P(c)
	if err != nil {
		log.Fatal(err)
	}
	return &rf
}

func checkMsg(f *nine.FCall, expected byte) {
	if f.MsgType == expected {
		fmt.Printf("Success! ")
	} else {
		fmt.Printf("FAILURE! ")
	}

	fmt.Printf("Got message type %d\n", f.MsgType)
	fmt.Printf("Full struct: %v\n", f)
}
