package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"

	"github.com/coolprototype-9000/cp9-proto/client"
)

//
// Represent the status of files
// from the coordinator's view
//
type fStatus int

const (
	unassigned fStatus = iota
	assigned   fStatus = iota
	complete   fStatus = iota
)

type Coordinator struct {
	nReduce int
	nMap    int
	mFiles  map[string]fStatus
	rFiles  map[string]fStatus
	state   TaskType
	lock    sync.Mutex
}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished. very much not there yet.
//
// note that we aren't calling this even though go thinks we
// might, hence the weird error that pops up when you start this thing
//
func (c *Coordinator) Done() bool {
	c.lock.Lock()
	ret := (c.state == Done)
	c.lock.Unlock()
	return ret
}

func (c *Coordinator) Tell(a *TellArgs, r *TellReply) error {
	c.lock.Lock()

	if c.state == Map {
		// If our functionality has been superseded...
		if c.mFiles[a.Filename] != assigned {

			// Trash every intermediate file (ignoring errors)
			for i := 0; i < c.nReduce; i++ {
				fname := MkIntermediateName(a.TaskId, i)
				p.Remove(fname)
			}

			c.lock.Unlock()
			return nil
		}

		// We should be good. Mark this file as complete
		// and shove everything into rFiles
		c.mFiles[a.Filename] = complete
		for i := 0; i < c.nReduce; i++ {
			fname := MkIntermediateName(a.TaskId, i)
			c.rFiles[fname] = unassigned
		}

		// If all files in mFiles are complete, switch state
		// to reducing from mapping
		shouldswitch := true
		for _, v := range c.mFiles {
			if v != complete {
				shouldswitch = false
				break
			}
		}

		if shouldswitch {
			fmt.Println("State is now REDUCE")
			c.state = Reduce
		}

	} else if c.state == Reduce {
		// First, is this really reduce
		if a.Task != Reduce {
			// Nope. Basically do nothing in this case.
			// We can leave its files be assuming its a
			// map since it didn't play into the reduce
			c.lock.Unlock()
			return nil
		}

		// No need to check if we've been beaten by someone else,
		// unlike with map. This is because the only thing that happens
		// with duplicate effort is a file just is renamed-on. Reduce has
		// deterministic output so should be ok. Move on and mark as done.
		flist := IntermediatesFor(a.TaskId)
		for _, f := range flist {
			c.rFiles[f] = complete
		}

		// Iterate one last time to ensure we're done?
		shouldswitch := true
		for _, v := range c.rFiles {
			if v != complete {
				shouldswitch = false
				break
			}
		}

		if shouldswitch {
			fmt.Println("We are done!")
			c.state = Done
		}
	}

	c.lock.Unlock()
	return nil
}

func (c *Coordinator) Ask(a *AskArgs, r *AskReply) error {
	// Serialize everything that comes through the coordinator
	// This is crap for performance but this whole function is
	// basically a critical section so
	c.lock.Lock()

	r.ReduceTasks = c.nReduce
	done := false

	// Jump over this for loop if we're
	// finished, no assignments left to
	// do...
	for !done {
		if c.state == Done {
			r.Task = Done
			done = true
		} else if c.state == Map {
			// Map: iterate through the files map
			// and figure out who to assign to who
			for k, v := range c.mFiles {
				if v == unassigned {
					// Assign
					c.mFiles[k] = assigned
					r.Task = Map
					r.TaskId = c.nMap
					r.Filename = k

					go c.watchdog([]string{k}, Map)
					c.nMap++
					done = true
					break
				}
			}

		} else {
			// Reduce is able to infer what filenames
			// it should be reducing over. So let's just
			// tell reduce what task it is and it'll figure
			// the rest, passing its output file instead of
			// any convoluted input. First iterate through
			// rFiles...
			for k, v := range c.rFiles {
				if v == unassigned {
					// Parse out the reduceId from this filename
					// and get all of its peers
					id := IdForIntermediate(k)
					flist := IntermediatesFor(id)

					// Assign everyone in flist to a new reducetask
					// with ID id
					for _, fname := range flist {
						c.rFiles[fname] = assigned
					}
					r.Task = Reduce
					r.TaskId = id
					r.Filename = fmt.Sprintf("mr-out-%d", id)

					go c.watchdog(flist, Reduce)

					done = true
					break
				}
			}
		}

		// Have we been assigned?
		// If not, sleep for a sec after
		// releasing the lock so others can
		// come and be assigned
		if !done {
			c.lock.Unlock()
			time.Sleep(1 * time.Second)
			c.lock.Lock()
		}
	}

	c.lock.Unlock()
	return nil
}

func (c *Coordinator) watchdog(fnames []string, t TaskType) {
	// Sit here, chillin, waiting for a while
	// This is called in a new goroutine...
	time.Sleep(10 * time.Second)
	c.lock.Lock()

	// Now that we back, sample tracker[fnames[0]]
	// for completeness. If it's not done, unassign
	// all filenames we've been told to watch
	if t == Map {
		if c.mFiles[fnames[0]] != complete {
			fmt.Println("Worker did a stupid!")
			c.mFiles[fnames[0]] = unassigned
		}
	} else {
		if c.rFiles[fnames[0]] != complete {
			fmt.Println("Worker did a stupid!")
			for _, fname := range fnames {
				c.rFiles[fname] = unassigned
			}
		}
	}

	c.lock.Unlock()

}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//

func MakeCoordinator(files []string, nReduce int, tp *client.Proc) *Coordinator {
	// The basics
	// Note: lock is zero-initialized automatically

	p = tp

	c := Coordinator{
		nReduce: nReduce,
		state:   Map,
	}

	// Set up the file map
	c.mFiles = make(map[string]fStatus)
	for _, f := range files {
		c.mFiles[f] = unassigned
	}

	c.rFiles = make(map[string]fStatus)

	// Spin off a server in a background thread and
	// give a reference to the caller
	c.server()
	return &c
}
