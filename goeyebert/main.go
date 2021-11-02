package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/allefr/eyebert/goeyebert/eyebert"
)

func main() {
	// read command line arguments
	port, patt, rate, freq, useJson := getClArgs()

	p := &eyebert.BERTParams{
		SerPort:  port,
		DataRate: rate,
		Pattern:  eyebert.Pattern(patt),
	}
	tester, err := eyebert.New(p)
	if err != nil {
		log.Fatalf("cannot init tester: %v\n\n", err)
	}
	defer tester.Close()

	// tester info
	tInfo, err := tester.GetTesterInfo()
	if err != nil {
		log.Fatalf("error: %v\n\n", err)
	}
	_ = print(tInfo, useJson)

	// sfp info
	sInfo, err := tester.GetSFPinfo()
	if err != nil {
		log.Fatalf("error: %v\n\n", err)
	}
	_ = print(sInfo, useJson)

	// test start!
	if err := tester.StartTest(); err != nil {
		log.Fatalf("error start test: %v\n\n", err)
	}

	var stats eyebert.BERStats
	for {
		stats, err = tester.GetStats()
		if err != nil {
			if err != eyebert.ErrTesterStuck {
				fmt.Printf("error: %v\n\n", err)
			} else {
				log.Fatalln(err)
				break
			}
		}
		_ = print(stats, useJson)

		time.Sleep(time.Duration(1000./freq) * time.Millisecond)
	}
}

func print(i interface{}, asJson bool) error {
	if asJson {
		b, err := json.Marshal(i)
		fmt.Println(string(b))
		return err
	}

	// need to check depending on interface
	switch t := i.(type) {
	case eyebert.BERTester:
		d := i.(eyebert.BERTester)
		fmt.Printf("BERT Model: %s - Version %s\n", d.Model, d.Version)
	case eyebert.SFPData:
		d := i.(eyebert.SFPData)
		fmt.Printf("SFP: %s (%s, %s) - %.1fkm %.1fnm %.1fdegC Rx %.2fdBm Tx %.2fdBm\n",
			d.SerialNum, d.Vendor, d.PartNum, d.DistanceKm, d.WaveLengthNm,
			d.Temperature, d.RxPow, d.TxPow)
	case eyebert.BERStats:
		d := i.(eyebert.BERStats)
		fmt.Printf("%12s\tBER: %13e (errCnt: %12e) eye: %4.2fUI %6.2fmV %22s\r",
			d.Duration, d.BER, d.ErrCnt, d.EyeHorzUI, d.EyeVertmV, d.Status)
	default:
		_ = t
		fmt.Println(i)
	}
	return nil
}

func getClArgs() (port, patt string, rate, freq float32, json bool) {
	// custom flag.Usage function similar to python argparse
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-h] ", os.Args[0])

		// optional flags first
		flag.VisitAll(func(f *flag.Flag) {
			n, _ := flag.UnquoteUsage(f)
			s := "[-" + f.Name
			if len(n) > 0 {
				s += " <" + n + ">"
			}
			s += "]"
			fmt.Fprintf(os.Stderr, "%s ", s)
		})
		// positional flag(s) after, cause flag.Parse() will ignore
		// any remaining flag if it can't parse the current one
		fmt.Fprintf(os.Stderr, "port\n\n")

		fmt.Fprintf(os.Stderr, "positional arguments:\n  port\tserial port\n\n")
		fmt.Fprintf(os.Stderr, "optional arguments:\n")
		flag.PrintDefaults()
	}

	freqPtr := flag.Float64("f", 1., "polling frequency [Hz]")
	rateGbpsPtr := flag.Float64("r", 0., "datarate [Gbps] (default taken from sfp)")
	flag.StringVar(&patt, "p", "PRBS7", "bert pattern as \"PRBS<7|9|11|15|23|31|58|63>\"")
	flag.BoolVar(&json, "j", false, "output as json")
	flag.Parse()

	freq = float32(*freqPtr)
	rate = float32(*rateGbpsPtr)

	// look for required positional argument
	posArgs := flag.Args()
	if len(posArgs) < 1 {
		fmt.Fprintf(os.Stderr, "serial port required\n\n")
		flag.Usage()
		os.Exit(-1)
	}
	port = posArgs[0]

	return
}
