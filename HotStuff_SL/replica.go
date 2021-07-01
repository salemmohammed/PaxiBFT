package HotStuff_SL
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
	*HotStuff
	mux sync.Mutex
}
const (
	HTTPHeaderSlot       = "Slot"
	HTTPHeaderBallot     = "Ballot"
	HTTPHeaderExecute    = "Execute"
	HTTPHeaderInProgress = "Inprogress"
)

var t int
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

	r.Register(Decide{},           r.handleDecide)

	r.Register(RoundRobin{},       r.handleRound)

	return r
}
func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("<-----------handleRequest----------->")
	if p.slot <= 0 {
		fmt.Print("-------------------HotStuff-------------------------")
	}
	if p.slot % 1000 == 0 {
		fmt.Print("-------------------HotStuff-------------------------")
	}
	p.slot++
	e, ok := p.log[p.slot]
	if !ok {
		log.Debugf("created")
		p.log[p.slot] = &entry{
			Ballot:    p.ballot,
			request:   &m,
			Timestamp: time.Now(),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			Q4:        PaxiBFT.NewQuorum(),
			active:    false,
			leader:    false,
			commit:    false,
			MyTurn:    false,
			Digest:    GetMD5Hash(&m),
			slot:      p.slot,
			Sent:      false,
		}
	}
	e = p.log[p.slot]
	t = e.Q1.Total()
	e.request = &m
	log.Debugf("-----------replica--------------")
	log.Debugf("e.request= %v" , e.request)
	log.Debugf("slot = %v", p.slot)

    e.Digest  = GetMD5Hash(&m)
	w := p.slot % e.Q1.Total() + 1
	p.Node_ID = PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(w))
	log.Debugf("p.Node_ID = %v", p.Node_ID)

	if p.Node_ID == p.ID() {
		fmt.Println("w=%v", w)
		log.Debugf(" The request appended = %v ", m.Command.Key)
		log.Debugf("leader")
		e.active = true
		e.leader = true
		p.ballot.Next(p.ID())
		log.Debugf("p.ballot %v ", p.ballot)
		e.Ballot = p.ballot
		e.Pstatus = PREPARED
		//p.Requests = append(p.Requests, &m)
	}

	if (p.ID() == p.Node_ID){
		log.Debugf("p.slot module e.Q2.Total1() == 0 = %v ", (p.slot % e.Q2.Total1()))
		if p.slot % e.Q2.Total1() == 0 && e.Sent == false{
			e.Sent = true
			log.Debugf("slot = %v", p.slot)
			p.HandleRequest(m,p.slot,t)
		}
		log.Debugf(" value = %v, e.MyTurn = %v ", e.slot , e.MyTurn)
		if p.slot > 0 && e.MyTurn == true && e.Sent == false{
			e.Sent = true
			log.Debugf("slot = %v", p.slot)
			p.HandleRequest(m,p.slot,t)
		}
	}
	e.Rstatus = RECEIVED
	log.Debugf("e.Pstatus = %v", e.Pstatus)
	log.Debugf("e.Cstatus = %v", e.Cstatus)
	if e.Cstatus == COMMITTED && e.Pstatus == PREPARED {
		log.Debug("late call")
		e.commit = true
		p.exec()
	}
}
func (p *HotStuff) handleRound(m RoundRobin) {
	log.Debugf("\n<-----------handleRound----------->\n")
	log.Debugf("p.requests = %v ", p.Requests)
	log.Debugf("m.Slot = %v ", m.Slot)
	log.Debugf("p.id = %v ", m.Id)
	if p.slot >= m.Slot {
		log.Debugf("p.slot >= m.Slot")
		e, ok := p.log[m.Slot]
		if !ok {
			log.Debugf("!ok")
		}else{
			log.Debugf("e.Sent %v ", e.Sent)
			if e.Sent == false {
				p.HandleRequest(*e.request, m.Slot, t)
			}
	}
	}else {
		log.Debugf("p.slot < m.Slot")
		_, ok := p.log[m.Slot]
		if !ok {
			log.Debugf("created")
			p.log[m.Slot] = &entry{
				Ballot:    p.ballot,
				request:   &m.Request,
				Timestamp: time.Now(),
				Q1:        PaxiBFT.NewQuorum(),
				Q2:        PaxiBFT.NewQuorum(),
				Q3:        PaxiBFT.NewQuorum(),
				Q4:        PaxiBFT.NewQuorum(),
				active:    false,
				leader:    false,
				commit:    false,
				Sent:      false,
				MyTurn:    true,
				Digest:    GetMD5Hash(&m.Request),
				slot:      m.Slot,
			}
		}
	}
}