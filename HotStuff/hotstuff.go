package HotStuff

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
}

func NewHotStuff(n PaxiBFT.Node, options ...func(*HotStuff)) *HotStuff {
	p := &HotStuff{
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
func (p *HotStuff) HandleRequest(r PaxiBFT.Request) {
	log.Debugf("\n<---R----HandleRequest----R------>\n")

	p.Broadcast(Prepare{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       p.slot,
		Request:	r,
	})
}
func (p *HotStuff) handlePrepare(m Prepare) {
	log.Debugf("\n<-------P-------------handlePrepare--------P---------->\n")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	log.Debugf("p.slots %v", p.slot)

	if m.Ballot > p.ballot {
		log.Debugf("m is bigger m.Ballot:%v, p.ballot:%v", m.Ballot, p.ballot)
		p.ballot = m.Ballot
	}

	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create the log")
		p.log[m.Slot] = &entry{
			Ballot:     p.ballot,
			commit:     false,
			Timestamp:  time.Now(),
			request:    &m.Request,
		}
	}

	e = p.log[m.Slot]
	e.Digest = GetMD5Hash(&m.Request)

	p.Send(m.ID, ActPrepare{
		Ballot:     m.Ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Digest:     e.Digest,
	})
	log.Debugf("++++++++++++++++++++++++++ handlePropose Done ++++++++++++++++++++++++++")
}
func (p *HotStuff) handleActPrepare(m ActPrepare){
	log.Debugf("\n\n\n<---------V-----------handleActPrepare----------V-------->")
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
	if e.Q1.Majority(){
		e.Q1.Reset()
		p.Broadcast(PreCommit{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Digest:    m.Digest,
	})
	}
}
func (p *HotStuff) handlePreCommit(m PreCommit) {
	log.Debugf("\n\n\n<---------V-----------handlePreCommit----------V-------->")
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
		Digest:    m.Digest,
	})
}
func (p *HotStuff) handleActPreCommit(m ActPreCommit) {
	log.Debugf("\n\n\n<---------V-----------handleActPreCommit----------V-------->")
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
func (p *HotStuff) handleCommit(m Commit) {
	log.Debugf("\n\n\n<---------V-----------handleCommit----------V-------->")
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
func (p *HotStuff) handleActCommit(m ActCommit) {
	log.Debugf("\n\n\n<---------V-----------handleActCommit----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}
	e, ok := p.log[m.Slot]
	if !ok{
		log.Debugf("Create the log")
		return
	}
	e.Q3.ACK(m.ID)
	if e.Q3.Majority(){
		e.Q3.Reset()
		p.Broadcast(Decide{
			Ballot:  p.ballot,
			ID:      p.ID(),
			Slot:    m.Slot,
			Digest: m.Digest,
		})
	}
}
func (p *HotStuff) handleDecide(m Decide) {
	log.Debugf("\n\n\n<---------V-----------handleDecide----------V-------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}

	e, ok := p.log[m.Slot]
	if !ok{
		log.Debugf("e, ok := p.log[m.Slot]")
		return
	}
	p.Send(m.ID, ActDecide{
			Ballot:  p.ballot,
			ID:      p.ID(),
			Slot:    m.Slot,
			Digest: m.Digest,
		})
	e.commit = true
	p.exec()
}

func (p *HotStuff) handleActDecide(m ActDecide) {
	log.Debugf("\n\n<----------C----------handleActDecide--------C---------->")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}
	e, ok := p.log[m.Slot]
	if !ok{
		log.Debugf("failed p.log")
		return
	}
	e.Q4.ACK(m.ID)
	if e.Q4.Majority() && p.ID() == e.request.NodeID{
		e.Q4.Reset()
		e.commit = true
		p.exec()
	}
}
func (p *HotStuff) exec() {
	for {
		log.Debugf("in exece: %v", p.execute)
		e, ok := p.log[p.execute]
		if !ok || !e.commit {
			log.Debugf("break")
			break
		}
		log.Debugf("e.active %v", e.active)

		value := p.Execute(e.request.Command)
		if e.request != nil && e.active{
			reply := PaxiBFT.Reply{
				Command:    e.request.Command,
				Value:      value,
				Properties: make(map[string]string),
			}
			reply.Properties[HTTPHeaderSlot] = strconv.Itoa(p.execute)
			reply.Properties[HTTPHeaderBallot] = e.Ballot.String()
			reply.Properties[HTTPHeaderExecute] = strconv.Itoa(p.execute)
			e.request.Reply(reply)
			e.request = nil
		}
		// TODO clean up the log periodically
		delete(p.log, p.execute)
		p.execute++
	}
}