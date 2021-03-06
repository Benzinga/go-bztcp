package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Benzinga/go-bztcp/bztcp"
)

const defaultAddr = "tcp-v1.benzinga.io:11337"

func main() {
	var (
		addr    string
		user    string
		key     string
		tls     bool
		verbose bool
	)

	startTime := time.Now()

	// Set up logging.
	log.SetFlags(log.Lshortfile)

	// Create context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse flags.
	flag.StringVar(&addr, "addr", defaultAddr, "address of bztcp server")
	flag.StringVar(&user, "user", "", "username to authenticate with")
	flag.StringVar(&key, "key", "", "key to authenticate with")
	flag.BoolVar(&tls, "tls", false, "whether or not to use TLS")
	flag.BoolVar(&verbose, "v", false, "enable verbose logging")
	flag.Parse()

	if verbose {
		log.Println("Benzinga TCP Client initializing.")
	}

	dialer := bztcp.Dial

	if tls {
		dialer = bztcp.DialTLS
	}

	// Dial server.
	if verbose {
		log.Printf("Connecting to '%s' as user '%s' (w/TLS: %t)\n", addr, user, tls)
	}

	conn, err := dialer(addr, user, key)
	if err != nil {
		log.Fatalln(err)
	}

	// Listen for signals.
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch

		if verbose {
			log.Println("Received signal. Exiting...")
		}

		cancel()
	}()

	// Start streaming.
	if verbose {
		log.Printf("Connected. Waiting for events.")
	}

	enc := json.NewEncoder(os.Stdout)

	err = conn.Stream(ctx, func(stream bztcp.StreamData) {
		err = enc.Encode(stream)
		if err != nil {
			log.Fatalln(err)
		}
	})
	if err != nil {
		log.Fatalln(err)
	}

	if verbose {
		log.Println("Finished. Runtime: ", time.Since(startTime), "Started At: ", startTime)
	}
}
