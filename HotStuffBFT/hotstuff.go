package HotStuffBFT

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
	NEWVIEW
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
	VC         status
	Digest     []byte
}
type HotStuffBFT struct {
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
}
func NewHotStuffBFT(n PaxiBFT.Node, options ...func(*HotStuffBFT)) *HotStuffBFT {
	p := &HotStuffBFT{
		Node:          	 	n,
		log:           	 	make(map[int]*entry, PaxiBFT.GetConfig().BufferSize),
		slot:          	 	-1,
		quorum:        	 	PaxiBFT.NewQuorum(),
		Requests:      	 	make([]*PaxiBFT.Request, 0),
		count:				0,
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
func (p *HotStuffBFT) HandleRequest(r PaxiBFT.Request) {
	log.Debugf("<-------HandleRequest---------->")
	p.Broadcast(Prepare{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       p.slot,
		Request:	r,
	})
}
func (p *HotStuffBFT) handlePrepare(m Prepare) {
	log.Debugf("\n\n<-------P-------------handlePrepare--------P---------->\n\n")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create the log")
		p.log[m.Slot] = &entry{
			Ballot:    	m.Ballot,
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
		}
	}
	e = p.log[m.Slot]

	w := (m.Slot+1) % e.Q1.Total() + 1

	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(w))
	log.Debugf("Node_ID = %v", Node_ID)
	log.Debugf("\n")
	if Node_ID != p.ID(){
		p.Send(Node_ID, Viewchange{
			Ballot:     m.Ballot,
			ID:         p.ID(),
			Slot:       m.Slot,
			Request:    m.Request,
		})

	}
	log.Debugf("++++++++++++++++++++++++++ handlePropose Done ++++++++++++++++++++++++++")
}
func (p *HotStuffBFT) handleViewchange(m Viewchange){
	log.Debugf("\n\n<---R----handleViewchange----R------>\n\n")

	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create the log")
		p.log[m.Slot] = &entry{
			Ballot:    	m.Ballot,
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
		}
	}
	e = p.log[m.Slot]
	e.Q4.ACK(m.ID)
	log.Debugf("e.Q4.Majority() = %v", e.Q4.Majority())
	log.Debugf("e.VC = %v", e.VC)
	log.Debugf("request = %v", m.Request)
	if e.Q4.Majority() && e.VC != NEWVIEW{
		e.Q4.Reset()
		e.VC = NEWVIEW
		e.leader = true
		p.Broadcast(AfterPrepare{
			Ballot:  m.Ballot,
			ID:      p.ID(),
			Slot:    m.Slot,
			Request: m.Request,
		})
	}

}
func (p *HotStuffBFT) handleAfterPrepare(m AfterPrepare){
	log.Debugf("<---R----handleAfterPrepare----R------>")

	e, ok := p.log[m.Slot]
	if !ok{
		log.Debugf("return")
		return
	}

	e.Pstatus = PREPARED
	p.Send(m.ID, ActAfterPrepare{
		Ballot:     m.Ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Digest:    GetMD5Hash(&m.Request),
	})
}
func (p *HotStuffBFT) handleActAfterPrepare(m ActAfterPrepare){
	log.Debugf("<---------V-----------handleActAfterPrepare----------V-------->")
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
	e.Pstatus = PREPARED
	if e.Q1.Majority(){
		e.Q1.Reset()
		p.Broadcast(PreCommit{
		Ballot:     m.Ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Digest:    m.Digest,
	})
	}
}
func (p *HotStuffBFT) handlePreCommit(m PreCommit) {
	log.Debugf("<---------V-----------handlePreCommit----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	p.Send(m.ID, ActPreCommit{
		Ballot:     m.Ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Digest:     m.Digest,
	})
}
func (p *HotStuffBFT) handleActPreCommit(m ActPreCommit) {
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
	if e.Q2.Majority(){
		e.Q2.Reset()
		p.Broadcast(Commit{
			Ballot:     p.ballot,
			ID:         p.ID(),
			Slot:       m.Slot,
			Digest:    m.Digest,
		})
	}
}
func (p *HotStuffBFT) handleCommit(m Commit) {
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
		Digest: m.Digest,
	})
}
func (p *HotStuffBFT) handleActCommit(m ActCommit) {
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
	if e.Q3.Majority() {
		e.Q3.Reset()
		p.Broadcast(Decide{
			Ballot: p.ballot,
			ID:     p.ID(),
			Slot:   m.Slot,
			Digest: m.Digest,
		})

		e.commit = true
		e.Cstatus = COMMITTED

		log.Debugf("e.leader = %v ", e.leader)
		log.Debugf("e.Pstatus = %v", e.Pstatus)
		log.Debugf("e.Cstatus = %v", e.Cstatus)
		log.Debugf("e.Rstatus = %v", e.Rstatus)
	}
	if e.Rstatus == RECEIVED && e.leader == true{
		log.Debugf("p.exec()")
		p.exec()
	}
}
func (p *HotStuffBFT) handleDecide(m Decide) {
	log.Debugf("<--------------------handleDecide------------------>")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}

	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Return")
		return
	}
	e.Cstatus = COMMITTED
	log.Debugf("e.Rstatus = %v", e.Rstatus)
	log.Debugf("e.Pstatus = %v", e.Pstatus)
	log.Debugf("e.Cstatus = %v", e.Cstatus)
	if e.Rstatus == RECEIVED{
		e.commit = true
		p.exec()
	}
}
func (p *HotStuffBFT) exec() {
	log.Debugf("<--------------------exec()------------------>")
	for {
		log.Debugf("p.execute %v", p.execute)
		e, ok := p.log[p.execute]
		if !ok || !e.commit || e.Rstatus != RECEIVED{
			log.Debugf("Break")
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
			log.Debugf("reply= %v", e.request)
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