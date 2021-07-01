package HotStuff_SL

import (
	"encoding/gob"
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
)
func init() {
	gob.Register(Prepare{})
	gob.Register(ActPrepare{})

	gob.Register(PreCommit{})
	gob.Register(ActPreCommit{})

	gob.Register(Commit{})
	gob.Register(ActCommit{})

	gob.Register(Decide{})
	gob.Register(ActDecide{})

	gob.Register(RoundRobin{})

}
type Prepare struct {
	Ballot 		PaxiBFT.Ballot
	ID     		PaxiBFT.ID
	Request 	PaxiBFT.Request
	Slot 		int
}
func (m Prepare) String() string {
	return fmt.Sprintf("Prepare {Ballot %v,Request %v, Slot %v, ID %v}", m.Ballot, m.Request, m.Slot, m.ID)
}
type ActPrepare struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Digest  []byte
	Slot 	int
	Request 	PaxiBFT.Request
}
func (m ActPrepare) String() string {
	return fmt.Sprintf("ActPrepare {Ballot %v, Digest %v, Slot %v , Request %v}", m.Ballot,  m.Digest,m.Slot, m.Request)
}
type PreCommit struct {
	Ballot 		PaxiBFT.Ballot
	ID     		PaxiBFT.ID
	Digest 	    []byte
	Slot 		int
	Request 	PaxiBFT.Request
}
func (m PreCommit) String() string {
	return fmt.Sprintf("PreCommit {Ballot %v,Digest %v, Slot %v , Request %v}", m.Ballot, m.Digest, m.Slot,m.Request)
}
type ActPreCommit struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Digest  []byte
	Slot 	int
	Request 	PaxiBFT.Request
}
func (m ActPreCommit) String() string {
	return fmt.Sprintf("ActPreCommit {Ballot %v, Digest %v, Slot %v}", m.Ballot,  m.Digest,m.Slot)
}
type Commit struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Digest  []byte
	Slot 	int
	Request 	PaxiBFT.Request
}
func (m Commit) String() string {
	return fmt.Sprintf("Commit {Ballot %v, Digest %v, Slot %v}", m.Ballot,  m.Digest, m.Slot)
}
type ActCommit struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Digest  []byte
	Slot 	int
	Request 	PaxiBFT.Request
}
func (m ActCommit) String() string {
	return fmt.Sprintf("ActCommit {Ballot %v, Digest %v,Slot %v}", m.Ballot,  m.Digest,m.Slot)
}
type Decide struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Digest []byte
	Slot 	int
	Request 	PaxiBFT.Request
}
func (m Decide) String() string {
	return fmt.Sprintf("Decide {Ballot %v, Digest %v, Slot %v}", m.Ballot,m.Digest,m.Slot)
}
type ActDecide struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Digest  []byte
	Slot 	int
	Request 	PaxiBFT.Request
}
func (m ActDecide) String() string {
	return fmt.Sprintf("ActDecide {Ballot %v, Digest %v,Slot %v}", m.Ballot,m.Digest,m.Slot)
}

type RoundRobin struct {
	Slot     		int
	//Id              PaxiBFT.ID
	Request         PaxiBFT.Request
	Id              PaxiBFT.ID
}
func (m RoundRobin) String() string {
	return fmt.Sprintf("RoundRobin {slot %v Request %v Id %v}", m.Slot, m.Request,m.Id)
}