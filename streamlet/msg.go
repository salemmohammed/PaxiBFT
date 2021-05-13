package streamlet

import (
	"encoding/gob"
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
)
func init() {

	gob.Register(Propose{})
	gob.Register(Vote{})

}
type Propose struct {
	Ballot 		PaxiBFT.Ballot
	ID     		PaxiBFT.ID
	Request 	PaxiBFT.Request
	Slot 		int
}
func (m Propose) String() string {
	return fmt.Sprintf("Propose {Ballot %v,Command %v, Slot %v}", m.Ballot, m.Request.Command, m.Slot)
}

type Vote struct {
	Ballot 	PaxiBFT.Ballot
	ID     	PaxiBFT.ID
	Slot 	int
	Digest []byte
}
func (m Vote) String() string {
	return fmt.Sprintf("Vote {Ballot %v, ID %v,Slot %v, Digest %v,}", m.Ballot, m.ID,m.Slot,m.Digest)
}