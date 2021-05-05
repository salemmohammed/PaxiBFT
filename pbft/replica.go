package pbft

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

// Replica for one Pbft instance
type Replica struct {
	PaxiBFT.Node
	*Pbft
}

// NewReplica generates new Pbft replica
func NewReplica(id PaxiBFT.ID) *Replica {
	r := new(Replica)

	r.Node = PaxiBFT.NewNode(id)
	r.Pbft = NewPbft(r)

	r.Register(PaxiBFT.Request{}, r.handleRequest)
	r.Register(PrePrepare{}, r.HandlePre)
	r.Register(Prepare{}, r.HandlePrepare)
	r.Register(Commit{}, r.HandleCommit)

	return r
}

func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("<---------------------handleRequest------------------------->")
	if p.slot == 0 {
		fmt.Println("-------------------PBFT-------------------------")
	}
	p.slot++ // slot number for every request
	if p.slot%10000 == 0 {
		fmt.Print("p.slot", p.slot)
	}
	//log.Debugf("p.slot %v", p.slot)
	p.requests = append(p.requests, &m)

	e, ok := p.logR[p.slot]
	if !ok {
		p.logR[p.slot] = &RequestSlot{
			request: &m,
			//slot:        p.slot,
			commit: false,
			active: false,
			//Concurrency: 0,
			Leader:   false,
			RecReqst: PaxiBFT.NewQuorum(),
			MissReq:  &m,
		}
	}
	e, ok = p.logR[p.slot]
	p.RecivedReq = true
	e.active = true
	e.request = &m

	_, ok1 := p.log[p.slot]
	if !ok1 {
		p.log[p.slot] = &entry{
			ballot:    p.ballot,
			view:      p.view,
			command:   m.Command,
			commit:    false,
			request:   &m,
			timestamp: time.Now(),
			Digest:    GetMD5Hash(&m),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			Q4:        PaxiBFT.NewQuorum(),
		}
	}
	_, ok1 = p.log[p.slot]
	p.Member.Addmember(m.NodeID)
	if p.ID() == PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(1)){
		log.Debugf("The view leader : %v ", p.ID())
		log.Debugf("The requests[%v] : %v ", p.slot, p.Pbft.requests[p.slot])
		e.Leader = true
		if p.activeView != true{
			p.ballot.Next(p.ID())
			p.view.Next(p.ID())
		}
		p.activeView = true
		p.Pbft.HandleRequest(*p.Pbft.requests[p.slot], p.slot)
	}

	//if p.ID() != PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(1)) {
	//	reply := PaxiBFT.Reply{
	//		Command:    m.Command,
	//		Properties: make(map[string]string),
	//		Timestamp:  time.Now().Unix(),
	//	}
	//	reply.Properties[HTTPHeaderSlot] = strconv.Itoa(p.Pbft.slot)
	//	reply.Properties[HTTPHeaderBallot] = p.Pbft.ballot.String()
	//	reply.Properties[HTTPHeaderExecute] = strconv.Itoa(p.Pbft.execute - 1)
	//	m.Reply(reply)
	//	return
	//}
}
