package streamlet

import (
	//"container/list"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"strconv"
	"time"
	"sync"
)
// log's entries
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

}

type RequestSlot struct {

	request    	  *PaxiBFT.Request
	RecReqst	  *PaxiBFT.Quorum
	commit     	  bool
	count 		  int
	Neibors  	  []PaxiBFT.ID
	active 		  bool
	Concurrency   int
	Leader		  bool
	slot		  int

}

type Tendermint struct {

	PaxiBFT.Node
	log      					map[int]*entry 				// log ordered by slot
	logR      					map[int]*RequestSlot 		// log ordered by slot for receiving requests
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
	Delta 						int

}
// NewPaxos creates new paxos instance
func NewTendermint(n PaxiBFT.Node, options ...func(*Tendermint)) *Tendermint {

	p := &Tendermint{
		Node:          	 	n,
		log:           	 	make(map[int]*entry, PaxiBFT.GetConfig().BufferSize),
		logR:           	make(map[int]*RequestSlot, PaxiBFT.GetConfig().BufferSize),
		slot:          	 	-1,
		quorum:        	 	PaxiBFT.NewQuorum(),
		Requests:      	 	make([]*PaxiBFT.Request, 0),
		Member:         	PaxiBFT.NewMember(),
		RequestFlag: 		false,
		count:				0,
		EarlyPropose:		false,
		Leader:				false,
		Plist:				make([]PaxiBFT.ID,0),
		Delta:				0,
	}
	for _, opt := range options {
		opt(p)
	}
	return p

}
// HandleRequest handles request and start phase 1 or phase 2
func (p *Tendermint) HandleRequest(r PaxiBFT.Request, s int) {

	log.Debugf("\n<---R----HandleRequest----R------>\n")
	log.Debugf("Sender ID %v, slot=%v", r.NodeID, s)

	time.Sleep(time.Duration(p.Delta)*time.Millisecond)
	p.logR[s].active = true

	p.log[s] = &entry{
		Ballot:    	p.ballot,
		commit:    	false,
		request:   	&r,
		Timestamp: 	time.Now(),
		Q1:			PaxiBFT.NewQuorum(),
		Q2: 		PaxiBFT.NewQuorum(),
		Q3: 		PaxiBFT.NewQuorum(),
		command:    r.Command,
		active:     p.logR[s].active,
	}

	p.Broadcast(Propose{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Request:    *p.log[s].request,
		View:       p.view,
		Slot:       s,
		Command:    p.log[s].command,
	})
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
	_, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create the log")
		p.log[m.Slot] = &entry{
			Ballot:    p.ballot,
			commit:    true,
			request:   &m.Request,
			Timestamp: time.Now(),
			command:   m.Command,
		}
	}
	_, ok = p.log[m.Slot]
	p.Send(m.ID, ActPreCommit{
		Ballot:  p.ballot,
		ID:      p.ID(),
		Slot:    m.Slot,
		Request: m.Request,
		Command: m.Command,
		Commit:  true,
	})
	p.exec()
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
	//p.Broadcast(PreCommit{
	//	Ballot:     p.ballot,
	//	ID:         p.ID(),
	//	Request:    m.Request,
	//	Slot:       m.Slot,
	//	Command:    m.Request.Command,
	//	Commit:     true,
	//})

	e.Q3.ACK(m.ID)
	if e.Q3.Majority() {
		e.commit = true
		e.Q3.Reset()
		p.exec()
	}
}

//func (p *Tendermint) HandlePreCommit(m PreCommit) {
//	log.Debugf("\n\n<----------C----------HandlePreCommit--------C---------->")
//	log.Debugf("m.slot %v", m.Slot)
//	log.Debugf("sender %v", m.ID)
//	if m.Ballot > p.ballot {
//		log.Debugf("m is bigger than p")
//		p.ballot = m.Ballot
//	}
//	e, ok := p.log[m.Slot]
//	if !ok {
//		return
//	}
//	if e.commit != true{
//		e.commit = m.Commit
//	}
//	p.exec()
//}

func (p *Tendermint) exec() {
		log.Debugf("<--------------------exec()------------------>")
		for {
			log.Debugf("p.execute %v", p.execute)
			e, ok := p.log[p.execute]
			if !ok{
				return
			}
			if !ok || !e.commit {
				log.Debugf("BREAK")
				break
			}
			value := p.Execute(e.command)

			if e.request != nil && e.active {
				reply := PaxiBFT.Reply{
					Command:    e.request.Command,
					Value:      value,
					Properties: make(map[string]string),
				}
				reply.Properties[HTTPHeaderSlot] = strconv.Itoa(p.execute)
				reply.Properties[HTTPHeaderBallot] = e.Ballot.String()
				reply.Properties[HTTPHeaderExecute] = strconv.Itoa(p.execute)
				//time.Sleep(50*time.Millisecond)
				e.request.Reply(reply)
				log.Debugf("********* Reply Primary *********")
				e.request = nil
				e.active = false
				p.view.Reset(p.ID())
			}

			if p.Leader == false {
				d, d1 := p.logR[p.execute]
				if !d1{
					log.Debugf("d is nil")
					break
				}
				if e.request != nil && !e.active {
					log.Debugf("********* Replica Request ********* ")
					p.mux.Lock()
					reply := PaxiBFT.Reply{
						Command:    d.request.Command,
						Value:      value,
						Properties: make(map[string]string),
					}
					p.mux.Unlock()
					reply.Properties[HTTPHeaderSlot] = strconv.Itoa(p.execute)
					reply.Properties[HTTPHeaderBallot] = e.Ballot.String()
					reply.Properties[HTTPHeaderExecute] = strconv.Itoa(p.execute)
					//time.Sleep(50*time.Millisecond)
					d.request.Reply(reply)
					d.request = nil
					log.Debugf("********* Reply Replicas *********")
				}
			}
			// TODO clean up the log periodically
			delete(p.log, p.execute)
			delete(p.logR, p.execute)
			p.execute++
		}
}