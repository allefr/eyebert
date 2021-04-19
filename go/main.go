package main

import (
	"fmt"
	"log"
	"time"

	"github.com/allefr/spectronix-eyeBERT/go/eyeBERT"
)

func main() {

	p := &eyeBERT.BERTParams{
		SerPort: "/dev/tty.usbmodem142101", // example on MacOs
		// DataRate: 10.3125,
		// Pattern:  eyeBERT.PattPRBS31,
	}
	tester, err := eyeBERT.New(p)
	if err != nil {
		log.Fatalf("cannot init tester: %v\n\n", err)
	}

	// tester info
	tInfo, err := tester.GetTesterInfo()
	if err != nil {
		log.Fatalf("error: %v\n\n", err)
	}
	fmt.Println(tInfo)

	// sfp info
	sInfo, err := tester.GetSFPinfo()
	if err != nil {
		log.Fatalf("error: %v\n\n", err)
	}
	fmt.Println(sInfo)

	// stats before starting the test
	stats, err := tester.GetStats()
	if err != nil {
		fmt.Printf("error: %v\n\n", err)
	}
	fmt.Println(stats)

	// test start!
	if err := tester.StartTest(); err != nil {
		log.Fatalf("error start test: %v\n\n", err)
	}

	for {
		stats, err = tester.GetStats()
		if err != nil {
			if err != eyeBERT.ErrTesterStuck {
				fmt.Printf("error: %v\n\n", err)
			} else {
				log.Fatalln(err)
				break
			}
		}
		fmt.Println(stats)
		time.Sleep(200 * time.Millisecond)
	}

	tester.Close()

}
