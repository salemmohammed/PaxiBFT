package main

import (
	//"encoding/binary"
	"flag"
	"github.com/salemmohammed/PaxiBFT/log"
	"github.com/salemmohammed/PaxiBFT/paxos"

	"github.com/salemmohammed/PaxiBFT"
)

var id = flag.String("id", "", "node id this client connects to")
var algorithm = flag.String("algorithm", "", "Client API type [paxos]")
var load = flag.Bool("load", false, "Load K keys into DB")
var master = flag.String("master", "", "Master address.")
var delta = flag.Int("delta", 0, "value of delta.")


type db struct {
	PaxiBFT.Client
}

func (d *db) Init() error {
	return nil
}

func (d *db) Stop() error {
	return nil
}


func (d *db) Write(k int, v []byte) error {
	key := PaxiBFT.Key(k)
	//value := make([]byte, binary.MaxVarintLen64)
	//value := make([]byte, 10000)
	//binary.ByteOrder(v)
	//binary.PutUvarint(value, uint64(v))
	err := d.PutMUL(key, v)
	//err := d.Put(key, value)
	return err
}

func main() {
	PaxiBFT.Init()

	if *master != "" {
		PaxiBFT.ConnectToMaster(*master, true, PaxiBFT.ID(*id))
	}

	d := new(db)
	switch *algorithm {
	case "tendermint":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "tendermintBFT":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "tendStar":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "hotstuff":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "HotStuff_SL":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "hotstuffBFT":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "pbft":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "pbftBFT":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "streamlet":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "streamletBFT":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "paxos":
		d.Client = paxos.NewClient(PaxiBFT.ID(*id))
	default:
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	}

	b := PaxiBFT.NewBenchmark(d)
	if *load {
		log.Debugf("Load in Clinet is started")
		b.Load()
	} else {
		log.Debugf("Run in Clinet is started")
		b.Run()
	}
}
