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
	gob.Register(RoundRobin{})
}
type Propose struct {
	Ballot 		PaxiBFT.Ballot
	ID     		PaxiBFT.ID
	Request 	PaxiBFT.Request
	Slot 		int
	ID_LIST_PR  PaxiBFT.Quorum
}
func (m Propose) String() string {
	return fmt.Sprintf("Propose {Ballot %v, Slot %v, ID_LIST_PR.AID %v}", m.Ballot, m.Slot, m.ID_LIST_PR.AID)
}
type PreVote struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	Slot 	int
	ID_LIST_PV  PaxiBFT.Quorum
}
func (m PreVote) String() string {
	return fmt.Sprintf("PreVote {Ballot %v,ID %v,Slot %v,ID_LIST_PV.AID %v}", m.Ballot, m.ID, m.Slot, m.ID_LIST_PV.AID)
}
type PreCommit struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Request PaxiBFT.Request
	Slot 	int
	Commit  bool
	ID_LIST_PC  PaxiBFT.Quorum
}
func (m PreCommit) String() string {
	return fmt.Sprintf("PreCommit {Ballot %v,Commit %v, Slot %v, ID_LIST_PC.AID %v}", m.Ballot, m.Commit,m.Slot, m.ID_LIST_PC.AID)
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