package main

import (
	"flag"
	"github.com/salemmohammed/PaxiBFT/HotStuff"
	"github.com/salemmohammed/PaxiBFT/HotStuffBFT"
	"github.com/salemmohammed/PaxiBFT/HotStuff_SL"
	"github.com/salemmohammed/PaxiBFT/paxos"
	"github.com/salemmohammed/PaxiBFT/pbftBFT"
	"github.com/salemmohammed/PaxiBFT/streamletBFT"
	"github.com/salemmohammed/PaxiBFT/tendermint"
	"github.com/salemmohammed/PaxiBFT/tendermintBFT"
	"sync"

	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"github.com/salemmohammed/PaxiBFT/pbft"
	"github.com/salemmohammed/PaxiBFT/streamlet"
	"github.com/salemmohammed/PaxiBFT/tendStar"
)

var algorithm = flag.String("algorithm", "", "Distributed algorithm")
var id = flag.String("id", "", "ID in format of Zone.Node.")
var simulation = flag.Bool("sim", false, "simulation mode")
var master = flag.String("master", "", "Master address.")

func replica(id PaxiBFT.ID) {
	if *master != "" {
		PaxiBFT.ConnectToMaster(*master, false, id)
	}

	log.Infof("node %v starting...", id)

	switch *algorithm {

	case "tendStar":
		tendStar.NewReplica(id).Run()
	case "pbft":
		pbft.NewReplica(id).Run()
	case "pbftBFT":
		pbftBFT.NewReplica(id).Run()
	case "streamlet":
		streamlet.NewReplica(id).Run()
	case "streamletBFT":
		streamletBFT.NewReplica(id).Run()
	case "tendermint":
		tendermint.NewReplica(id).Run()
	case "tendermintBFT":
		tendermintBFT.NewReplica(id).Run()
	case "hotstuff":
		HotStuff.NewReplica(id).Run()
	case "HotStuff_SL":
		HotStuff_SL.NewReplica(id).Run()
	case "hotstuffBFT":
		HotStuffBFT.NewReplica(id).Run()
	case "paxos":
		paxos.NewReplica(id).Run()

	default:
		panic("Unknown algorithm")
	}
}
func main() {
	PaxiBFT.Init()

	if *simulation {
		var wg sync.WaitGroup
		wg.Add(1)
		PaxiBFT.Simulation()
		for id := range PaxiBFT.GetConfig().Addrs {
			n := id
			go replica(n)
		}
		wg.Wait()
	} else {
		replica(PaxiBFT.ID(*id))
	}
}
