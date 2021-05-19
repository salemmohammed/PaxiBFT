package tendermintBFT
import (
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
	PR	  		  *PaxiBFT.Quorum
	PV    		  *PaxiBFT.Quorum
	PC    		  *PaxiBFT.Quorum
	active 		  bool
	Leader		  bool
	Pstatus       status
	Rstatus       status
	Cstatus       status
	VC            status
	VCs           *PaxiBFT.Quorum
}
type TendermintBFT struct {
	PaxiBFT.Node
	log          map[int]*entry // log ordered by slot
	config       []PaxiBFT.ID
	execute      int                // next execute slot number
	active       bool               // active leader
	ballot       PaxiBFT.Ballot     // highest ballot number
	slot         int                // highest slot number
	quorum       *PaxiBFT.Quorum    // phase 1 quorum
	Requests     []*PaxiBFT.Request // phase 1 pending requests
	Member       *PaxiBFT.Memberlist
	count        int
	Leader       bool
	EarlyPropose bool
	mux          sync.Mutex
	Plist        []PaxiBFT.ID
}
func NewTendermintBFT(n PaxiBFT.Node, options ...func(*TendermintBFT)) *TendermintBFT {
	p := &TendermintBFT{
		Node:          	 	n,
		log:           	 	make(map[int]*entry, PaxiBFT.GetConfig().BufferSize),
		slot:          	 	-1,
		quorum:        	 	PaxiBFT.NewQuorum(),
		Requests:      	 	make([]*PaxiBFT.Request, 0),
		Member:         	PaxiBFT.NewMember(),
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
func (p *TendermintBFT) HandleRequest(r PaxiBFT.Request) {
	log.Debugf("<---R----HandleRequest----R------>\n")

	p.Member.Addmember(r.NodeID)
	log.Debugf("Nighbors %v", p.Member.Neibors)

	p.count = 0
	e := p.log[p.slot]
	e.PR.AID_ID(p.ID())

	for _, v := range p.Member.Neibors {
		p.count++
		e.PR.ACK(v)
		e.PR.ACK(p.ID())
		e.PR.AID_ID(v)

		p.Send(v, Propose{
			Ballot:     p.ballot,
			ID:         p.ID(),
			Request:    r,
			Slot:       p.slot,
			ID_LIST_PR: *e.PR,
		})

		if p.count >= e.PR.Total()/2{
			break
		}
	}
	log.Debugf("e.PR.ID : %v", e.PR.AID)
}
func (p *TendermintBFT) handlePropose(m Propose) {
	log.Debugf("<-------P-------------handlePropose--------P---------->\n")
	log.Debugf("Sender:%v ", m.ID)
	log.Debugf("m.ID_LIST_PR.AID = %v", m.ID_LIST_PR.AID)

	if m.Ballot > p.ballot {
		log.Debugf("m is bigger m.Ballot:%v, p.ballot:%v", m.Ballot, p.ballot)
		p.ballot = m.Ballot
	}

	e, ok := p.log[m.Slot]
	if !ok {
		p.log[m.Slot] = &entry{
			Ballot:    p.ballot,
			commit:    false,
			request:   &m.Request,
			Timestamp: time.Now(),
			PR:        PaxiBFT.NewQuorum(),
			PV:        PaxiBFT.NewQuorum(),
			PC:        PaxiBFT.NewQuorum(),
			VCs:        PaxiBFT.NewQuorum(),
			active:    false,
			Leader:    false,
		}
	}
	e = p.log[m.Slot]

	for _, i1 := range m.ID_LIST_PR.AID {
		flagMatch := false
		for _, i2 := range e.PR.AID {
			if i1 == i2 {
				flagMatch = true
			}
		}
		if flagMatch == false {
			e.PR.AID = append(e.PR.AID, i1)
		}
	}
	for _, i2 := range e.PR.AID {
		log.Debugf("list : %v", i2)
	}
	e.PR.ACK(p.ID())
	p.Member.Addmember(p.ID())
	log.Debugf("list of neighbors %v", p.Member.Neibors)
	p.count = 0
	for i := len(p.Member.Neibors) - 1; i >= 0; i-- {
		found := false
		for _, i2 := range e.PR.AID {
			if p.Member.Neibors[i] == i2 {
				found = true
				break
			}
		}
		if !found {
			p.count++
			e.PR.AID = append(e.PR.AID, p.Member.Neibors[i])
			e.PR.ACK(p.Member.Neibors[i])
			p.Send(p.Member.Neibors[i], Propose{
				Ballot:     p.ballot,
				ID:         p.ID(),
				Request:    m.Request,
				Slot:       m.Slot,
				ID_LIST_PR: *e.PR,
			})

		}
		if p.count >= e.PR.Total()/2 {
			log.Debugf("Only half time sending")
			break
		}
	}

	log.Debugf("---------------------------------------------")
	w := (m.Slot+1)%e.PR.Total() + 1
	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(w))
	log.Debugf("Node_ID = %v", Node_ID)
	log.Debugf("---------------------------------------------")

	if p.ID() != Node_ID && e.VC != NEWCHAMGED{
		e.VC = NEWCHAMGED

		e.VCs.ACK(p.ID())
		e.VCs.AID_ID(p.ID())

		p.Send(Node_ID, ViewChange{
			Ballot:  m.Ballot,
			ID:      p.ID(),
			Slot:    m.Slot,
			Request: m.Request,
			ID_LIST_PR: *e.PR,
			ID_LIST_VC: *e.VCs,
			NID: Node_ID,
		})
	}
}

func (p *TendermintBFT) handleAfterPropose(m ViewChange) {

	log.Debugf("<-------P-------------handleAfterPropose--------P---------->\n")
	log.Debugf("Sender:%v ", m.ID)
	log.Debugf("m.ID_LIST_PR.AID = %v", m.ID_LIST_PR.AID)
	log.Debugf("m.ID_LIST_PR.AID = %v", m.ID_LIST_VC.AID)

	if m.Ballot > p.ballot {
		log.Debugf("m is bigger m.Ballot:%v, p.ballot:%v", m.Ballot, p.ballot)
		p.ballot = m.Ballot
	}

	e, ok := p.log[m.Slot]
	if !ok {
		p.log[m.Slot] = &entry{
			Ballot:    p.ballot,
			commit:    false,
			request:   &m.Request,
			Timestamp: time.Now(),
			PR:        PaxiBFT.NewQuorum(),
			PV:        PaxiBFT.NewQuorum(),
			PC:        PaxiBFT.NewQuorum(),
			VCs:        PaxiBFT.NewQuorum(),
			active:    false,
			Leader:    false,
		}
	}
	e = p.log[m.Slot]
	e.Pstatus = PREPARED


	for _, i1 := range m.ID_LIST_PR.AID {
		flagMatch := false
		for _, i2 := range e.PR.AID {
			if i1 == i2 {
				flagMatch = true
			}
		}
		if flagMatch == false {
			e.PR.AID = append(e.PR.AID, i1)
		}
	}
	for _, i2 := range e.PR.AID {
		log.Debugf("list : %v", i2)
	}
	e.PR.ACK(p.ID())
	p.Member.Addmember(p.ID())
	log.Debugf("list of neighbors %v", p.Member.Neibors)
	p.count = 0
	for i := len(p.Member.Neibors) - 1; i >= 0; i-- {
		found := false
		for _, i2 := range e.PR.AID {
			if p.Member.Neibors[i] == i2 {
				found = true
				break
			}
		}
		if !found {
			p.count++
			e.PR.AID = append(e.PR.AID, p.Member.Neibors[i])
			e.PR.ACK(p.Member.Neibors[i])
			p.Send(p.Member.Neibors[i], Propose{
				Ballot:     p.ballot,
				ID:         p.ID(),
				Request:    m.Request,
				Slot:       m.Slot,
				ID_LIST_PR: *e.PR,
			})

		}
		if p.count >= e.PR.Total()/2 {
			log.Debugf("Only half time sending")
			break
		}
	}
	p.count = 0
	if len(e.PR.AID) > len(p.Member.Neibors) {
		for _, v1 := range p.Member.Neibors {
			value := false
			for _, i3 := range e.PV.AID {
				if i3 == v1 {
					value = true
					break
				}
			}
			if !value {
				p.count++
				e.PV.ACK(v1)
				e.PV.AID = append( e.PV.AID, v1)
				//p.Send(v1, PreVote{
				//	Ballot:     p.ballot,
				//	ID:         p.ID(),
				//	Slot:       m.Slot,
				//	Request:    m.Request,
				//	ID_LIST_PV: *e.PV,
				//})

				p.Send(m.NID, ViewChange{
					Ballot:  m.Ballot,
					ID:      p.ID(),
					Slot:    m.Slot,
					Request: m.Request,
					ID_LIST_PR: *e.PR,
				})
			}
			if p.count >= e.PV.Total()/2 {
				log.Debugf("Only half time sending")
				break
			}
		}
	}
	log.Debugf("e.PR.ID: %v", e.PR.ID )
	log.Debugf("members %v", p.Member.Neibors)
	log.Debugf("e.PV.AID: %v", e.PV.AID )
	log.Debugf("e.PC.AID: %v",e.PC.AID )
	log.Debugf("++++++++++++++++++++++++++ handlePropose Done ++++++++++++++++++++++++++")
}
func (p *TendermintBFT) handleViewChange(m ViewChange) {

}

func (p *TendermintBFT) HandlePreVote(m PreVote) {
	log.Debugf("<---------V-----------HandlePreVote----------V-------->\n")
	log.Debugf("Sender ID %v", m.ID)
	log.Debugf("m.slot %v ", m.Slot)
	log.Debugf("Members  %v ", p.Member.Size())

	log.Debugf("m.ID_LIST_PV.AID = %v", m.ID_LIST_PV.AID)

	if m.Ballot > p.ballot {
		log.Debugf("m.ballot is bigger")
		p.ballot = m.Ballot
	}
	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("We cannot allocate the log b/c prevote b/f request")
		return
	}
	for _, i1 := range m.ID_LIST_PV.AID{
		flagMatch := false
		for _, i2 := range e.PV.AID{
			if i1 == i2{
				flagMatch = true
			}
		}
		if flagMatch == false{
			e.PV.AID = append( e.PV.AID, i1)
		}
	}
	p.count = 0
	p.Member.Addmember(p.ID())
	for _, v1 := range p.Member.Neibors {
		found := false
		for _, v2 := range e.PV.AID {
			if v1 == v2{
				found = true
				break
			}
		}
		if !found {
			p.count++
			e.PV.ACK(m.ID)
			e.PV.ACK(v1)
			e.PV.ACK(p.ID())
			e.PV.AID = append( e.PV.AID, v1)
			p.Send(v1, PreVote{
				Ballot:     p.ballot,
				ID:         p.ID(),
				Slot:       m.Slot,
				Request:    m.Request,
				ID_LIST_PV: *e.PV,
			})
		}
		if p.count >= e.PV.Total()/2 {
			log.Debugf("Only half time sending")
			break
		}
	}
	log.Debugf("e.PC.AID %v", e.PC.AID)
	log.Debugf("e.PV.AID %v", e.PV.AID)

	p.count = 0
	if len(e.PV.AID) > len(p.Member.Neibors) {
		for _, v1 := range p.Member.Neibors {
			value := false
			for _, v3 := range e.PC.AID {
				if v1 == v3 {
					value = true
					break
				}
			}
			if !value {
				p.count++
				e.PC.ACK(v1)
				e.PC.ACK(p.ID())
				e.PC.AID = append( e.PC.AID, v1)
				p.Send(v1, PreCommit{
					Ballot:  p.ballot,
					ID:      p.ID(),
					Slot:    m.Slot,
					Request: m.Request,
					ID_LIST_PC: *e.PC,
					Commit: false,
				})

			}
			if p.count >= e.PC.Total()/2 {
				log.Debugf("Only half time sending")
				break
			}
		}
	}
}

func (p *TendermintBFT) HandlePreCommit(m PreCommit) {
	log.Debugf("<----------C----------HandlePreCommit--------C---------->\n")

	log.Debugf("Sender ID %v", m.ID)
	log.Debugf("m.slot %v ", m.Slot)

	if m.Ballot > p.ballot {
		log.Debugf("m is bigger than p")
		p.ballot = m.Ballot
	}
	log.Debugf("command key  %v ", m.Request.Command.Key)
	log.Debugf("Members  %v ", p.Member.Size())

	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("We cannot allocate the log b/c precommit is old or early")
		log.Debugf("Old MSG")
		return
	}
	//log.Debugf(" p.Member.ClientSize() %v ", p.Member.ClientSize())
	log.Debugf("e.PC.AID %v", e.PC.AID)
	log.Debugf("e.PV.AID %v", e.PV.AID)
	if !e.request.Command.Equal(m.Request.Command) && e.request == nil {
		log.Debugf("Not consistent")
		return
	}
	for _, i1 := range m.ID_LIST_PC.AID{
		flagMatch := false
		for _, i2 := range e.PC.AID{
			if i1 == i2{
				flagMatch = true
			}
		}
		if flagMatch == false{
			e.PC.AID = append( e.PC.AID, i1)
		}
	}
	log.Debugf("e.PV.AID %v", e.PV.AID)
	log.Debugf("e.PC.AID %v", e.PC.AID)
	p.count = 0
	for _, v1 := range p.Member.Neibors {
		found := false
		for _, v2 := range e.PC.AID {
			if v1 == v2{
				found = true
				break
			}
		}
		if !found {
			p.count++
			e.PC.ACK(v1)
			e.PC.AID = append(e.PC.AID, v1)

			p.Send(v1,PreCommit{
				Ballot:     p.ballot,
				ID:         p.ID(),
				Slot:       m.Slot,
				Request:    m.Request,
				ID_LIST_PC: *e.PC,
				Commit:     false,
			})
		}
		if p.count >= e.PC.Total()/2 {
			log.Debugf("Only half time sending")
			break
		}
	}

	log.Debugf("p.log[m.Slot].ID_LIST_PV %v", e.PV.AID)
	log.Debugf("p.log[m.Slot].ID_LIST_PC %v", e.PC.AID)
	e.Cstatus = COMMITTED
	if len(e.PC.AID) > len(p.Member.Neibors) && e.Rstatus == RECEIVED{
		e.commit = true
		e.Ballot = p.ballot
		log.Debugf("e.commit = %v", e.commit)
		p.exec()
	}
}
func (p *TendermintBFT) exec() {
	log.Debugf("<--------------------exec()------------------>")
	for {
		log.Debugf("p.execute %v", p.execute)
		e, ok := p.log[p.execute]
		if !ok || !e.commit {
			log.Debugf("break")
			break
		}
		value := p.Execute(e.request.Command)
		if e.request != nil && e.active && e.Leader {
			reply := PaxiBFT.Reply{
				Command:    e.request.Command,
				Value:      value,
				Properties: make(map[string]string),
			}
			log.Debugf("*********  Primary *********")
			log.Debugf("Reply = %v", e.request)
			e.request.Reply(reply)
			e.request = nil
			log.Debugf("********* Reply Primary *********")
		}
		if e.request != nil {
			log.Debugf("********* Replica ********* ")
			reply := PaxiBFT.Reply{
				Command:    e.request.Command,
				Value:      value,
				Properties: make(map[string]string),
			}
			log.Debugf("Reply = %v", e.request)
			e.request.Reply(reply)
			e.request = nil
			log.Debugf("********* Reply Replicas *********")
		}
		// TODO clean up the log periodically
		delete(p.log, p.execute)
		p.execute++
	}
}