package HotStuff

import (
	"encoding/gob"
	"fmt"
	"github.com/ailidani/paxi"
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

}
type Prepare struct {
	Ballot 		paxi.Ballot
	ID     		paxi.ID
	Request 	paxi.Request
	Slot 		int
	Command 	paxi.Command
	View		paxi.View
	Leader      bool
}
func (m Prepare) String() string {
	return fmt.Sprintf("Prepare {Ballot %v,Command %v, Slot %v, Leader %v}", m.Ballot, m.Command, m.Slot, m.Leader)
}
type ActPrepare struct {
	Ballot 	paxi.Ballot
	ID     	paxi.ID
	Command paxi.Command
	Slot 	int
}
func (m ActPrepare) String() string {
	return fmt.Sprintf("ActPrepare {Ballot %v, Command %v, Slot %v}", m.Ballot,  m.Command,m.Slot)
}
type PreCommit struct {
	Ballot 		paxi.Ballot
	ID     		paxi.ID
	Request 	paxi.Request
	Slot 		int
	Command 	paxi.Command
	View		paxi.View
}
func (m PreCommit) String() string {
	return fmt.Sprintf("PreCommit {Ballot %v,Command %v, Slot %v}", m.Ballot, m.Command, m.Slot)
}
type ActPreCommit struct {
	Ballot 	paxi.Ballot
	ID     	paxi.ID
	Request paxi.Request
	Command paxi.Command
	Slot 	int
}
func (m ActPreCommit) String() string {
	return fmt.Sprintf("ActPreCommit {Ballot %v, Command %v, Slot %v}", m.Ballot,  m.Command,m.Slot)
}
type Commit struct {
	Ballot 	paxi.Ballot
	ID     	paxi.ID
	Request paxi.Request
	Command paxi.Command
	Slot 	int
	Commit  bool
}
func (m Commit) String() string {
	return fmt.Sprintf("Commit {Ballot %v, Command %v, Commit %v, Slot %v}", m.Ballot,  m.Command, m.Commit,m.Slot)
}
type ActCommit struct {
	Ballot 	paxi.Ballot
	ID     	paxi.ID
	Request paxi.Request
	Command paxi.Command
	Slot 	int
}
func (m ActCommit) String() string {
	return fmt.Sprintf("ActCommit {Ballot %v, Command %v,Slot %v}", m.Ballot,  m.Command,m.Slot)
}
type Decide struct {
	Ballot 	paxi.Ballot
	ID     	paxi.ID
	Request paxi.Request
	Command paxi.Command
	Slot 	int
}
func (m Decide) String() string {
	return fmt.Sprintf("Decide {Ballot %v, Command %v, Slot %v}", m.Ballot,  m.Command,m.Slot)
}
type ActDecide struct {
	Ballot 	paxi.Ballot
	ID     	paxi.ID
	Request paxi.Request
	Command paxi.Command
	Slot 	int
	Commit  bool
}
func (m ActDecide) String() string {
	return fmt.Sprintf("ActDecide {Ballot %v, Command %v,Slot %v}", m.Ballot,  m.Command,m.Slot)
}