package HotStuff

import (
	"fmt"
	"github.com/ailidani/paxi"
	"github.com/ailidani/paxi/log"
	//"math"
	//"strconv"
	//"strings"
	"sync"
	)

// Replica for one Tendermint instance
type Replica struct {
	paxi.Node
	*HotStuff
	mux sync.Mutex
}
const (
	HTTPHeaderSlot       = "Slot"
	HTTPHeaderBallot     = "Ballot"
	HTTPHeaderExecute    = "Execute"
	HTTPHeaderInProgress = "Inprogress"
)
// NewReplica generates new Pbft replica
func NewReplica(id paxi.ID) *Replica {
	log.Debugf("Replica started \n")
	r := new(Replica)

	r.Node = paxi.NewNode(id)
	r.HotStuff = NewHotStuff(r)

	r.Register(paxi.Request{},  r.handleRequest)

	r.Register(Prepare{},       r.handlePrepare)
	r.Register(ActPrepare{},    r.handleActPrepare)

	r.Register(PreCommit{},     r.handlePreCommit)
	r.Register(ActPreCommit{},  r.handleActPreCommit)

	r.Register(Commit{},        r.handleCommit)
	r.Register(ActCommit{},     r.handleActCommit)

	r.Register(Decide{},        r.handleDecide)
	r.Register(ActDecide{},     r.handleActDecide)

	return r
}
func (p *Replica) handleRequest(m paxi.Request) {
	log.Debugf("\n<-----------Leader of the Request----------->\n")
	if p.slot <= 0 {
		fmt.Print("-------------------HotStuff-------------------------")
	}
	fmt.Println("Request received %v", m.Command.Count)
	p.slot = m.Command.Count

	p.Requests = append(p.Requests, &m)

	_, ok := p.logR[p.slot]
	if !ok {
		p.logR[p.slot] = &RequestSlot{
			request:  &m,
			slot:     p.slot,
			RecReqst: paxi.NewQuorum(),
			commit:   false,
			count:    0,
			active:   false,
			Leader:   false,
		}
	}
	e, ok := p.logR[p.slot]


	e.Leader = true
	p.ballot.Next(p.ID())
	log.Debugf("p.ballot %v ", p.ballot)
	log.Debugf("<<<<<<<<<<<<<<<<<< The leader for >>>>>>>>>>>>>>>>>>>>> %v",p.slot)
	p.view.Next(p.ID())
	log.Debugf("Sender ID %v, slot=%v", m.NodeID, p.slot)

	p.HotStuff.HandleRequest(m,p.slot)
}