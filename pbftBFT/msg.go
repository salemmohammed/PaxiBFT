package pbftBFT

import (
	"encoding/gob"
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
)

func init() {
	gob.Register(PrePrepare{})
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
}

func (m PrePrepare) String() string {
	return fmt.Sprintf("PrePrepare {Ballot=%v , slot=%v, Request=%v}", m.Ballot,m.Slot,m.Request)
}

// Prepare message
type Prepare struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	View   	PaxiBFT.View
	Slot   	int
	Digest 	[]byte
	Command PaxiBFT.Command
	Request PaxiBFT.Request
}

func (m Prepare) String() string {
	return fmt.Sprintf("Prepare {Ballot=%v, ID=%v, View=%v, slot=%v, command=%v}", m.Ballot,m.ID,m.View,m.Slot,m.Command)
}

// Commit  message
type Commit struct {
	Ballot   PaxiBFT.Ballot
	ID    	 PaxiBFT.ID
	View 	 PaxiBFT.View
	Slot     int
	Digest 	 []byte
	Command  PaxiBFT.Command
	Request  PaxiBFT.Request
}

func (m Commit) String() string {
	return fmt.Sprintf("Commit {Ballot=%v, ID=%v, View=%v, Slot=%v, command=%v}", m.Ballot,m.ID,m.View,m.Slot, m.Command)
}

// ViewChange  message
type ViewChange struct {
	ID    	 PaxiBFT.ID
	Slot     int
	Request  PaxiBFT.Request
}

func (m ViewChange) String() string {
	return fmt.Sprintf("ViewChange {p.ID=%v, Slot=%v, Request=%v}", m.ID, m.Slot, m.Request)
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