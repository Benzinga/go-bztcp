package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"

	bztcp "github.com/Benzinga/go-bztcp"
)

const defaultAddr = "tcp-v1.benzinga.com:11337"

func main() {
	var addr, user, key string

	// Set up logging.
	log.SetFlags(log.Lshortfile)

	// Create context.
	context, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse flags.
	flag.StringVar(&addr, "addr", defaultAddr, "address of bztcp server")
	flag.StringVar(&user, "user", "", "username to authenticate with")
	flag.StringVar(&key, "key", "", "key to authenticate with")
	flag.Parse()

	// Dial server.
	conn, err := bztcp.Dial(addr, user, key)
	if err != nil {
		log.Fatalln(err)
	}

	// Listen for signals.
	go func() {
		ch := make(chan os.Signal)
		signal.Notify(ch, os.Interrupt)
		<-ch
		cancel()
	}()

	// Start streaming.
	enc := json.NewEncoder(os.Stdout)
	err = conn.Stream(context, func(stream bztcp.StreamData) {
		err = enc.Encode(stream)
		if err != nil {
			log.Fatalln(err)
		}
	})
	if err != nil {
		log.Fatalln(err)
	}
}
