package HotStuff

import (
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"strconv"
	"sync"
	"time"
)

// Replica for one Tendermint instance
type Replica struct {
	PaxiBFT.Node
	*HotStuff
	mux sync.Mutex
}
const (
	HTTPHeaderSlot       = "Slot"
	HTTPHeaderBallot     = "Ballot"
	HTTPHeaderExecute    = "Execute"
	HTTPHeaderInProgress = "Inprogress"
)
func NewReplica(id PaxiBFT.ID) *Replica {
	log.Debugf("Replica started \n")
	r := new(Replica)
	r.Node = PaxiBFT.NewNode(id)
	r.HotStuff = NewHotStuff(r)
	r.Register(PaxiBFT.Request{},  r.handleRequest)

	r.Register(Prepare{},          r.handlePrepare)
	r.Register(ActPrepare{},       r.handleActPrepare)

	r.Register(PreCommit{},        r.handlePreCommit)
	r.Register(ActPreCommit{},     r.handleActPreCommit)

	r.Register(Commit{},           r.handleCommit)
	r.Register(ActCommit{},        r.handleActCommit)

	r.Register(Decide{},        r.handleDecide)
	r.Register(ActDecide{},     r.handleActDecide)

	return r
}
func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("\n<-----------Leader of the Request----------->\n")
	if p.slot <= 0 {
		fmt.Print("-------------------HotStuff-------------------------")
	}

	p.Requests = append(p.Requests, &m)

	e, ok := p.log[p.slot]
	if !ok {
		p.log[p.slot] = &entry{
			Ballot:    	p.ballot,
			commit:    	false,
			request:   	&m,
			Timestamp: 	time.Now(),
			Q1:			PaxiBFT.NewQuorum(),
			Q2: 		PaxiBFT.NewQuorum(),
			Q3: 		PaxiBFT.NewQuorum(),
			Q4: 		PaxiBFT.NewQuorum(),
			active:     false,
			leader:     false,
			Digest:     GetMD5Hash(&m),
		}
	}
	e = p.log[p.slot]
	log.Debugf("e.request= %v" , e.request)
	e.request = &m
    e.Digest  = GetMD5Hash(&m)
	w := p.slot % e.Q1.Total() + 1
	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(w))
	log.Debugf("Node_ID = %v", Node_ID)

	if Node_ID == p.ID(){
		log.Debugf("leader")
		e.active = true
	}

	if e.active == true {
		e.leader = true
		p.ballot.Next(p.ID())
		log.Debugf("p.ballot %v ", p.ballot)
		e.Ballot = p.ballot
		e.Pstatus = PREPARED
		p.HandleRequest(m)
	}

	e.Rstatus = RECEIVED
	log.Debugf("e.Pstatus = %v", e.Pstatus)
	if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED{
		log.Debug("late call")
		e.commit = true
		p.exec()
	}
}