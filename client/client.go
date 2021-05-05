package main

import (
	"encoding/binary"
	"flag"
	"github.com/salemmohammed/PaxiBFT/log"

	"github.com/salemmohammed/PaxiBFT"
)

var id = flag.String("id", "", "node id this client connects to")
var algorithm = flag.String("algorithm", "", "Client API type [paxos]")
var load = flag.Bool("load", false, "Load K keys into DB")
var master = flag.String("master", "", "Master address.")
var delta = flag.Int("delta", 1, "value of delta.")


type db struct {
	PaxiBFT.Client
}

func (d *db) Init() error {
	return nil
}

func (d *db) Stop() error {
	return nil
}


func (d *db) Read(k int) (value []int) {
	//log.Debugf("Read")
	key := PaxiBFT.Key(k)
	v, _ := d.GetMUL(key)
	//log.Debugf("after the Read in client")
	if len(v) == 0 {
		//log.Debugf("len(v) %v", len(v))
		return nil
	}
	var m []int
	for _,v := range v{
		x, _ := binary.Uvarint(v)
		//log.Debugf("x %v", x)
		m = append(m,int(x))
	}
	//log.Debugf("m %v", m)
	return m
}

func (d *db) Write(k, v int) []error {
	key := PaxiBFT.Key(k)
	value := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(value, uint64(v))
	err := d.Put(key, value)
	return err
}

func main() {
	PaxiBFT.Init()

	if *master != "" {
		PaxiBFT.ConnectToMaster(*master, true, PaxiBFT.ID(*id))
	}

	d := new(db)
	switch *algorithm {
	case "paxos":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "tendermint":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
	case "tendStar":
		d.Client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))
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
