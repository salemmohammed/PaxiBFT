package tendStar

import (
	//"container/list"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"strconv"
	"sync"
	"time"
)
// log's entries
type status int8
const (
	NONE status = iota
	PREPREPARED
	PREPARED
	COMMITTED
	RECEIVED
	NEWCHAMGED
)
type entry struct {

	Ballot     	  PaxiBFT.Ballot
	commit     	  bool
	request    	  *PaxiBFT.Request
	Timestamp  	  time.Time
	command    	  PaxiBFT.Command

	Q1	  		  *PaxiBFT.Quorum
	Q2    		  *PaxiBFT.Quorum
	Q3    		  *PaxiBFT.Quorum

	active 		  bool
	Precommitmsg  bool

	Leader		  bool

	MyTurn        bool
	Node_ID       PaxiBFT.ID
	Pstatus    status
	Cstatus    status
	Rstatus	   status
	Slot       int
	Sent        bool

}

type Tendermint struct {

	PaxiBFT.Node
	log      					map[int]*entry 				// log ordered by slot
	config 						[]PaxiBFT.ID
	execute 					int             			// next execute slot number
	active  					bool		    			// active leader
	ballot  					PaxiBFT.Ballot     			// highest ballot number
	slot    					int             			// highest slot number
	quorum   					*PaxiBFT.Quorum    			// phase 1 quorum
	Requests 					[]*PaxiBFT.Request 			// phase 1 pending requests
	view     					PaxiBFT.View 	    			// view number
	Member						*PaxiBFT.Memberlist
	MyCommand       			PaxiBFT.Command
	MyRequests					*PaxiBFT.Request
	TemID						PaxiBFT.ID
	RequestFlag	   				bool
	c              				chan PaxiBFT.Request
	done           				chan bool
	ChooseID	  				PaxiBFT.ID
	count 		   				int
	Created						bool
	Leader						bool
	EarlyPropose				bool
	mux 						sync.Mutex
	Plist						[]PaxiBFT.ID

	Sent          				bool
	MyTurn        				bool
	Node_ID                     PaxiBFT.ID
}
// NewPaxos creates new paxos instance
func NewTendermint(n PaxiBFT.Node, options ...func(*Tendermint)) *Tendermint {

	p := &Tendermint{
		Node:          	 	n,
		log:           	 	make(map[int]*entry, PaxiBFT.GetConfig().BufferSize),

		slot:          	 	-1,
		quorum:        	 	PaxiBFT.NewQuorum(),
		Requests:      	 	make([]*PaxiBFT.Request, 0),
		Member:         	PaxiBFT.NewMember(),
		RequestFlag: 		false,
		count:				0,
		EarlyPropose:		false,
		Leader:				false,
		Plist:				make([]PaxiBFT.ID,0),
	}
	for _, opt := range options {
		opt(p)
	}
	return p

}
// HandleRequest handles request and start phase 1 or phase 2
func (p *Tendermint) HandleRequest(r PaxiBFT.Request, s int, total int) {

	log.Debugf("\n<---R----HandleRequest----R------>\n")
	log.Debugf("Sender ID %v, slot=%v", r.NodeID, s)

	p.Broadcast(Propose{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Request:    r,
		View:       p.view,
		Slot:       s,
	})

	w := ((s % total + 1) + 1)
	if w > total{
		w = (w % total)
	}
	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(w))
	log.Debugf("---Node_ID--- %v", Node_ID)
	p.Send(Node_ID, RoundRobin{Slot: s+1, Request: r, Id: p.ID()})
	log.Debugf("<---End----HandleRequest----End------>")
}
// handle propose from primary
func (p *Tendermint) handlePropose(m Propose) {
	log.Debugf("\n<-------P-------------handlePropose--------P---------->\n")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	if m.Ballot > p.ballot {
		log.Debugf("m is bigger m.Ballot:%v, p.ballot:%v", m.Ballot, p.ballot)
		p.ballot = m.Ballot
	}
	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create the log")
		p.log[m.Slot] = &entry{
			Ballot:    	p.ballot,
			request:   	&m.Request,
			Timestamp: 	time.Now(),
			Q1:			PaxiBFT.NewQuorum(),
			Q2: 		PaxiBFT.NewQuorum(),
			Q3: 		PaxiBFT.NewQuorum(),
			active:     false,
			Leader:		false,
			commit:    	false,
			Sent:       false,
			MyTurn:     false,
			Slot:       p.slot,
		}
	}
	e = p.log[m.Slot]
	p.Send(m.ID, ActPropose{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Request:    *e.request,
	})
	log.Debugf("++++++++++++++++++++++++++ handlePropose Done ++++++++++++++++++++++++++")
}

func (p *Tendermint) HandleActPropose(m ActPropose) {
	log.Debugf("\n\n\n<---------V-----------HandleActPropose----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}
	e, ok := p.log[m.Slot]
	if !ok{
		return
	}
	e.Q1.ACK(m.ID)
	if e.active && e.Q1.Majority(){
		e.Q1.Reset()
		p.Broadcast(PreVote{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Request:    m.Request,
	})
}
}

func (p *Tendermint) HandlePreVote(m PreVote) {
	log.Debugf("\n\n\n<---------V-----------HandlePreVote----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}

	p.Send(m.ID, ActPreVote{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Request:    m.Request,
	})
}

func (p *Tendermint) HandleActPreVote(m ActPreVote) {
	log.Debugf("\n\n\n<---------V-----------HandleActPreVote----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}
	e, ok := p.log[m.Slot]
	if !ok{
		return
	}
	e.Q2.ACK(m.ID)
	if e.active && e.Q2.Majority(){
		e.Q2.Reset()
		p.Broadcast(PreCommit{
			Ballot:     p.ballot,
			ID:         p.ID(),
			Slot:       m.Slot,
			Request:    m.Request,
			Commit:		true,
		})
	}
}

func (p *Tendermint) HandlePreCommit(m PreCommit) {
	log.Debugf("\n\n<----------C----------PreCommit--------C---------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}

	if p.ID() != m.ID{
		p.Send(m.ID, ActPreCommit{
			Ballot:  p.ballot,
			ID:      p.ID(),
			Slot:    m.Slot,
			Request: m.Request,
		})
	}

	e, ok := p.log[m.Slot]
	if !ok{
		log.Debugf("failed p.log")
		return
	}
	if m.Commit == false && e.commit == false{
		log.Debugf("return")
		return
	}
	e.commit = true
	e.Cstatus = COMMITTED
	e.Pstatus = PREPARED
	if e.Rstatus == RECEIVED && e.Leader == false{
		p.exec()
	}
}
func (p *Tendermint) HandleActPreCommit(m ActPreCommit) {
	log.Debugf("\n\n<----------C----------HandlePreCommit--------C---------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	if m.Ballot > p.ballot {
		log.Debugf("m is bigger than p")
		p.ballot = m.Ballot
	}
	e, ok := p.log[m.Slot]
	if !ok {
		return
	}
	e.Q3.ACK(m.ID)
	if e.Q3.Majority() {
		e.commit = true
		e.Q3.Reset()
		p.exec()
	}
}
func (p *Tendermint) exec() {
	log.Debugf("<--------------------exec()------------------>")
	for {
		log.Debugf("p.execute %v", p.execute)
		e, ok := p.log[p.execute]
		if !ok || !e.commit {
			log.Debugf("Break")
			break
		}
		if e.Rstatus != RECEIVED {
			log.Debugf("Not RECEIVED")
			break
		}
		value := p.Execute(e.request.Command)
		reply := PaxiBFT.Reply{
			Command:    e.request.Command,
			Value:      value,
			Properties: make(map[string]string),
		}

		if e.request != nil && e.Leader {
			e.request.Reply(reply)
			log.Debugf("********* Reply Primary *********")
			e.request = nil
		}else{
			log.Debugf("********* Replica Request ********* ")
			log.Debugf("reply= %v", e.request)
			e.request.Reply(reply)
			e.request = nil
			log.Debugf("********* Reply Replicas *********")

		}
		// TODO clean up the log periodically
		delete(p.log, p.execute)
		p.execute++
	}
}