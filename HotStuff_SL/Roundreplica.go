package HotStuff_SL

import (
	"github.com/salemmohammed/PaxiBFT"
)

var List []PaxiBFT.ID
var round int = 0
var i int

func (p *Replica) NextReplica(lst []PaxiBFT.ID,size int) (PaxiBFT.ID) {

	id := lst[round]
	round++
	round = round % size
	return id
}

//func (p *Replica) handleRound(s int, id PaxiBFT.ID) {
//	log.Debugf("-------------handleRound------------")
//	log.Debugf("current p.requests = %v", p.Requests)
//	log.Debugf("Id = %v ", id)
//	log.Debugf("s = %v ", s)
//	p.Broadcast1(RoundRobin{ID: id,slot: s})
//}