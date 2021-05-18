package streamletBFT

import (
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"strconv"
	"sync"
	"time"
)

type Replica struct {
	PaxiBFT.Node
	*StreamletBFT
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
	r.StreamletBFT = NewStreamletBFT(r)
	r.Register(PaxiBFT.Request{}, 	  r.handleRequest)
	r.Register(Propose{},         	  r.handlePropose)
	r.Register(Vote{},  		  	  r.HandleVote)
	r.Register(ViewChange{},  		  r.HandleViewChange)
	r.Register(ProposeAfterFailure{}, r.HandleProposeAfterFailure)

	return r
}
func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("\n<-----------handleRequest----------->\n")
	if p.slot <= 0 {
		fmt.Print("-------------------streamletBFT-------------------------")
	}
	p.slot++
	if p.slot % 1000 == 0 {
		fmt.Print("p.slot", p.slot)
	}
	p.Requests = append(p.Requests, &m)
	e, ok := p.log[p.slot]
	if !ok{
		p.log[p.slot] = &entry{
			Ballot:    	p.ballot,
			request:   	&m,
			Timestamp: 	time.Now(),
			Q1:			PaxiBFT.NewQuorum(),
			Q2: 		PaxiBFT.NewQuorum(),
			Q3: 		PaxiBFT.NewQuorum(),
			active:     false,
			Leader:		false,
			commit:    	false,
		}
	}
	e = p.log[p.slot]
	log.Debugf("e.request= %v" , e.request)
	e.request = &m
	w := p.slot % e.Q1.Total() + 1
	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(w))
	log.Debugf("Node_ID = %v", Node_ID)

	if Node_ID == p.ID(){
		log.Debugf("leader")
		time.Sleep(500 * time.Millisecond)
		e.active = true
	}

	if e.active == true {
		e.Leader = true
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