package pbftBFT
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
)
// log's entries
type entry struct {
	ballot    PaxiBFT.Ballot
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


type entryBFT struct {
	commit    bool
	request   *PaxiBFT.Request
	timestamp time.Time
	Digest    []byte
}

type Pbftbft struct {
	PaxiBFT.Node
	config          []PaxiBFT.ID
	N               PaxiBFT.Config
	log             map[int]*entry       // log ordered by slot
	BFTlog          map[int]*entryBFT
	slot            int                  // highest slot number
	view            PaxiBFT.View            // view number
	ballot          PaxiBFT.Ballot          // highest ballot number
	execute         int                  // next execute slot number
	requests        []*PaxiBFT.Request
	quorum          *PaxiBFT.Quorum // phase 1 quorum
	RecivedReq      bool
}
func NewPbftBFT(n PaxiBFT.Node, options ...func(*Pbftbft)) *Pbftbft {
	p := &Pbftbft{
		Node:            n,
		log:             make(map[int]*entry, PaxiBFT.GetConfig().BufferSize),
		BFTlog:          make(map[int]*entryBFT, PaxiBFT.GetConfig().BufferSize),
		quorum:          PaxiBFT.NewQuorum(),
		slot:            -1,
		requests:        make([]*PaxiBFT.Request, 0),
		RecivedReq:      false,
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
func (p *Pbftbft) HandleRequest(r PaxiBFT.Request, s int) {
	log.Debugf("<--------------------HandleRequest------------------>")
	e := p.log[s]
	e.Digest = GetMD5Hash(&r)
	log.Debugf("[p.ballot.ID %v, p.ballot %v ]", p.ballot.ID(), p.ballot)
	log.Debugf("PrePrepare will be called")
	p.PrePrepare(&r, &e.Digest, s)
}
func (p *Pbftbft) PrePrepare(r *PaxiBFT.Request, s *[]byte, slt int) {
	log.Debugf("<--------------------PrePrepare------------------>")
	p.Broadcast(PrePrepare{
		Ballot:     p.ballot,
		ID:         p.ID(),
		Slot:       slt,
		Request:    *r,
		Digest:     *s,
	})
	log.Debugf("++++++ PrePrepare Done ++++++")
}

func (p *Pbftbft) HandlePre(m PrePrepare) {
	log.Debugf("<--------------------HandlePre------------------>")
	log.Debugf(" Sender  %v ", m.ID)
	log.Debugf(" m.Slot  %v ", m.Slot)
	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(1))
	// non leader node suspcious the leader
	Digest := GetMD5Hash(&m.Request)

	p.BFTlog[m.Slot] = &entryBFT{
		commit: false,
		request: &m.Request,
		timestamp: time.Now(),
		Digest: Digest,
	}

	if Node_ID != p.ID() {
		log.Debugf("Sart View Change Message ")
		p.Broadcast(ViewChange{
			ID: 	 p.ID(),
			Slot:    m.Slot,
			Digest:  Digest,
		})
	}
}
func (p *Pbftbft) HandleViewChange(m ViewChange) {
	log.Debugf("<--------------------HandleViewChange------------------>")

		e, ok := p.log[m.Slot]
		if !ok {
			log.Debugf("Create a log")
			p.log[m.Slot] = &entry{
				command:   m.Request.Command,
				commit:    false,
				active:    false,
				Leader:    false,
				request:   &m.Request,
				timestamp: time.Now(),
				Digest:    GetMD5Hash(&m.Request),
				Q1:        PaxiBFT.NewQuorum(), // view change
				Q2:        PaxiBFT.NewQuorum(),
				Q3:        PaxiBFT.NewQuorum(),
				Q4:        PaxiBFT.NewQuorum(),
			}
		}
		e = p.log[m.Slot]
		e.Q1.ACK(m.ID)
		if e.Q1.Majority(){
			New_Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(2))
			log.Debugf("Sart New Change Message ")
			p.Broadcast(NewChange{
				ID: 	 New_Node_ID,
				Slot:    p.slot,
				Request: m.Request,
			})
		}
}
func (p *Pbftbft) HandleNewChange(m NewChange) {
	log.Debugf("<--------------------HandleNewChange------------------>")
	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create a log")
		p.log[m.Slot] = &entry{
			command:   m.Request.Command,
			commit:    false,
			active:    false,
			Leader:    false,
			request:   &m.Request,
			timestamp: time.Now(),
			Digest:    GetMD5Hash(&m.Request),
			Q1:        PaxiBFT.NewQuorum(), // view change
			Q2:        PaxiBFT.NewQuorum(), // new change
			Q3:        PaxiBFT.NewQuorum(),
			Q4:        PaxiBFT.NewQuorum(),
		}
	}
	e = p.log[m.Slot]
	e.Q2.ACK(m.ID)
	e.Digest = GetMD5Hash(&m.Request)
	if e.Q2.Majority(){
		New_Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(2))
		if New_Node_ID == p.ID(){
			p.Broadcast(SecondPrePrepare{
				ID:         p.ID(),
				Slot:       m.Slot,
				Digest:    e.Digest,
			})
		}
	}
}
func (p *Pbftbft) HandlePreAfterChange(m SecondPrePrepare) {
	log.Debugf("<--------------------HandlePre------------------>")
	log.Debugf(" Sender  %v ", m.ID)
	log.Debugf(" m.Slot  %v ", m.Slot)
	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(1))
	// non leader node suspcious the leader
	//e := p.log[m.Slot]

	if Node_ID != p.ID() {
		log.Debugf("Sart View Change Message ")
		p.Broadcast(Prepare{
			ID: 	 p.ID(),
			Slot:    m.Slot,
			Digest:  m.Digest,
		})
	}
}

func (p *Pbftbft) HandlePrepare(m Prepare) {
	log.Debugf("<--------------------HandlePrepare------------------>")
	log.Debugf(" Sender  %v ", m.ID)
	log.Debugf("p.slot=%v", p.slot)
	log.Debugf("m.slot=%v", m.Slot)

	e, ok := p.log[m.Slot]

	if !ok {
		log.Debugf("we create a log")
		p.log[m.Slot] = &entry{
			ballot:    p.ballot,
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
func (p *Pbftbft) HandleCommit(m Commit) {
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

func (p *Pbftbft) exec() {
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
