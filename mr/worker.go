package mr

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"net/rpc"
	"strings"
	"time"

	"github.com/coolprototype-9000/cp9-proto/client"
	"github.com/coolprototype-9000/cp9-proto/nine"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string, tp *client.Proc) {

	setupCon()
	p = tp

	// Forever...
	for {
		a := AskArgs{}
		r := AskReply{}

		// Get a task
		if !call("Coordinator.Ask", &a, &r) {
			log.Fatal("failed to Ask")
		}

		// Start working on whatever we have to do
		// If we're done, return straightaway
		if r.Task == Done {
			break
		} else if r.Task == Map {
			fmt.Println("got map request")
			// Pop open the file assigned
			content := ReadFrom(r.Filename)

			// Call map
			kva := mapf(r.Filename, content)
			fmt.Println("map finished")

			// Done. Create r.ReduceTasks files
			flist := make([]int, r.ReduceTasks)

			for i := 0; i < r.ReduceTasks; i++ {
				f := p.Create(MkIntermediateName(r.TaskId, i), nine.ORDWR, nine.PUR|nine.PUW|nine.PGR|nine.PGW|nine.POR|nine.POW)
				if f < 0 {
					log.Fatal("failed to create intermediate file")
				}
				flist[i] = f
			}

			fmt.Println("created intermediate files")

			// Iterate through all keys in kva, using the hash
			// to pick files to write to
			fcnt := make(map[int]string)
			for _, kv := range kva {
				target := ihash(kv.Key) % r.ReduceTasks
				fcnt[flist[target]] += fmt.Sprintf("%s %s\n", kv.Key, kv.Value)
			}

			fmt.Println("populated files")

			for fd, s := range fcnt {
				p.Write(fd, s)
			}
			fmt.Println("flushed files")

			// Close every file
			for i := 0; i < r.ReduceTasks; i++ {
				p.Close(flist[i])
			}

			// Tell coordinator we are done
			ta := TellArgs{
				Task:     Map,
				TaskId:   r.TaskId,
				Filename: r.Filename,
			}

			tr := TellReply{}
			if !call("Coordinator.Tell", &ta, &tr) {
				log.Fatal("failed to Tell")
			}
			fmt.Println("done")

		} else {
			fmt.Println("got reduce request")

			// Reduce: get the list of files for our id
			flist := IntermediatesFor(r.TaskId)

			// Our objective is to consolidate everything in our
			// intermediates into a map like key, []value
			masterMap := make(map[string][]string)

			// for each file, parse line by line and accumulate our
			// values inside a new map
			for _, fname := range flist {
				f := p.Open(fname, nine.ORDWR)
				if f < 0 {
					log.Fatal("failed to open intermediate file (in worker)")
				}

				fContent := p.Read(f, ^uint32(0))
				scanner := bufio.NewScanner(strings.NewReader(fContent))
				for scanner.Scan() {
					kv := ParseIntermediateEnt(scanner.Text())
					if _, ok := masterMap[kv.Key]; ok {
						masterMap[kv.Key] = append(masterMap[kv.Key], kv.Value)
					} else {
						masterMap[kv.Key] = make([]string, 1)
						masterMap[kv.Key][0] = kv.Value
					}
				}
				p.Close(f)
				fmt.Println("aggregated a file into master map")
			}

			// make a temporary file
			rand.Seed(time.Now().Unix())
			tfName := fmt.Sprintf("mr-tmp-out%d", rand.Uint64())
			tf := p.Create(tfName, nine.ORDWR, nine.PUR|nine.PUW|nine.PGR|nine.PGW|nine.POR|nine.POX)
			if tf < 0 {
				log.Fatal("failed to open tempfile for reduce")
			}

			fmt.Println("created tempfile")

			// ship off each key/[]value to reduce
			var s string
			for k, v := range masterMap {
				output := reducef(k, v)
				s += fmt.Sprintf("%v %v\n", k, output)
			}
			fmt.Println("reduce finished")

			p.Write(tf, s)
			fmt.Println("reduce flushed")

			// close the tempfile and rename it atomically
			p.Close(tf)
			err := p.Rename(tfName, r.Filename)
			if err < 0 {
				log.Fatalf("failed to perform final rename: %s", p.Errstr())
			}
			fmt.Printf("made final rename\n")

			// Tell coordinator we are done
			ta := TellArgs{
				Task:     Reduce,
				TaskId:   r.TaskId,
				Filename: r.Filename,
			}

			tr := TellReply{}
			if !call("Coordinator.Tell", &ta, &tr) {
				log.Fatal("failed to Tell")
			}

			fmt.Println("done")
		}
	}

}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//

var c *rpc.Client

func setupCon() {
retryDial:
	nc, err := rpc.DialHTTP("tcp", "23.28.10.45"+":5630")
	if err != nil {
		fmt.Printf("Failed to dial: %s. Retrying in 5\n", err.Error())
		time.Sleep(5 * time.Second)
		goto retryDial
	}
	c = nc

}
func call(rpcname string, args interface{}, reply interface{}) bool {
	err := c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
