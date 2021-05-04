package mr

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/coolprototype-9000/cp9-proto/client"
)

// Helper functions
// All of these will log.Fatal if
// they encounter an error, which
// simplifies calling code at the expense
// of reliability. Generally anything in
// here shouldn't fail unless something is
// seriously wrong.

var p *client.Proc

func ReadFrom(fname string) string {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalf("failed to open %v for reading", fname)
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("failed to read %v", fname)
	}
	f.Close()

	return string(content)
}

func MkIntermediateName(mapId int, reduceId int) string {
	return fmt.Sprintf("mr-%d-%d.tmp", mapId, reduceId)
}

func IntermediatesFor(reduceId int) []string {
	// Check the working directory for all files which match
	// the suffix *[reduceId].tmp
	glob := fmt.Sprintf("*%d.tmp", reduceId)
	matches, err := filepath.Glob(glob)
	if err != nil {
		log.Fatalf("failed to glob for %s", glob)
	}

	return matches
}

func IdForIntermediate(intermd string) int {
	si := strings.LastIndex(intermd, "-")
	ei := strings.LastIndex(intermd, ".")

	id, err := strconv.Atoi(intermd[si+1 : ei])
	if err != nil {
		log.Fatalf("probable invalid filename %s", intermd)
	}

	return id
}

func ParseIntermediateEnt(ent string) KeyValue {
	kvstr := strings.Fields(ent)
	kv := KeyValue{kvstr[0], kvstr[1]}
	return kv
}
