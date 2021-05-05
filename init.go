package PaxiBFT

import (
	"flag"
	"net/http"

	"github.com/salemmohammed/PaxiBFT/log"
)

// Init setup paxi package
func Init() {
	flag.Parse()
	log.Setup()
	config.Load()
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 1000
}
