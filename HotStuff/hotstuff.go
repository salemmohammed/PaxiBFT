package HotStuff

import (
	//"container/list"
	"github.com/ailidani/paxi"
	"github.com/ailidani/paxi/log"
	"strconv"
	"sync"
	"time"
)
// log's entries
type entry struct {
	Ballot     	  paxi.Ballot
	commit     	  bool
	request    	  *paxi.Request
	Timestamp  	  time.Time
	command    	  paxi.Command
	Q1	  		  *paxi.Quorum
	Q2    		  *paxi.Quorum
	Q3    		  *paxi.Quorum
	Q4    		  *paxi.Quorum
	active 		  bool
	Precommitmsg  bool
}
type RequestSlot struct {
	request    	  *paxi.Request
	RecReqst	  *paxi.Quorum
	commit     	  bool
	count 		  int
	Neibors  	  []paxi.ID
	active 		  bool
	Concurrency   int
	Leader		  bool
	slot		  int
}
type HotStuff struct {
	paxi.Node
	log      					map[int]*entry 				// log ordered by slot
	logR      					map[int]*RequestSlot 		// log ordered by slot for receiving requests
	config 						[]paxi.ID
	execute 					int             			// next execute slot number
	active  					bool		    			// active leader
	ballot  					paxi.Ballot     			// highest ballot number
	slot    					int             			// highest slot number
	quorum   					*paxi.Quorum    			// phase 1 quorum
	Requests 					[]*paxi.Request 			// phase 1 pending requests
	view     					paxi.View 	    			// view number
	MyCommand       			paxi.Command
	MyRequests					*paxi.Request
	TemID						paxi.ID
	RequestFlag	   				bool
	c              				chan paxi.Request
	done           				chan bool
	ChooseID	  				paxi.ID
	count 		   				int
	Created						bool
	Leader						bool
	EarlyPropose				bool
	mux 						sync.Mutex
	Plist						[]paxi.ID
	Member                      *paxi.Memberlist
}
func NewHotStuff(n paxi.Node, options ...func(*HotStuff)) *HotStuff {
	p := &HotStuff{
		Node:          	 	n,
		log:           	 	make(map[int]*entry, paxi.GetConfig().BufferSize),
		logR:           	make(map[int]*RequestSlot, paxi.GetConfig().BufferSize),
		slot:          	 	-1,
		quorum:        	 	paxi.NewQuorum(),
		Requests:      	 	make([]*paxi.Request, 0),
		RequestFlag: 		false,
		count:				0,
		EarlyPropose:		false,
		Leader:				false,
		Plist:				make([]paxi.ID,0),
		Member:         	paxi.NewMember(),
	}
	for _, opt := range options {
		opt(p)
	}
	return p
}
// HandleRequest handles request and start phase 1 or phase 2
func (p *HotStuff) HandleRequest(r paxi.Request, s int) {
	log.Debugf("\n<---R----HandleRequest----R------>\n")

	p.logR[s].active = true
	p.log[s] = &entry{
		Ballot:    	p.ballot,
		commit:    	false,
		request:   	&r,
		Timestamp: 	time.Now(),
		Q1:			paxi.NewQuorum(),
		Q2: 		paxi.NewQuorum(),
		Q3: 		paxi.NewQuorum(),
		Q4: 		paxi.NewQuorum(),
		command:    r.Command,
		active:     p.logR[s].active,
	}
	p.Broadcast(Prepare{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       s,
		Command:    p.log[s].command,
		Request:	r,
		Leader:     true,
	})
}
// handle propose from primary
func (p *HotStuff) handlePrepare(m Prepare) {
	log.Debugf("\n<-------P-------------handlePrepare--------P---------->\n")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)
	log.Debugf("p.slots %v", p.slot)

	if m.Ballot > p.ballot {
		log.Debugf("m is bigger m.Ballot:%v, p.ballot:%v", m.Ballot, p.ballot)
		p.ballot = m.Ballot
	}
	_, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create the log")
		p.log[m.Slot] = &entry{
			Ballot:     p.ballot,
			commit:     false,
			Timestamp:  time.Now(),
			request:    &m.Request,
			command:    m.Command,
		}
	}

	e1, ok1 := p.logR[m.Slot]
	if !ok1 {
		p.logR[m.Slot] = &RequestSlot{
			Leader:   false,
		}
	}
	e1, ok1 = p.logR[m.Slot]
	e1.Leader = false

	_, ok = p.log[m.Slot]
	p.Send(m.ID, ActPrepare{
		Ballot:     m.Ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Command:    p.log[m.Slot].command,
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
		return
	}
	e.Q1.ACK(m.ID)
	if p.logR[m.Slot].active && e.Q1.Majority(){
		e.Q1.Reset()
		p.Broadcast(PreCommit{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       m.Slot,
		Command:    m.Command,
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
		Command:    m.Command,
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
		return
	}
	e.Q2.ACK(m.ID)
	if p.logR[m.Slot].active && e.Q2.Majority(){
		e.Q2.Reset()
		p.Broadcast(Commit{
			Ballot:     p.ballot,
			ID:         p.ID(),
			Slot:       m.Slot,
			Command:    m.Command,
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
		Command: m.Command,
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
	if p.logR[m.Slot].active && e.Q3.Majority(){
		e.Q3.Reset()
		p.Broadcast(Decide{
			Ballot:  p.ballot,
			ID:      p.ID(),
			Slot:    m.Slot,
			Command: m.Command,
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
	_, ok1 := p.logR[m.Slot]
	if !ok1{
		log.Debugf("e1, ok1 := p.logR[m.Slot]")
		return
	}

	p.Send(m.ID, ActDecide{
			Ballot:  p.ballot,
			ID:      p.ID(),
			Slot:    m.Slot,
			Command: m.Command,
		})
	e.commit = true
	//	//e1.Leader = false
	//	log.Debugf("Start exec()")
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
	_, ok1 := p.logR[m.Slot]
	if !ok1{
		log.Debugf("failed p.logR[m.Slot]")
		return
	}
	e.Q4.ACK(m.ID)
	if e.Q4.Majority() && p.ID() == e.request.NodeID{
		//e.Q4.Reset()
		e.commit = true
		p.exec()
	}
}
func (p *HotStuff) exec() {
	for {
		log.Debugf("in exece: %v", p.execute)
		e, ok := p.log[p.execute]
		if !ok {
			log.Debugf("return")
			return
		}

		if !ok || !e.commit {
			log.Debugf("break")
			break
		}
		log.Debugf("e.active %v", e.active)
		// log.Debugf("Replica %s execute [s=%d, cmd=%v]", p.ID(), p.execute, e.command)
		value := p.Execute(e.command)
		if e.request != nil && e.active{
			reply := paxi.Reply{
				Command:    p.logR[p.execute].request.Command,
				Value:      value,
				Properties: make(map[string]string),
			}
			reply.Properties[HTTPHeaderSlot] = strconv.Itoa(p.execute)
			reply.Properties[HTTPHeaderBallot] = e.Ballot.String()
			reply.Properties[HTTPHeaderExecute] = strconv.Itoa(p.execute)

			p.logR[p.execute].request.Reply(reply)
			e.request = nil
			e1, _ := p.logR[p.execute]
			e1.Leader = false
			e.active = false
		}
		// TODO clean up the log periodically
		log.Debugf("Deletion")
		delete(p.log, p.execute)
		delete(p.logR, p.execute)
		p.execute++
	}
}