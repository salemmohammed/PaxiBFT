package tendermint

import (
	"encoding/gob"
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
)
func init() {
	gob.Register(Propose{})
	gob.Register(PreVote{})
	gob.Register(PreCommit{})
}
// Propose
type Propose struct {
	Ballot 		PaxiBFT.Ballot
	ID     		PaxiBFT.ID
	Request 	PaxiBFT.Request
	Slot 		int
	Command 	PaxiBFT.Command
	View		PaxiBFT.View
	ID_LIST_PR  PaxiBFT.Quorum
	Active      bool
}
func (m Propose) String() string {
	return fmt.Sprintf("Propose {Ballot %v,Command %v, Slot %v, ID_LIST_PR.AID %v, Active %v}", m.Ballot, m.Command, m.Slot, m.ID_LIST_PR.AID, m.Active)
}
// PreVote
type PreVote struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	Slot 	int
	Command PaxiBFT.Command
	View	PaxiBFT.View
	ID_LIST_PV  PaxiBFT.Quorum
	Active      bool
}
func (m PreVote) String() string {
	return fmt.Sprintf("PreVote {Ballot %v,ID %v,Command %v, Slot %v, View %v, ID_LIST_PV.AID %v}", m.Ballot, m.ID, m.Command, m.Slot, m.View, m.ID_LIST_PV.AID)
}
// PreCommit  message
type PreCommit struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	PreVote PaxiBFT.Quorum
	Command PaxiBFT.Command
	Slot 	int
	Commit  bool
	ID_LIST_PC  PaxiBFT.Quorum
	Active      bool
}
func (m PreCommit) String() string {
	return fmt.Sprintf("PreCommit {Ballot %v, Command %v, Commit %v, Slot %v, ID_LIST_PC.AID %v}", m.Ballot,  m.Command, m.Commit,m.Slot, m.ID_LIST_PC.AID)
}