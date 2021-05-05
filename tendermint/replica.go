package tendermint

import (
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"math"
	"strconv"
	"strings"
	"sync"
	)

// Replica for one Tendermint instance
type Replica struct {
	PaxiBFT.Node
	*Tendermint
	mux sync.Mutex
}
const (
	HTTPHeaderSlot       = "Slot"
	HTTPHeaderBallot     = "Ballot"
	HTTPHeaderExecute    = "Execute"
	HTTPHeaderInProgress = "Inprogress"
)
// NewReplica generates new Pbft replica
func NewReplica(id PaxiBFT.ID) *Replica {
	log.Debugf("Replica started \n")
	r := new(Replica)

	r.Node = PaxiBFT.NewNode(id)
	r.Tendermint = NewTendermint(r)

	r.Register(PaxiBFT.Request{}, r.handleRequest)
	r.Register(Propose{},      r.handlePropose)
	r.Register(PreVote{},      r.HandlePreVote)
	r.Register(PreCommit{},    r.HandlePreCommit)

	return r
}
func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("\n<-----------handleRequest----------->\n")
	if p.slot <= 0 {
		fmt.Print("-------------------Tendermint-------------------------")
	}
	p.slot++ // slot number for every request
	if p.slot % 1000 == 0 {
		fmt.Print("p.slot", p.slot)
	}


	log.Debugf("Node's ID:%v, command:%v ", p.ID(), m.Command.Key)
	//p.slot++
	log.Debugf("p.slot %v", p.slot)
	p.Member.Addmember(m.NodeID)
	p.Requests = append(p.Requests, &m)


	e, ok := p.logR[p.slot]
	if !ok {
		p.logR[p.slot] = &RequestSlot{
			request:     &m,
			slot:		 p.slot,
			RecReqst:    PaxiBFT.NewQuorum(),
			commit:      false,
			count:       0,
			Neibors:     make([]PaxiBFT.ID, 0),
			active:      false,
			Concurrency: 0,
			Leader:      false,
		}
	}
	e, ok = p.logR[p.slot]

	if e.Leader == false {
		num := int(math.Mod(p.quorum.INC(), 2))
		log.Debugf(" the number is %v", num)
		num = num + 1
		var s string = string(p.Tendermint.ID())
		var b string
		if strings.Contains(s, ".") {
			split := strings.SplitN(s, ".", 2)
			b = split[1]
		} else {
			b = s
		}
		i, _ := strconv.Atoi(b)
		log.Debugf("<<<<<<<<<<<<<<<< num >>>>>>>>>>>>>>>>>> %v, i%v, p.slot:%v", num, i, p.slot)
		if i == num {
			e.Leader = true
			p.ballot.Next(p.ID())
			log.Debugf("p.ballot %v ", p.ballot)
			log.Debugf("<<<<<<<<<<<<<<<<<< The leader for >>>>>>>>>>>>>>>>>>>>> %v ", i)
			p.Tendermint.view.Next(p.ID())
		}else{
			e.request = &m
		}
	}
	if e.Leader == true {
		e.commit = true
		log.Debugf("p.Requests[%v]: %v", e.slot, e.request)
		p.Tendermint.HandleRequest(*e.request, e.slot)
	} //else{
	//	log.Debugf(" Late message and this will be executed")
	//	p.HandlePreCommit(PreCommit{
	//		Ballot:     p.ballot,
	//		ID:         p.ID(),
	//		Request:    m,
	//		Command:   m.Command,
	//		Slot:       p.slot,
	//		Commit:     false,
	//	})
	//}
}
