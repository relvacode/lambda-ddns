package main

import (
	"context"
	"flag"
	"github.com/relvacode/lambda-ddns/ddns"
	"log"
)

var (
	flagHostname = flag.String("hostname", "", "Target DDNS hostname")
	flagRegion   = flag.String("region", "", "AWS Region")
)

func main() {
	flag.Parse()

	securityGroupIds := flag.Args()
	if len(securityGroupIds) == 0 {
		log.Fatalln("No security groups provided")
	}

	if *flagHostname == "" {
		log.Fatalln("No target hostname provided")
	}
	if *flagRegion == "" {
		log.Fatalln("No AWS region provided")
	}

	handler, err := ddns.New(securityGroupIds, *flagRegion, *flagHostname)
	if err != nil {
		log.Fatalln(err)
	}

	err = handler.Update(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
}
