package pbft

import (
	"encoding/gob"
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
)

func init() {
	gob.Register(PrePrepare{})
	gob.Register(Prepare{})
	gob.Register(Commit{})
}

// <PrePrepare,seq,v,s,d(m),m>
type PrePrepare struct {
	Ballot 		PaxiBFT.Ballot
	ID     		PaxiBFT.ID
	View 		PaxiBFT.View
	Slot    	int
	Request 	PaxiBFT.Request
	Digest 		[]byte
	ActiveView	bool
	Command 	PaxiBFT.Command
}

func (m PrePrepare) String() string {
	return fmt.Sprintf("PrePrepare {Ballot=%v , View=%v, slot=%v, Command=%v}", m.Ballot,m.View,m.Slot,m.Command)
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