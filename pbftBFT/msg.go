package pbftBFT

import (
	"encoding/gob"
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
)

func init() {
	gob.Register(PrePrepare{})
	gob.Register(SecondPrePrepare{})
	gob.Register(ViewChange{})
	gob.Register(NewChange{})
	gob.Register(Prepare{})
	gob.Register(Commit{})
}

// <PrePrepare,seq,v,s,d(m),m>
type PrePrepare struct {
	Ballot 		PaxiBFT.Ballot
	ID     		PaxiBFT.ID
	Slot    	int
	Request 	PaxiBFT.Request
	Digest 		[]byte
	Node_ID     PaxiBFT.ID
}

func (m PrePrepare) String() string {
	return fmt.Sprintf("PrePrepare {Ballot=%v , slot=%v, Request=%v}", m.Ballot,m.Slot,m.Request)
}

// Prepare message
type Prepare struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Slot   	int
	Digest 	[]byte
}

func (m Prepare) String() string {
	return fmt.Sprintf("Prepare {Ballot=%v, ID=%v, slot=%v, Digest=%v}", m.Ballot,m.ID,m.Slot,m.Digest)
}

// Commit  message
type Commit struct {
	Ballot   PaxiBFT.Ballot
	ID    	 PaxiBFT.ID
	Slot     int
	Digest 	 []byte
}

func (m Commit) String() string {
	return fmt.Sprintf("Commit {Ballot=%v, ID=%v, Slot=%v, Digest}", m.Ballot,m.ID,m.Slot,m.Digest)
}

// ViewChange  message
type ViewChange struct {
	ID    	 PaxiBFT.ID
	Slot     int
	Request  PaxiBFT.Request
}

func (m ViewChange) String() string {
	return fmt.Sprintf("ViewChange {p.ID=%v, Slot=%v, Request=%v}", m.ID, m.Slot,m.Request.Command)
}

// ViewChange  message
type NewChange struct {
	ID    	 PaxiBFT.ID
	Slot     int
	Request  PaxiBFT.Request
}

func (m NewChange) String() string {
	return fmt.Sprintf("NewChange {p.ID=%v, Slot=%v, Request=%v}", m.ID, m.Slot, m.Request)
}

type SecondPrePrepare struct {
	ID     		PaxiBFT.ID
	Slot    	int
	Digest 		[]byte
}

func (m SecondPrePrepare) String() string {
	return fmt.Sprintf("SecondPrePrepare {ID=%v , slot=%v, Digest=%v}", m.ID,m.Slot,m.Digest)
}