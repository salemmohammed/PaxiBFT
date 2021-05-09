package pbft

import (
	"crypto/md5"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"time"
)

type status int8

const (
	NONE status = iota
	PREPREPARED
	PREPARED
	COMMITTED
	RECEIVED
)

// log's entries
type entry struct {
	ballot    PaxiBFT.Ballot
	view      PaxiBFT.View
	command   PaxiBFT.Command
	commit    bool
	active    bool
	Leader    bool
	request   *PaxiBFT.Request
	timestamp time.Time
	Digest    []byte
	Q1        *PaxiBFT.Quorum
	Q2        *PaxiBFT.Quorum
	Q3        *PaxiBFT.Quorum
	Q4        *PaxiBFT.Quorum
	Pstatus    status
	Cstatus    status
	Rstatus	   status
}

// pbft instance
type Pbft struct {
	PaxiBFT.Node

	config          []PaxiBFT.ID
	N               PaxiBFT.Config
	log             map[int]*entry       // log ordered by slot


	slot            int                  // highest slot number
	view            PaxiBFT.View            // view number
	ballot          PaxiBFT.Ballot          // highest ballot number
	execute         int                  // next execute slot number
	requests        []*PaxiBFT.Request
	quorum          *PaxiBFT.Quorum // phase 1 quorum
	ReplyWhenCommit bool
	RecivedReq      bool
	Member          *PaxiBFT.Memberlist
}

// NewPbft creates new pbft instance
func NewPbft(n PaxiBFT.Node, options ...func(*Pbft)) *Pbft {
	p := &Pbft{
		Node:            n,
		log:             make(map[int]*entry, PaxiBFT.GetConfig().BufferSize),

		quorum:          PaxiBFT.NewQuorum(),
		slot:            -1,

		requests:        make([]*PaxiBFT.Request, 0),
		ReplyWhenCommit: false,
		RecivedReq:      false,
		Member:          PaxiBFT.NewMember(),
	}
	for _, opt := range options {
		opt(p)
	}
	return p
}

// Digest message
func GetMD5Hash(r *PaxiBFT.Request) []byte {
	hasher := md5.New()
	hasher.Write([]byte(r.Command.Value))
	return []byte(hasher.Sum(nil))
}

func (p *Pbft) HandleRequest(r PaxiBFT.Request, s int) {
	log.Debugf("<--------------------HandleRequest------------------>")

	e := p.log[s]
	e.Digest = GetMD5Hash(&r)
	log.Debugf("[p.ballot.ID %v, p.ballot %v ]", p.ballot.ID(), p.ballot)
	log.Debugf("PrePrepare will be called")
	p.PrePrepare(&r, &e.Digest, s)
}

// Pre_prepare starts phase 1 PrePrepare
// the primary will send <<pre-prepare,v,n,d(m)>,m>
func (p *Pbft) PrePrepare(r *PaxiBFT.Request, s *[]byte, slt int) {
	log.Debugf("<--------------------PrePrepare------------------>")

	_, ok := p.log[p.slot]
	if !ok {
		p.log[p.slot] = &entry{
			ballot:    p.ballot,
			view:      p.view,
			command:   r.Command,
			commit:    false,
			active:    false,
			Leader:    false,
			request:   r,
			timestamp: time.Now(),
			Digest:    GetMD5Hash(r),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			Q4:        PaxiBFT.NewQuorum(),
		}
	}

	p.Broadcast(PrePrepare{
		Ballot:     p.ballot,
		ID:         p.ID(),
		View:       p.view,
		Slot:       slt,
		Request:    *r,
		Digest:     *s,
		Command:    r.Command,
	})
	log.Debugf("++++++ PrePrepare Done ++++++")
}

// HandleP1a handles Pre_prepare message
func (p *Pbft) HandlePre(m PrePrepare) {
	log.Debugf("<--------------------HandlePre------------------>")

	log.Debugf(" Sender  %v ", m.ID)

	log.Debugf(" m.Slot  %v ", m.Slot)

	if m.Ballot > p.ballot {
		log.Debugf("m.Ballot > p.ballot")
		p.ballot = m.Ballot
		p.view = m.View
	}

	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create a log")
		p.log[m.Slot] = &entry{
			ballot:    p.ballot,
			view:      p.view,
			command:   m.Command,
			commit:    false,
			active:    false,
			Leader:    false,
			request:   &m.Request,
			timestamp: time.Now(),
			Digest:    GetMD5Hash(&m.Request),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			Q4:        PaxiBFT.NewQuorum(),
		}
	}
	e, ok = p.log[m.Slot]

	e.Digest = GetMD5Hash(&m.Request)
	for i, v := range e.Digest {
		if v != m.Digest[i] {
			return
		}
	}
	log.Debugf("m.Ballot=%v , p.ballot=%v, m.view=%v", m.Ballot, p.ballot, m.View)
	log.Debugf("at the prepare handling")
	p.Broadcast(Prepare{
		Ballot:  p.ballot,
		ID:      p.ID(),
		View:    m.View,
		Slot:    m.Slot,
		Digest:  m.Digest,
	})
	log.Debugf("++++++ HandlePre Done ++++++")
}

// HandlePrepare starts phase 2 HandlePrepare
func (p *Pbft) HandlePrepare(m Prepare) {
	log.Debugf("<--------------------HandlePrepare------------------>")
	log.Debugf(" Sender  %v ", m.ID)
	log.Debugf("p.slot=%v", p.slot)
	log.Debugf("m.slot=%v", m.Slot)

	e, ok := p.log[m.Slot]

	if !ok {
		log.Debugf("we create a log")
		p.log[m.Slot] = &entry{
			ballot:    p.ballot,
			view:      p.view,
			command:   m.Command,
			commit:    false,
			active:    false,
			Leader:    false,
			request:   &m.Request,
			timestamp: time.Now(),
			Digest:    GetMD5Hash(&m.Request),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			Q4:        PaxiBFT.NewQuorum(),
		}
	}
	e, ok = p.log[m.Slot]
	e.Q1.ACK(m.ID)

	if e.Q1.Majority(){
		e.Q1.Reset()
		e.Pstatus = PREPARED
		p.Broadcast(Commit{
			Ballot:  p.ballot,
			ID:      p.ID(),
			View:    p.view,
			Slot:    m.Slot,
			Digest:  m.Digest,
		})
	}
	if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED{
		e.commit = true
		p.exec()
	}
	log.Debugf("++++++ HandlePrepare Done ++++++")
}

// HandleCommit starts phase 3
func (p *Pbft) HandleCommit(m Commit) {
	log.Debugf("<--------------------HandleCommit------------------>")
	log.Debugf(" Sender  %v ", m.ID)
	log.Debugf("m.slot=%v", m.Slot)
	log.Debugf("p.slot=%v", p.slot)
	if p.execute > m.Slot{
		log.Debugf("old message")
		return
	}
	e, exist := p.log[m.Slot]
	if !exist {
		log.Debugf("create a log")
		p.log[m.Slot] = &entry{
			ballot:    p.ballot,
			view:      p.view,
			command:   m.Command,
			commit:    false,
			active:    false,
			Leader:    false,
			request:   &m.Request,
			timestamp: time.Now(),
			Digest:    GetMD5Hash(&m.Request),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			Q4:        PaxiBFT.NewQuorum(),
		}
	}
	e, exist = p.log[m.Slot]
	e.Q2.ACK(m.ID)


	log.Debugf("Q2 size =%v", e.Q2.Size())
	if e.Q2.Majority(){
		e.Cstatus = COMMITTED
	}
	if (e.Q2.Majority() || e.Cstatus == COMMITTED )&& e.Pstatus == PREPARED  && e.Rstatus == RECEIVED{
		e.Q2.Reset()
		e.commit = true
		p.exec()
	}
	log.Debugf("********* Commit End *********** ")
}

func (p *Pbft) exec() {
	log.Debugf("<--------------------exec()------------------>")
	for {
		log.Debugf("p.execute %v", p.execute)
		e, ok := p.log[p.execute]
		if !ok || !e.commit {
			log.Debugf("Break")
			break
		}
		value := p.Execute(e.command)
		log.Debugf("value=%v", value)

		reply := PaxiBFT.Reply{
			Command:    e.command,
			Value:      value,
			Properties: make(map[string]string),
		}

		if e.request != nil && e.Leader{
			log.Debugf(" ********* Primary Request ********* %v", *e.request)
			e.request.Reply(reply)
			log.Debugf("********* Reply Primary *********")
			e.request = nil
		}else{
			log.Debugf("********* Replica Request ********* ")
			log.Debugf("p.ID() =%v", p.ID())
			e.request.Reply(reply)
			e.request = nil
			log.Debugf("********* Reply Replicas *********")
		}
		// TODO clean up the log periodically
		delete(p.log, p.execute)
		p.execute++
	}
}
