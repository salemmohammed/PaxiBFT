package HotStuff_SL

import (
	"crypto/md5"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"strconv"
	"sync"
	"time"
)
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
	Q1	  		  *PaxiBFT.Quorum
	Q2    		  *PaxiBFT.Quorum
	Q3    		  *PaxiBFT.Quorum
	Q4    		  *PaxiBFT.Quorum
	active 		  bool
	leader        bool
	Pstatus    status
	Cstatus    status
	Rstatus	   status
	Digest     []byte
	MyTurn      bool
	slot        int
	Node_ID     PaxiBFT.ID
	Sent        bool
}
type HotStuff struct {
	PaxiBFT.Node
	log      					map[int]*entry 				// log ordered by slot
	config 						[]PaxiBFT.ID
	execute 					int             			// next execute slot number
	active  					bool		    			// active leader
	ballot  					PaxiBFT.Ballot     			// highest ballot number
	slot    					int             			// highest slot number
	quorum   					*PaxiBFT.Quorum    			// phase 1 quorum
	Requests 					[]*PaxiBFT.Request 			// phase 1 pending requests
	c              				chan PaxiBFT.Request
	count 		   				int
	leader						bool
	mux 						sync.Mutex
	Missedrequest    	        []*PaxiBFT.Request
	Node_ID                     PaxiBFT.ID
	Sent                        bool
}
func NewHotStuff(n PaxiBFT.Node, options ...func(*HotStuff)) *HotStuff {
	p := &HotStuff{
		Node:          	 	n,
		log:           	 	make(map[int]*entry, PaxiBFT.GetConfig().BufferSize),
		slot:          	 	-1,
		quorum:        	 	PaxiBFT.NewQuorum(),
		Requests:      	 	make([]*PaxiBFT.Request,0),
	}
	for _, opt := range options {
		opt(p)
	}
	return p
}
func GetMD5Hash(r *PaxiBFT.Request) []byte {
	hasher := md5.New()
	hasher.Write([]byte(r.Command.Value))
	return []byte(hasher.Sum(nil))
}
func RemoveIndex(s []*PaxiBFT.Request, index int) []*PaxiBFT.Request {
	if len(s) > 0 {
		return append(s[:index], s[index+1:]...)
	}else {
		return append(s[:index])
	}
}
func (p *HotStuff) HandleRequest(r PaxiBFT.Request, slot int,total int) {
	log.Debugf("<---Start----HandleRequest----Start------>")
	log.Debugf("Request in  loop =%v", r)

	p.Broadcast(Prepare{
	Ballot:     p.ballot,
	ID:         p.ID(),
	Slot:       slot,
	Request:	r,
	})
	w := ((slot % total + 1) + 1)
	if w > total{
		w = (w % total)
	}
	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(w))
	log.Debugf("---Node_ID--- %v", Node_ID)
	p.Send(Node_ID, RoundRobin{Slot: slot+1, Request: r, Id: p.ID()})
	log.Debugf("<---End----HandleRequest----End------>")
}
func (p *HotStuff) handlePrepare(m Prepare) {
	log.Debugf("<-------P-------------handlePrepare--------P---------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)


	//log.Debugf("ListRequest = %v", ListRequest)
	log.Debugf("p.Requests = %v ", p.Requests)

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
			Q4: 		PaxiBFT.NewQuorum(),
			active:     false,
			leader:     false,
			commit:    	false,
			Digest:     GetMD5Hash(&m.Request),
			slot:       m.Slot,
			MyTurn:     false,
		}
	}
	e = p.log[m.Slot]

	p.Send(m.ID, ActPrepare{
		Ballot:     m.Ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Digest:     e.Digest,
		Request:	m.Request,
	})
	log.Debugf("++++++++++++++++++++++++++ handlePropose Done ++++++++++++++++++++++++++")
}
func (p *HotStuff) handleActPrepare(m ActPrepare){
	log.Debugf("<---------V-----------handleActPrepare----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}

	e, ok := p.log[m.Slot]
	if !ok{
		log.Debugf("return")
		return
	}
	e.Q1.ACK(m.ID)
	if e.Q1.Majority() && e.active{
		e.Q1.Reset()
		p.Broadcast(PreCommit{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Digest:     m.Digest,
		Request:	m.Request,
	})
	}
}
func (p *HotStuff) handlePreCommit(m PreCommit) {
	log.Debugf("<---------V-----------handlePreCommit----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}

	p.Send(m.ID, ActPreCommit{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Digest:     m.Digest,
		Request:	m.Request,
	})
}
func (p *HotStuff) handleActPreCommit(m ActPreCommit) {
	log.Debugf("<---------V-----------handleActPreCommit----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}
	e, ok := p.log[m.Slot]
	if !ok{
		log.Debugf("Return")
		return
	}
	e.Q2.ACK(m.ID)
	if e.active && e.Q2.Majority(){
		e.Q2.Reset()
		p.Broadcast(Commit{
			Ballot:     p.ballot,
			ID:         p.ID(),
			Slot:       m.Slot,
			Digest:    m.Digest,
			Request:	m.Request,
		})
	}
}
func (p *HotStuff) handleCommit(m Commit) {
	log.Debugf("<---------V-----------handleCommit----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}
	p.Send(m.ID, ActCommit{
		Ballot:  p.ballot,
		ID:      p.ID(),
		Slot:    m.Slot,
		Digest:  m.Digest,
		Request: m.Request,
	})
}
func (p *HotStuff) handleActCommit(m ActCommit) {
	log.Debugf("<---------V-----------handleActCommit----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}
	e, ok := p.log[m.Slot]
	if !ok{
		log.Debugf("Return")
		return
	}
	e.Q3.ACK(m.ID)
	if e.active && e.Q3.Majority() {
		e.Q3.Reset()
		p.Broadcast(Decide{
			Ballot: p.ballot,
			ID:     p.ID(),
			Slot:   m.Slot,
			Digest: m.Digest,
			Request:	m.Request,
		})
		e.commit = true
		e.Cstatus = COMMITTED
		log.Debugf("e.Rstatus = %v, e.leader = %v", e.Rstatus, e.leader)
		if e.Rstatus == RECEIVED && e.leader == true{
			p.exec()
		}
	}
}
func (p *HotStuff) handleDecide(m Decide) {
	log.Debugf("<--------------------handleDecide------------------>")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}

	e, ok := p.log[m.Slot]
	if !ok{
		log.Debugf("Return")
		return
	}

	e.Cstatus = COMMITTED
	e.Pstatus = PREPARED
	log.Debugf("e.Pstatus = %v", e.Pstatus)
	log.Debugf("e.Cstatus = %v", e.Cstatus)
	if e.Rstatus == RECEIVED{
		e.commit = true
		p.exec()
	}
}
func (p *HotStuff) exec() {
	log.Debugf("<--------------------exec()------------------>")
	for {
		log.Debugf("p.execute %v", p.execute)
		e, ok := p.log[p.execute]
		if !ok{
			log.Debugf("Break ok")
			break
		}
		if !e.commit {
		log.Debugf("Break commit")
		break
		}
		if e.Rstatus != RECEIVED {
			log.Debugf("Not RECEIVED")
			break
		}
		value := p.Execute(e.request.Command)

		if e.request != nil && e.leader{
			reply := PaxiBFT.Reply{
				Command:    e.request.Command,
				Value:      value,
				Properties: make(map[string]string),
			}
			e.request.Reply(reply)
			log.Debugf("********* Reply Primary *********")
			e.request = nil
		}
		if e.request != nil && e.leader == false && e.Rstatus == RECEIVED{
			log.Debugf("********* Replica Request ********* ")
			log.Debugf("reply= %v", e.request)
			reply := PaxiBFT.Reply{
				Command:    e.request.Command,
				Value:      value,
				Properties: make(map[string]string),
			}
			e.request.Reply(reply)
			e.request = nil
			log.Debugf("********* Reply Replicas *********")

		}
		// TODO clean up the log periodically
		delete(p.log, p.execute)
		p.execute++
	}
}