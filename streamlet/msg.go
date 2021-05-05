package streamlet

import (
	"encoding/gob"
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
)
func init() {

	gob.Register(Propose{})
	gob.Register(PreCommit{})
	gob.Register(ActPreCommit{})

}
// Propose
type Propose struct {
	Ballot 		PaxiBFT.Ballot
	ID     		PaxiBFT.ID
	Request 	PaxiBFT.Request
	Slot 		int
	Command 	PaxiBFT.Command
	View		PaxiBFT.View
}
func (m Propose) String() string {
	return fmt.Sprintf("Propose {Ballot %v,Command %v, Slot %v}", m.Ballot, m.Command, m.Slot)
}
// PreCommit  message
type PreCommit struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	Command PaxiBFT.Command
	Slot 	int
	Commit  bool
}
func (m PreCommit) String() string {
	return fmt.Sprintf("PreCommit {Ballot %v, Command %v, Commit %v, Slot %v}", m.Ballot,  m.Command, m.Commit,m.Slot)
}
// ActPropose  message
type ActPreCommit struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	Command PaxiBFT.Command
	Slot 	int
	Commit  bool
}
func (m ActPreCommit) String() string {
	return fmt.Sprintf("ActPreCommit {Ballot %v, Command %v,Slot %v, Commit %v,}", m.Ballot,  m.Command,m.Slot,m.Commit)
}