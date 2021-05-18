package streamletBFT

import (
	"crypto/md5"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"strconv"
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
	active 		  bool
	Leader 		  bool
	slot          int
	Pstatus    status
	Cstatus    status
	Rstatus	   status
}

type StreamletBFT struct {
	PaxiBFT.Node
	log      					map[int]*entry 				// log ordered by slot
	config 						[]PaxiBFT.ID
	execute 					int             			// next execute slot number
	ballot  					PaxiBFT.Ballot     			// highest ballot number
	slot    					int             			// highest slot number
	Requests 					[]*PaxiBFT.Request 			// phase 1 pending request
	MyRequests					*PaxiBFT.Request
	Leader						bool
	Delta 						int
}
func NewStreamletBFT(n PaxiBFT.Node, options ...func(*StreamletBFT)) *StreamletBFT {
	p := &StreamletBFT{
		Node:          	 	n,
		log:           	 	make(map[int]*entry, PaxiBFT.GetConfig().BufferSize),
		slot:          	 	-1,
		Requests:      	 	make([]*PaxiBFT.Request, 0),
		Leader:				false,
		Delta:				0,
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

func (p *StreamletBFT) HandleRequest(r PaxiBFT.Request) {
	log.Debugf("\n<---R----HandleRequest----R------>\n")

	p.Broadcast(Propose{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot: 		p.slot,
		Request:    r,
	})
}
func (p *StreamletBFT) handlePropose(m Propose) {
	log.Debugf("<--------------------handlePropose----------------->\n")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create the log")
		p.log[m.Slot] = &entry{
			Ballot:    m.Ballot,
			request:   &m.Request,
			Timestamp: time.Now(),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			active:    false,
			Leader:    false,
			commit:    false,
		}
	}
	e = p.log[m.Slot]
	e.Pstatus = PREPARED

	w := (m.Slot+1)%e.Q1.Total() + 1
	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(w))

	log.Debugf("Node_ID = %v", Node_ID)
	log.Debugf("m.Request = %v", m.Request)
	if p.ID() != Node_ID {
		p.Send(Node_ID, ViewChange{
			Ballot:  m.Ballot,
			ID:      p.ID(),
			Slot:    m.Slot,
			Request: m.Request,
		})
	}
}
func (p *StreamletBFT) HandleViewChange(m ViewChange) {
	log.Debugf("<--------------------HandleViewChange----------------->")
	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create the log")
		p.log[m.Slot] = &entry{
			Ballot:    m.Ballot,
			request:   &m.Request,
			Timestamp: time.Now(),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			active:    false,
			Leader:    false,
			commit:    false,
		}
	}
	e = p.log[m.Slot]
	e.Pstatus = PREPARED
	e.Q2.ACK(m.ID)
	log.Debugf("majority = %v" , e.Q2.Size())
	if e.Q2.Majority(){
		e.Q2.Reset()
		p.Broadcast(ProposeAfterFailure{
			Ballot:  m.Ballot,
			ID:      p.ID(),
			Slot:    m.Slot,
			Request: m.Request,
		})
	}
	if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED{
		log.Debug("late call")
		e.commit = true
		p.exec()
	}
}
func (p *StreamletBFT) HandleProposeAfterFailure(m ProposeAfterFailure) {
	log.Debugf("<--------------------HandleProposeAfterFailure----------------->")
	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create the log")
		p.log[m.Slot] = &entry{
			Ballot:    m.Ballot,
			request:   &m.Request,
			Timestamp: time.Now(),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			active:    false,
			Leader:    false,
			commit:    false,
		}
	}
	e = p.log[m.Slot]
	e.Pstatus = PREPARED
	time.Sleep(500 * time.Millisecond)
	p.Broadcast(Vote{
			Ballot:  m.Ballot,
			ID:      p.ID(),
			Slot:    m.Slot,
			Digest:  GetMD5Hash(&m.Request),
	})
	if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED{
		log.Debug("late call")
		e.commit = true
		p.exec()
	}
}

func (p *StreamletBFT) HandleVote(m Vote) {
	log.Debugf("<--------------------HandleVote------------------>\n")
	log.Debugf("m.slot %v", m.Slot)
	log.Debugf("sender %v", m.ID)

	_, ok := p.log[m.Slot]
	if !ok {
		if m.Slot < p.execute{
			log.Debugf("old")
			return
		}
		log.Debugf("Create the log")
		p.log[m.Slot] = &entry{
			Ballot:    	m.Ballot,
			request:   	&PaxiBFT.Request{
				Command:    PaxiBFT.Command{},
				Properties: nil,
				Timestamp:  0,
				NodeID:     "",
			},
			Timestamp: 	time.Now(),
			Q1:			PaxiBFT.NewQuorum(),
			Q2: 		PaxiBFT.NewQuorum(),
			Q3: 		PaxiBFT.NewQuorum(),
			active:     false,
			Leader:		false,
			commit:    	false,
		}
	}
	e := p.log[m.Slot]
	e.Q3.ACK(m.ID)
	if e.Q3.Majority(){
		log.Debugf("majority")
		log.Debugf("e.Pstatus = %v", e.Pstatus)
		log.Debugf("e.RECEIVED = %v", e.Rstatus)
		e.Cstatus = COMMITTED
	}
	if (e.Q3.Majority() || e.Cstatus == COMMITTED )&& e.Pstatus == PREPARED  && e.Rstatus == RECEIVED{
		e.commit = true
		e.Q3.Reset()
		p.exec()
	}
}
func (p *StreamletBFT) exec() {
		log.Debugf("<--------------------exec()------------------>")
		for {
			log.Debugf("p.execute %v", p.execute)
			e, ok := p.log[p.execute]
			if !ok || !e.commit {
				log.Debugf("Break")
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