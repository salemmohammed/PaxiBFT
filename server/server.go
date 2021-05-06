package main

import (
	"flag"
	"github.com/salemmohammed/PaxiBFT/tendermint"
	"sync"

	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"github.com/salemmohammed/PaxiBFT/paxos"
	"github.com/salemmohammed/PaxiBFT/pbft"
	"github.com/salemmohammed/PaxiBFT/tendStar"
	"github.com/salemmohammed/PaxiBFT/streamlet"
)

var algorithm = flag.String("algorithm", "tendStar", "Distributed algorithm")
var id = flag.String("id", "", "ID in format of Zone.Node.")
var simulation = flag.Bool("sim", false, "simulation mode")
var delta = flag.Int("delta", 1, "value of delta.")

var master = flag.String("master", "", "Master address.")

func replica(id PaxiBFT.ID) {
	if *master != "" {
		PaxiBFT.ConnectToMaster(*master, false, id)
	}

	log.Infof("node %v starting...", id)

	switch *algorithm {

	case "paxos":
		paxos.NewReplica(id).Run()

	case "tendStar":
		tendStar.NewReplica(id).Run()

	case "pbft":
		pbft.NewReplica(id).Run()

	case "streamlet":
		streamlet.NewReplica(id,*delta).Run()

	case "tendermint":
		tendermint.NewReplica(id).Run()

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