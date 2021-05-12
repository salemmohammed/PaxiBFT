package pbftBFT

import (
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"strconv"
	"time"
)
const (
	HTTPHeaderSlot    = "Slot"
	HTTPHeaderBallot  = "Ballot"
	HTTPHeaderExecute = "Execute"
)
type Replica struct {
	PaxiBFT.Node
	*Pbftbft
}
func NewReplica(id PaxiBFT.ID) *Replica {

	r := new(Replica)
	r.Node = PaxiBFT.NewNode(id)
	r.Pbftbft = NewPbftBFT(r)

	r.Register(PaxiBFT.Request{}, r.handleRequest       )
	r.Register(PrePrepare{},      r.HandlePre           )
	r.Register(ViewChange{},      r.HandleViewChange    )
	r.Register(NewChange{},       r.HandleNewChange     )
	r.Register(SecondPrePrepare{},r.HandlePreAfterChange)
	r.Register(Prepare{},         r.HandlePrepare       )
	r.Register(Commit{},          r.HandleCommit        )

	return r
}
func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("<---------------------handleRequest------------------------->")
	p.slot++ // slot number for every request
	if p.slot%1000 == 0 {
		fmt.Print("p.slot", p.slot)
	}

	e, ok := p.log[p.slot]
	if !ok {
		p.log[p.slot] = &entry{
			ballot:    p.ballot,
			commit:    false,
		    active:    false,
			Leader:    false,
			request:   &m,
			timestamp: time.Now(),
			Digest:    GetMD5Hash(&m),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			Q4:        PaxiBFT.NewQuorum(),

		}
	}
	e = p.log[p.slot]
	e.request = &m

	log.Debugf("p.slot = %v ", p.slot)
	log.Debugf("Key = %v ", m.Command.Key)

	if e.commit{
		log.Debugf("Executed")
		p.exec()
	}

	if p.slot == 0 {
		fmt.Println("-------------------PBFTBFT-------------------------")
	}

	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(1))
	log.Debugf("Node_ID = %v", Node_ID)

	if Node_ID == p.ID(){
		log.Debugf("The Leader is malicious = %v", p.ID())
		e.active = true
		e.Leader = true
	}

	p.ballot.Next(p.ID())
	p.requests = append(p.requests, &m)
	if e.Leader{
		p.Pbftbft.HandleRequest(m, p.slot)
	}
	e.Rstatus = RECEIVED
	if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED{
		e.commit = true
		p.exec()
	}
}
