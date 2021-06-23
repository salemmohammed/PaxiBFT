package tendStar

import (
	"encoding/gob"
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
)
func init() {

	gob.Register(Propose{})
	gob.Register(PreVote{})
	gob.Register(PreCommit{})
	gob.Register(ActPropose{})
	gob.Register(ActPreCommit{})
	gob.Register(ActPreVote{})
	gob.Register(RoundRobin{})

}
// Propose
type Propose struct {
	Ballot 		PaxiBFT.Ballot
	ID     		PaxiBFT.ID
	Request 	PaxiBFT.Request
	Slot 		int
	View		PaxiBFT.View
}
func (m Propose) String() string {
	return fmt.Sprintf("Propose {Ballot %v,request %v, Slot %v}", m.Ballot, m.Request, m.Slot)
}
// PreVote
type PreVote struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	Slot 	int
	View	PaxiBFT.View
}
func (m PreVote) String() string {
	return fmt.Sprintf("PreVote {Ballot %v,ID %v,Request %v, Slot %v}", m.Ballot, m.ID, m.Request, m.Slot)
}
// PreCommit  message
type PreCommit struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	Slot 	int
	Commit  bool
}
func (m PreCommit) String() string {
	return fmt.Sprintf("PreCommit {Ballot %v, Command %v, Request %v, Slot %v}", m.Ballot,  m.Request, m.Commit,m.Slot)
}
// ActPropose  message
type ActPropose struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	Slot 	int
}
func (m ActPropose) String() string {
	return fmt.Sprintf("ActPropose {Ballot %v, Reqeust %v, Slot %v}", m.Ballot,  m.Request,m.Slot)
}
// ActPropose  message
type ActPreCommit struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	Slot 	int
}
func (m ActPreCommit) String() string {
	return fmt.Sprintf("ActPreCommit {Ballot %v, Request %v,Slot %v}", m.Ballot,  m.Request,m.Slot)
}
// ActPropose  message
type ActPreVote struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	Slot 	int
}
func (m ActPreVote) String() string {
	return fmt.Sprintf("ActPreVote {Ballot %v, Request %v, Slot %v}", m.Ballot,  m.Request,m.Slot)
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