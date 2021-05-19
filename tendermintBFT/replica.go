package tendermintBFT

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
	*TendermintBFT
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
	r.TendermintBFT = NewTendermintBFT(r)

	r.Register(PaxiBFT.Request{}, r.handleRequest)
	r.Register(Propose{},      r.handlePropose)
	r.Register(ViewChange{},      r.handleAfterPropose)
	r.Register(PreVote{},      r.HandlePreVote)
	r.Register(PreCommit{},    r.HandlePreCommit)

	return r
}
func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("<-----------handleRequest----------->")
	if p.slot <= 0 {
		fmt.Print("-------------------Tendermint-------------------------")
	}
	p.slot++
	if p.slot % 1000 == 0 {
		fmt.Print("p.slot", p.slot)
	}

	log.Debugf("p.slot %v", p.slot)
	p.Member.Addmember(m.NodeID)
	p.Requests = append(p.Requests, &m)

	e,ok := p.log[p.slot]
	if !ok{
		p.log[p.slot] = &entry{
			Ballot:    	p.ballot,
			commit:    	false,
			request:   	&m,
			Timestamp: 	time.Now(),
			PR:			PaxiBFT.NewQuorum(),
			PV: 		PaxiBFT.NewQuorum(),
			PC: 		PaxiBFT.NewQuorum(),
			active:    false,
			Leader:    false,
			VCs: 		PaxiBFT.NewQuorum(),
		}
	}
	e = p.log[p.slot]
	e.request = &m
	w := p.slot % e.PR.Total() + 1
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