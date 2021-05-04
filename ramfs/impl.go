package ramfs

import (
	"errors"
	"fmt"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

func (r *RamFs) Register(devNumber uint32) error {
	r.fidTable = make(map[uint64]map[fidDesc]uint64)
	r.devNumber = devNumber
	r.statTable = make(map[uint64]nine.Stat)
	r.dirTable = make(map[uint64][]uint64)
	r.fileTable = make(map[uint64][]byte)
	return nil
}

// "Fresh introduction from a user on the client machine to the server"
func (r *RamFs) Attach(conId uint64, f nine.Fid, uname string) (nine.Qid, error) {
	// "An error is returned if fid is alReady in use"
	if _, ok := r.getFid(conId, f); ok {
		return nine.Qid{}, errors.New("fid alReady in use for session")
	}

	// Note: no authentication is performed, because no afid is passed yet
	// "The client will have a connection to the root directory of the desired
	// file tree, represented by f"

	// Make the root dir if it doesn't exist
	if !r.idExists(0) {
		r.statTable[0] = nine.Stat{
			DevType: nine.DevRamFs,
			DevNo:   r.devNumber,
			Q:       nine.Qid{Flags: nine.FDir},
			Mode:    nine.PUR | nine.PUW | nine.PUX | nine.PGR | nine.PGW | nine.PGX | nine.POR | nine.POW | nine.POX | (nine.FDir << nine.FStatOffset),
			Atime:   mkTimeStamp(),
			Mtime:   mkTimeStamp(),
			Name:    "/",
			Uid:     uname,
			Gid:     uname,
			Muid:    uname,
		}

		// This is a directory, so populate dirTable
		r.dirTable[0] = make([]uint64, 0)
	}

	// Actually attach the client
	fd := fidDesc{
		fid:   f,
		owner: uname,
	}

	// The FidTable for conId might not exist yet!
	if r.fidTable[conId] == nil {
		r.fidTable[conId] = make(map[fidDesc]uint64)
	}
	r.fidTable[conId][fd] = 0

	// The root directory exists no matter what now
	// so we can attach the client and move on
	return r.statTable[0].Q, nil
}

// "The Walk request carries as args an existing fid and a proposed newfid,
// which the client wishes to associate with the result of traversing the
// directory hierarchy using the successive pathname elements wname."
func (r *RamFs) Walk(conId uint64, f nine.Fid, nf nine.Fid, wname []string) ([]nine.Qid, error) {
	// "the fid must be valid in the current session"
	if _, ok := r.getFid(conId, f); !ok {
		return make([]nine.Qid, 0), errors.New("fid not valid in current session")
	}

	// "...newfid may not be in use unless it is the same as fid"
	if _, ok := r.getFid(conId, nf); ok && nf != f {
		return make([]nine.Qid, 0), errors.New("newfid != fid, alReady in use")
	}

	// passed early checks
	fd, _ := r.getFid(conId, f)
	nfd := fidDesc{
		fid:   nf,
		owner: fd.owner,
	}

	id := r.fidTable[conId][fd]

	// "the fid must represent a directory unless zero pathnames are specified"
	if len(wname) != 0 && !r.idIsDir(id) {
		return make([]nine.Qid, 0), errors.New("starting fid doesn't represent dir")
	}

	// "the fid must not have been Opened for I/O by Open or Create"
	if r.isOpenByMe(conId, f) {
		return make([]nine.Qid, 0), errors.New("fid is Open for I/O")
	}
	r.updateATime(id)

	// "it is legal for nwname to be zero, in which case the Walk will succeed
	// and nf will represent the same file as f. the remainder of this discussion
	// assumes this is not the case, i.e. nwname > 0"
	// "the value of nwqid is zero if nwname is zero"
	if len(wname) == 0 {
		r.fidTable[conId][nfd] = id
		return []nine.Qid{}, nil
	}

	// "nwname path name elements wname are Walked in order, elementwise. for the
	// first Walk to succeed, the file identified by f must be a directory and the
	// implied user of the request must have permission to search the directory"
	// ".. (dot-dot) represents the parent" "the current directory [.] is not used"
	wqid := make([]nine.Qid, 0)

	for i, wn := range wname {
		// Check that we're a directory. This never
		// should happen if i == 0
		r.updateATime(id)
		if !r.idIsDir(id) {
			return wqid, nil
		}
		var err error
		id, err = r.descend(id, fd.owner, wn)
		if err != nil {
			if i == 0 {
				// "if the first element cannot be Walked, RError is returned"
				return make([]nine.Qid, 0), err
			} else {
				// "otherwise Walk returns RWalk containing qids corresponding to
				// "successful descends. nwqid is therefore nwname or index of first
				// "descend that failed"
				return wqid, nil
			}
		}
		wqid = append(wqid, r.statTable[id].Q)
	}

	// "If it is equal, nf will be affected, in which case nf represents the file
	// reached by the final descend in the message"
	r.fidTable[conId][nfd] = id
	return wqid, nil
}

// "the Create request asks the file server to Create a new file with the name"
// "supplied, in the directory dir represented by f."
func (r *RamFs) Create(conId uint64, f nine.Fid, name string, perm uint32, mode byte) (nine.Qid, error) {
	// verify that f exists and is a directory currently
	dfd, ok := r.getFid(conId, f)
	if !ok {
		return nine.Qid{}, errors.New("fid not valid in current session")
	}

	// must be a dir
	did := r.fidTable[conId][dfd]
	r.updateATime(did)
	if !r.idIsDir(did) {
		return nine.Qid{}, errors.New("fid is not a directory so can't Create in it")
	}

	// "...requires Write permission in the directory"
	if !r.checkPerms(did, dfd.owner)[1] {
		return nine.Qid{}, errors.New("permission denied to Write to directory")
	}

	// "it is an error if the fid is alReady the product of a successful Open/Create"
	if dfd.open {
		return nine.Qid{}, errors.New("fid is currently Open, Create is illegal")
	}

	// "creating a file in a directory where name alReady exists will be rejected"
	ids := r.dirTable[did]
	for _, nid := range ids {
		n := r.statTable[nid].Name
		if n == name {
			return nine.Qid{}, errors.New("name alReady exists in directory")
		}
	}

	// Create file ID and assign permissions. Also determine
	// if the client wants a directory
	nid := r.genId()
	var isDir bool

	// [ formula specified in the man pages ]
	if perm&(nine.FDir<<nine.FStatOffset) > 0 {
		isDir = true
	} else {
		isDir = false
	}

	// Check that the directory mode is legal
	if isDir {
		openTypeBits := mode & 0b11
		if mode&nine.OTRUNC > 0 {
			return nine.Qid{}, errors.New("cannot trunc a directory")
		} else if mode&nine.ORCLOSE > 0 {
			return nine.Qid{}, errors.New("cannot rclose a directory")
		} else if openTypeBits == nine.OWRITE || openTypeBits == nine.ORDWR {
			return nine.Qid{}, errors.New("it is illegal to Write a directory")
		}
	}

	// Make the Stat
	// "the owner of the file is the user id of the request, group is same as group
	// of dir, and the permissions are as above"
	ns := nine.Stat{
		DevType: nine.DevRamFs,
		DevNo:   r.devNumber,
		Q: nine.Qid{
			Flags: byte(perm >> nine.FStatOffset),
			Id:    nid,
		},
		Mode:  perm,
		Atime: mkTimeStamp(),
		Mtime: mkTimeStamp(),
		Name:  name,
		Uid:   dfd.owner,
		Gid:   r.statTable[did].Gid,
		Muid:  dfd.owner,
	}

	// Commit the Stat to our file tables
	r.statTable[nid] = ns
	if isDir {
		r.dirTable[nid] = make([]uint64, 0)
	} else {
		r.fileTable[nid] = make([]byte, 0)
	}

	// Add the new id to the parent directory
	r.dirTable[did] = append(r.dirTable[did], nid)

	// Prepare to remap dfd
	delete(r.fidTable[conId], dfd)

	// "Finally, the newly Created file is Opened according to mode, which is not
	// checked against permissions in perm. However, it is illegal to open a dir as
	// writeable"
	dfd.open = true
	dfd.openMode = mode

	// Reinstall dfd and done
	r.DumpFs()
	r.fidTable[conId][dfd] = nid
	return r.statTable[nid].Q, nil
}

func (r *RamFs) Open(conId uint64, f nine.Fid, mode byte) (nine.Qid, error) {
	// The fid must exist, obviously
	fd, ok := r.getFid(conId, f)
	if !ok {
		return nine.Qid{}, errors.New("fid requested for Open is not in use")
	}
	id := r.fidTable[conId][fd]
	r.updateATime(id)

	// "it is an error if the fid is alReady the product of a successful Open/Create"
	if fd.open {
		return nine.Qid{}, errors.New("fid is currently Open, Open is illegal")
	}

	// Make initial permission validation check
	fperm := r.checkPerms(id, fd.owner)
	openTypeBits := mode & 0b11

	// If we want to Open as Read and can't Read, that's sad etc.
	if !fperm[0] && (openTypeBits == nine.OREAD || openTypeBits == nine.ORDWR) {
		return nine.Qid{}, errors.New("permission denied, no Read access given")
	} else if !fperm[1] && (openTypeBits == nine.OWRITE || openTypeBits == nine.ORDWR) {
		return nine.Qid{}, errors.New("permission denied, no Write access given")
	} else if !fperm[2] && (openTypeBits == nine.OEXEC) {
		return nine.Qid{}, errors.New("permission denied, no exec access given")
	}

	// Basic permissions check satisfied, now to the nitpicks
	// "It is illegal to Write to a directory, truncate it, or
	// attempt to Remove on close it"
	if r.idIsDir(id) {
		if mode&nine.OTRUNC > 0 {
			return nine.Qid{}, errors.New("cannot trunc a directory")
		} else if mode&nine.ORCLOSE > 0 {
			return nine.Qid{}, errors.New("cannot rclose a directory")
		} else if openTypeBits == nine.OWRITE || openTypeBits == nine.ORDWR {
			return nine.Qid{}, errors.New("it is illegal to Write a directory")
		}
	}

	// "If the mode has the OTRUNC bit set, the file is to be truncated which"
	// "requires Write permission. If the file is append only, nothing happens if
	// "Write permission is given."
	// At this rate we know we're a file, so fine to make this check
	if mode&nine.OTRUNC > 0 {
		if !fperm[1] {
			return nine.Qid{}, errors.New("cannot trunc without Write permission")
		} else if r.statTable[id].Mode&(nine.FAppend<<nine.FStatOffset) == 0 {
			r.updateMTimeAs(fd.owner, id)
			r.fileTable[id] = make([]byte, 0)
		}
	}

	// "If the mode has ORCLOSE set, the file is to be Removed when the fid is Clunked,"
	// "which requires Write access to the parent directory"
	if mode&nine.ORCLOSE > 0 {
		parent, err := r.getParent(id)
		if err != nil {
			return nine.Qid{}, errors.New("failed to get parent to verify ORCLOSE condition")
		} else if !r.checkPerms(parent, fd.owner)[1] {
			return nine.Qid{}, errors.New("cannot ORCLOSE without Write perm to parent dir")
		}
	}

	// "If the file is marked for exclusive use, only one client can have it Open
	// at a time."
	if r.statTable[id].Mode&(nine.FExcl<<nine.FStatOffset) > 0 {
		if r.isOpenByAnyone(id) {
			return nine.Qid{}, errors.New("file is Open by someone")
		}
	}

	// SHOULD be good to go. Open the file and migrate the mapping
	delete(r.fidTable[conId], fd)
	fd.open = true
	fd.openMode = mode
	r.fidTable[conId][fd] = id
	return r.statTable[id].Q, nil
}

func (r *RamFs) Read(conId uint64, f nine.Fid, offset uint64, count uint32) ([]byte, error) {
	// The fid must exist, of course
	fd, ok := r.getFid(conId, f)
	if !ok {
		return make([]byte, 0), errors.New("fid requested for Read is not in use")
	}
	// "The fid must be Opened for Reading"
	if !fd.open || (fd.openMode&0b11 == nine.OWRITE || fd.openMode&0b11 == nine.OEXEC) {
		return make([]byte, 0), errors.New("fid is not Opened for Reading")
	}

	// "For directories, Read returns an integral number of dirents"
	id := r.fidTable[conId][fd]
	r.updateATime(id)

	if r.idIsDir(id) {
		// "Seeking to other than the beginning is illegal in a directory"
		// We are going to be stricter about this. Practically this is limiting,
		// and definitely a TODO item due to limited message sizing, but tracking
		// the last return value is something I don't have time for right now.
		if offset != 0 {
			return make([]byte, 0), errors.New("directory seeking is illegal")
		}

		// Make a buffer to return things into
		scnt := int(count)
		ret := make([]byte, 0)
		for _, id := range r.dirTable[id] {
			StatBytes := nine.MarshalStat(r.statTable[id])
			penalty := len(StatBytes)

			scnt -= penalty
			if scnt < 0 {
				break
			}
			ret = append(ret, StatBytes...)
		}

		return ret, nil
	} else {
		// Is file. If offset is over the end, return zero
		// with no error according to spec
		if offset >= uint64(len(r.fileTable[id])) {
			return make([]byte, 0), nil
		}

		// Else, there's clearly something to Read. Do so one byte at a time
		ret := make([]byte, 0)
		for i := offset; i < uint64(count) && i < uint64(len(r.fileTable[id])); i++ {
			ret = append(ret, r.fileTable[id][i])
		}
		return ret, nil
	}
}

func (r *RamFs) Write(conId uint64, f nine.Fid, offset uint64, data []byte) (uint32, error) {
	// The fid must exist, of course
	fd, ok := r.getFid(conId, f)
	if !ok {
		return 0, errors.New("fid requested for Write is not in use")
	}
	// "The fid must be Opened for writing"
	// It is illegal to Write to a directory, but this is alReady checked for
	if !fd.open || (fd.openMode&0b11 == nine.OREAD || fd.openMode&0b11 == nine.OEXEC) {
		return 0, errors.New("fid is not Opened for writing")
	}

	id := r.fidTable[conId][fd]

	// We don't support the offset being over the end, just set it equal to the end
	// "If the file is append only, the data is placed at the end of the file"
	// "regardless of offset"
	if offset > uint64(len(r.fileTable[id])) || r.statTable[id].Mode&(nine.FAppend<<nine.FStatOffset) != 0 {
		offset = uint64(len(r.fileTable[id]))
	}

	if len(data) == 0 {
		return 0, nil
	}
	r.updateMTimeAs(fd.owner, id)

	// Let's get writing
	// Update the Stat's length view once done
	r.fileTable[id] = append(r.fileTable[id][:offset], append(data, r.fileTable[id][offset:]...)...)

	st := r.statTable[id]
	st.Len = uint64(len(r.fileTable[id]))
	r.statTable[id] = st

	return uint32(len(data)), nil
}

func (r *RamFs) Clunk(conId uint64, f nine.Fid) error {
	// The fid must exist, of course
	fd, ok := r.getFid(conId, f)
	if !ok {
		return errors.New("fid requested for Clunk is not in use")
	}

	// Now, Clunk!
	delete(r.fidTable[conId], fd)
	return nil
}

func (r *RamFs) Remove(conId uint64, f nine.Fid) error {
	// The fid must exist, of course
	fd, ok := r.getFid(conId, f)
	if !ok {
		r.Clunk(conId, f)
		return errors.New("fid requested for Remove is not in use")
	}

	// Get the parent directory and check permissions
	id := r.fidTable[conId][fd]
	pID, err := r.getParent(id)
	if err != nil {
		r.Clunk(conId, f)
		return err
	}

	// "Request will fail if the client does not have Write permission to the parent dir"
	perms := r.checkPerms(pID, fd.owner)
	if !perms[1] {
		r.Clunk(conId, f)
		return errors.New("permission denied to Remove file from directory")
	}

	// Actually delete the file
	r.updateMTimeAs(fd.owner, pID)
	delete(r.statTable, id)
	if r.idIsDir(id) {
		delete(r.dirTable, id)
	} else {
		delete(r.fileTable, id)
	}

	for i, ent := range r.dirTable[pID] {
		if ent == id {
			r.dirTable[pID] = append(r.dirTable[pID][:i], r.dirTable[pID][i+1:]...)
			break
		}
	}

	fmt.Printf("HERERERERERERER")

	r.Clunk(conId, f)
	return nil
}

func (r *RamFs) Stat(conId uint64, f nine.Fid) (nine.Stat, error) {
	// The fid must exist, of course
	fd, ok := r.getFid(conId, f)
	if !ok {
		return nine.Stat{}, errors.New("fid requested for Stat is not in use")
	}

	// "Stat request requires no special permissions"
	id := r.fidTable[conId][fd]
	return r.statTable[id], nil
}

func (r *RamFs) Wstat(conId uint64, f nine.Fid, ns nine.Stat) error {
	// The fid must exist, of course
	fd, ok := r.getFid(conId, f)
	if !ok {
		return errors.New("fid requested for wStat is not in use")
	}

	// s is a *copy* of the old Stat, which we will modify and
	// then commit once all changes go through. Go through the fields, and
	// ensure illegal ones aren't marked...
	id := r.fidTable[conId][fd]
	perms := r.checkPerms(id, fd.owner)
	s := r.statTable[id]

	// ILLEGAL
	if ns.Size != 0 || ns.DevType != 0 || ns.DevNo != 0 {
		return errors.New("cannot modify kernel-use fields of Stat")
	} else if uint(ns.Q.Flags)+uint(ns.Q.Id)+uint(ns.Q.Version) != 0 {
		return errors.New("cannot modify the QID")
	}

	// "Mode can be changed by the owner, or the group leader of the file's group"
	// We have no notion of file ownership in this fs. If the file's owner is you,
	// or the GID is you, you good.
	if (s.Uid == fd.owner || s.Gid == fd.owner) && ns.Mode != 0 {
		// "The directory bit cannot be changed"
		if ns.Mode&(nine.FDir<<nine.FStatOffset) != 0 {
			return errors.New("cannot change the directory bit")
		}

		// It's ok to let the client do whatever they want besides
		// this, for now. Could be bugs for SURE here, but the spec says
		// it's fine???
		s.Mode = ns.Mode
	}

	if ns.Atime != 0 {
		return errors.New("cannot modify atime, but why would you?")
	}

	// Mtime has the same rules as Mode
	// However, I'm enforcing accuracy on the damn timestamp
	mustUpdateMTime := false
	if (s.Uid == fd.owner || s.Gid == fd.owner) && ns.Mtime != 0 {
		mustUpdateMTime = true
	}

	// "Length can be changed by anyone with Write permission, affecting
	// the file's contents. It is an error to set the length of a directory
	// to a non-zero value, and we may choose to reject length sets for other reasons"
	// Go Remove the files from the directory yourselves, idiot, in other words
	mustTrunc := false
	if perms[1] && ns.Len > 0 {
		if r.idIsDir(id) {
			return errors.New("cannot modify length of a directory")
		} else if ns.Len > uint64(len(r.fileTable[id])) {
			return errors.New("cannot increase length of a file, go Write to it")
		}
		s.Len = ns.Len
		mustTrunc = true
	}

	// Changing UID is very illegal, as is MUID, but GID is ok if conditions
	// are satisfied which our strict group definition does not satisfy
	if ns.Uid != "" {
		return errors.New("it is explicitly illegal to change owner of a file")
	} else if ns.Muid != "" {
		return errors.New("cannot change MUID, go wStat the mtime")
	} else if ns.Gid != "" {
		return errors.New("advanced group functionality is not implemented")
	}

	// If we've gotten this far, we can Write the new Stat
	r.statTable[id] = s
	if mustTrunc {
		r.fileTable[id] = r.fileTable[id][:s.Len]
		mustUpdateMTime = true
	}
	if mustUpdateMTime {
		r.updateMTimeAs(fd.owner, id)
	}
	return nil
}

func (r *RamFs) Goodbye(conId uint64) error {
	delete(r.fidTable, conId)
	return nil
}
