package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/salemmohammed/PaxiBFT"
)

var file = flag.String("log", "log.csv", "")

func main() {
	flag.Parse()

	h := PaxiBFT.NewHistory()

	err := h.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}

	n := h.Linearizable()

	fmt.Println(n)
}
