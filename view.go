package PaxiBFT

import (
	"fmt"
	"github.com/salemmohammed/PaxiBFT/log"
	"math"
	"strconv"
	"strings"
)

// Ballot is ballot number type combines 32 bits of natual number and 32 bits of node id into uint64
type View uint64

func NewView(n float64, m float64, id ID) View {
	num := int(math.Mod(n,m))
	//log.Debugf("num=%v",num)
	return View( num<<32 | id.Zone()<<16 | id.Node())
}
// N returns first 32 bit of ballot
func (v View) N() int {
	return int(uint64(v) >> 32)
}

// ID return node id as last 32 bits of ballot
func (v View) ID() ID {
	zone := int(uint32(v) >> 16)
	node := int(uint16(v))
	return NewID(zone, node)
}
// ID return node id as last 32 bits of ballot
func (v View) Reset(id ID) ID {
	return NewIDRest(0, 0)
}

// Next generates the next ballot number given node id
func (v *View) Next(id ID) {
	//log.Debugf("id=%v",id)
	var n float64 = float64(config.N())
	var s string= string(id)
	var b string
	if strings.Contains(s, ".") {
		split := strings.SplitN(s, ".", 2)
		b = split[1]
	} else {
		b = s
	}
	node, err := strconv.ParseUint(b, 10, 64)
	if err != nil {
		log.Errorf("Failed to convert Node %s to int\n", b)
	}
	var m float64 = float64(int(node))
	*v = NewView(m,n,id)
}

func (v View) String() string {
	return fmt.Sprintf("%d",v.N())
}