package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
	"os"
	"strconv"
	"strings"
)

var id = flag.String("id", "", "node id this client connects to")
var algorithm = flag.String("algorithm", "", "Client API type [paxos]")
var master = flag.String("master", "", "Master address.")
var delta = flag.Int("delta", 0, "value of delta.")



func usage() string {
	s := "Usage:\n"
	s += "\t get key\n"
	s += "\t put key value\n"
	s += "\t consensus key\n"
	s += "\t crash id time\n"
	s += "\t partition time ids...\n"
	s += "\t exit\n"
	return s
}

var client PaxiBFT.Client
var admin PaxiBFT.AdminClient

func run(cmd string, args []string) {
	switch cmd {
	case "put":
		if len(args) < 2 {
			fmt.Println("put KEY VALUE")
			return
		}
	case "consensus":
		if len(args) < 1 {
			fmt.Println("consensus KEY")
			return
		}
		k, _ := strconv.Atoi(args[0])
		v := admin.Consensus(PaxiBFT.Key(k))
		fmt.Println(v)

	case "crash":
		if len(args) < 2 {
			fmt.Println("crash id time(s)")
			return
		}
		id := PaxiBFT.ID(args[0])
		time, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("second argument should be integer")
			return
		}
		admin.Crash(id, time)

	case "partition":
		if len(args) < 2 {
			fmt.Println("partition time ids...")
			return
		}
		time, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("time argument should be integer")
			return
		}
		ids := make([]PaxiBFT.ID, 0)
		for _, s := range args[1:] {
			ids = append(ids, PaxiBFT.ID(s))
		}
		admin.Partition(time, ids...)

	case "exit":
		os.Exit(0)

	case "help":
		fallthrough
	default:
		fmt.Println(usage())
	}
}

func main() {
	PaxiBFT.Init()

	if *master != "" {
		PaxiBFT.ConnectToMaster(*master, true, PaxiBFT.ID(*id))
	}

	admin = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))

	client = PaxiBFT.NewHTTPClient(PaxiBFT.ID(*id))

	if len(flag.Args()) > 0 {
		run(flag.Args()[0], flag.Args()[1:])
	} else {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("PaxiBFT $ ")
			text, _ := reader.ReadString('\n')
			words := strings.Fields(text)
			if len(words) < 1 {
				continue
			}
			cmd := words[0]
			args := words[1:]
			run(cmd, args)
		}
	}
}
