package main

import (
	"fmt"
	"log"
	"time"

	"github.com/allefr/spectronix-eyeBERT/go/eyeBERT"
)

func main() {
	fmt.Println(eyeBERT.PattPRBS31)

	p := &eyeBERT.BERTParams{
		SerPort:  "/dev/tty.usbmodem142101",
		DataRate: 10.3125,
		Pattern:  eyeBERT.PattPRBS31,
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
			fmt.Printf("error: %v\n\n", err)
		}
		fmt.Println(stats)
		time.Sleep(200 * time.Millisecond)
	}

	tester.Close()

}
