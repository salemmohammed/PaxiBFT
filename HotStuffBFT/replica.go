package HotStuffBFT
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
	*HotStuffBFT
	mux sync.Mutex
}
const (
	HTTPHeaderSlot       = "Slot"
	HTTPHeaderBallot     = "Ballot"
	HTTPHeaderExecute    = "Execute"
	HTTPHeaderInProgress = "Inprogress"
)
func NewReplica(id PaxiBFT.ID) *Replica {
	log.Debugf("Replica started")
	r := new(Replica)
	r.Node = PaxiBFT.NewNode(id)
	r.HotStuffBFT = NewHotStuffBFT(r)
	r.Register(PaxiBFT.Request{},  r.handleRequest)
	//*******************************************************
	r.Register(Prepare{},          r.handlePrepare)
	//*******************************************************
	r.Register(Viewchange{},       r.handleViewchange)
	//*******************************************************
	r.Register(AfterPrepare{},     r.handleAfterPrepare)
	r.Register(ActAfterPrepare{},  r.handleActAfterPrepare)
	//*******************************************************
	r.Register(PreCommit{},        r.handlePreCommit)
	r.Register(ActPreCommit{},     r.handleActPreCommit)
	//*******************************************************
	r.Register(Commit{},           r.handleCommit)
	r.Register(ActCommit{},        r.handleActCommit)
	//*******************************************************
	r.Register(Decide{},           r.handleDecide)
	//*******************************************************
	return r
}
func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("<-----------handleRequest----------->")
	if p.slot <= 0 {
		fmt.Print("-------------------HotStuffBFT-------------------------")
	}
	p.slot++
	p.Requests = append(p.Requests, &m)
	e, ok := p.log[p.slot]
	if !ok {
		p.log[p.slot] = &entry{
			Ballot:    	p.ballot,
			request:   	&m,
			Timestamp: 	time.Now(),
			Q1:			PaxiBFT.NewQuorum(),
			Q2: 		PaxiBFT.NewQuorum(),
			Q3: 		PaxiBFT.NewQuorum(),
			Q4: 		PaxiBFT.NewQuorum(),
			active:     false,
			leader:     false,
			commit:    	false,
			Digest:     GetMD5Hash(&m),
		}
	}
	e = p.log[p.slot]
	e.request = &m

	log.Debugf("-------------------------")
	log.Debugf("request= %v" , m)
	log.Debugf("e.request= %v" , e.request)
	log.Debugf("slot = %v", p.slot)
	log.Debugf("-------------------------")
    e.Digest  = GetMD5Hash(&m)
	w := p.slot % e.Q1.Total() + 1
	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(w))
	log.Debugf("Node_ID = %v", Node_ID)

	if Node_ID == p.ID(){
		log.Debugf("\n\n-------------\n\n")
		log.Debugf("leader       = %v ", p.ID())
		log.Debugf("\n\n-------------\n\n")
		e.active = true
	}
	e.Rstatus = RECEIVED
	if e.active == true {
		//e.leader = true
		p.ballot.Next(p.ID())
		log.Debugf("p.ballot %v ", p.ballot)
		e.Ballot = p.ballot
		e.Pstatus = PREPARED
		p.HandleRequest(m)
	}
	log.Debugf("e.Pstatus = %v", e.Pstatus)
	log.Debugf("e.Cstatus = %v", e.Cstatus)
	log.Debugf("e.Rstatus = %v", e.Rstatus)

	if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED{
		log.Debug("late call")
		e.commit = true
		p.exec()
	}
}