package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/luislopes-nf/Heartbleed/heartbleed"
)

var usageMessage = `This is a tool for detecting OpenSSL Heartbleed vulnerability (CVE-2014-0160).

Usage:  %s [flags] server_name[:port]

The default port is 443 (HTTPS).
If a URL is supplied in server_name, it will be parsed to extract the host, but not the protocol.

The following flags are recognized:
`

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage, os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	var (
		service = flag.String("service", "https", fmt.Sprintf(
			`Specify a service name to test (using STARTTLS if necessary).
		Besides HTTPS, currently supported services are:
		%s`, heartbleed.Services))
		check_cert = flag.Bool("check-cert", false, "check the server certificate")
		good_request = flag.Bool("good-request", false, "Send a good heartbeat request instead")
		add_padding = flag.Bool("add-padding", false, "Add padding to heartbeat request")
		no_payload = flag.Bool("no-payload", false, "Don't add any payload to heartbeat request")
	)
	var payload = []byte("birdsandbees")
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
	}

	tgt := &heartbleed.Target{
		Service: *service,
		HostIp:  flag.Arg(0),
	}

	// Parse the host out of URLs
	u, err := url.Parse(tgt.HostIp)
	if err == nil && u.Host != "" {
		tgt.HostIp = u.Host
		if u.Scheme != "" {
			tgt.Service = u.Scheme
		}
	}
	if *no_payload {
		payload = []byte("")
	}
	out, err := heartbleed.Heartbleed(tgt,
		payload, !(*check_cert),
		*good_request, *add_padding)
	if err == heartbleed.Safe {
		log.Printf("%v - SAFE", tgt.HostIp)
		os.Exit(0)
	} else if err != nil {
		if err.Error() == "Please try again" {
			log.Printf("%v - TRYAGAIN: %v", tgt.HostIp, err)
			os.Exit(2)
		} else {
			log.Printf("%v - ERROR: %v", tgt.HostIp, err)
			os.Exit(2)
		}
	} else {
		log.Printf("%v\n", out)
		log.Printf("%v - VULNERABLE", tgt.HostIp)
		os.Exit(1)
	}
}
