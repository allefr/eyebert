package eyeBERT

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

type Serial interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Flush() error
	Close() error
}

type BERTester struct {
	Manufacturer string
	Model        string
	Version      string
}

type BERStats struct {
	Datetime  string
	Duration  time.Duration
	BER       float64
	BitCnt    float64
	ErrCnt    float64
	RateGpbs  float32
	Pattern   string
	SFPrxPow  float32
	SFPtxPow  float32
	SFPtemp   float32
	EyeHorzUI float32
	EyeVertmV float32
	Status    string
}

type SFPData struct {
	Vendor       string
	PartNum      string
	SerialNum    string
	WaveLengthNm float32
	DistanceKm   float32
	MinRateGbps  float32
	MaxRateGbps  float32
	Temperature  float32
	RxPow        float32
	TxPow        float32
}

type BERTDriver interface {
	GetTesterInfo() (BERTester, error)
	GetSFPinfo() (SFPData, error)
	GetStats() (BERStats, error)
	ResetStats() error
	StartTest() error
	StopTest() error
	SetWaveLength(float32) error
	SetDataRate(float32) error
	SetPattern(Pattern) error
	SetSFPtxEnable(bool) error
	Close() error
}

type BERTParams struct {
	SerPort  string
	DataRate float32
	Pattern  Pattern
}

type device struct {
	conn Serial
	m    sync.RWMutex
}

// supported patterns
type Pattern string

const (
	PattLowPow Pattern = "LOW_POW"
	PattPRBS7  Pattern = "PRBS7"    // PRBS 2^7-1
	PattPRBS9  Pattern = "PRBS9"    // PRBS 2^9-1
	PattPRBS11 Pattern = "PRBS11"   // PRBS 2^11-1
	PattPRBS15 Pattern = "PRBS15"   // PRBS 2^15-1
	PattPRBS23 Pattern = "PRBS23"   // PRBS 2^23-1
	PattPRBS31 Pattern = "PRBS31"   // PRBS 2^31-1
	PattPRBS58 Pattern = "PRBS58"   // PRBS 2^58-1
	PattPRBS63 Pattern = "PRBS63"   // PRBS 2^63-1
	PattLoopBk Pattern = "LOOPBACK" // data on the input is retransmitted on the output
)

var patterns = map[Pattern]string{
	PattLowPow: "0",
	PattPRBS7:  "7",
	PattPRBS9:  "9",
	PattPRBS11: "1",
	PattPRBS15: "5",
	PattPRBS23: "2",
	PattPRBS31: "3",
	PattPRBS58: "8",
	PattPRBS63: "6",
	PattLoopBk: "L",
}

var (
	prevBitCount float64
	timeStart    time.Time
	stopTime     bool
	stopDuration time.Duration
)

func New(p *BERTParams) (d BERTDriver, err error) {
	driver := &device{}

	// cannot continue with no serial port
	if p.SerPort == "" {
		err = ErrSerPort
		return
	}

	c := &serial.Config{
		Name:        p.SerPort,
		Baud:        115200, // doesn't matter, as long as it's supported by the OS
		ReadTimeout: time.Second * 1,
	}
	driver.conn, err = serial.OpenPort(c)
	if err != nil {
		return
	}

	prevBitCount = -1.
	timeStart = time.Now()
	err = driver.StopTest()
	if err != nil {
		return
	}

	if p.DataRate != 0. {
		err = driver.SetDataRate(p.DataRate)
	}
	if err != nil {
		return
	}

	if p.Pattern != "" {
		err = driver.SetPattern(p.Pattern)
	}
	if err != nil {
		return
	}

	err = driver.ResetStats()
	if err != nil {
		return
	}

	d = driver
	return
}

func (d *device) write(s string) error {
	// d.conn.Flush()
	// do not use "=", it crashes the tester
	s = strings.ReplaceAll(s, "=", " ")
	_, err := d.conn.Write([]byte(s + "\r\n"))

	return err
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (d *device) read(l int) (b []byte, err error) {
	var n_ int

	// cannot read more than 1020 bytes (on MacOs at least)
	for n := l; n > 0; n -= 1020 {
		b_ := make([]byte, minInt(n, 1020))

		n_, err = d.conn.Read(b_)
		if err != nil && err.Error() != "EOF" {
			return
		}
		// if no read, tester is most likely stuck
		if n_ == 0 {
			err = ErrTesterStuck
			// log.Fatalf("%v", TesterStuck)
			return
		}

		b = append(b, b_[:n_]...)
	}

	return
}

func (d *device) writeRead(cmd string, l int, wait time.Duration) (b []byte, err error) {
	d.m.Lock()
	defer d.m.Unlock()

	if err = d.write(cmd); err != nil {
		return
	}

	if l == 0 {
		// wait a bit afer sendig a command only
		time.Sleep(100 * time.Millisecond)
		return
	}

	time.Sleep(wait)

	b, err = d.read(l)

	return
}

func (d *device) GetTesterInfo() (r BERTester, err error) {
	var b []byte
	b, err = d.writeRead("?", 80, 10*time.Millisecond)
	if err != nil {
		return
	}

	strn := strings.Fields(string(b))
	if len(strn) >= 3 {
		r.Manufacturer = "Spectronix"
		r.Model = strn[0] + " " + strn[1]
		r.Version = strings.ReplaceAll(strn[2], ":", "")
	} else {
		err = ErrReply
	}

	return
}

func (d *device) GetSFPinfo() (s SFPData, err error) {
	var b []byte
	b, err = d.writeRead("sfp", 2500, 250*time.Millisecond)
	if err != nil {
		return
	}

	// this has multiple lines
	strn := strings.Split(string(b), "\n")
	if len(strn) >= 12 {
		var f float64
		for _, ln := range strn {
			if strings.Contains(ln, "SFP Vendor") {
				s.Vendor = strings.Fields(ln)[2]
			} else if strings.Contains(ln, "Part Number") {
				s.PartNum = strings.Fields(ln)[2]
			} else if strings.Contains(ln, "SN") {
				s.SerialNum = strings.Fields(ln)[1]
			} else if strings.Contains(ln, "Wavelength") {
				f, err = strconv.ParseFloat(strings.Fields(ln)[1], 32)
				s.WaveLengthNm = float32(f)
			} else if strings.Contains(ln, "Temperature") {
				f, err = strconv.ParseFloat(strings.Fields(ln)[1], 32)
				s.Temperature = float32(f)
			} else if strings.Contains(ln, "Rx Power") {
				f, err = strconv.ParseFloat(strings.Fields(ln)[2], 32)
				s.RxPow = float32(f)
			} else if strings.Contains(ln, "Tx Power") {
				f, err = strconv.ParseFloat(strings.Fields(ln)[2], 32)
				s.TxPow = float32(f)
				// no more useful lines after this, can exit
				break
			} else if strings.Contains(ln, "Media") {
				f, err = strconv.ParseFloat(strings.Fields(ln)[2], 32)
				s.DistanceKm = float32(f)
			} else if strings.Contains(ln, "Bit Rate") {
				spl_ := strings.Fields(ln)
				f, _ = strconv.ParseFloat(spl_[2], 32)
				s.MinRateGbps = float32(f)
				f, err = strconv.ParseFloat(spl_[4], 32)
				s.MaxRateGbps = float32(f)
			}
		}
	} else if strings.Contains(strn[0], "No SFP Inserted") {
		err = ErrNoSFP
	} else {
		err = ErrReply
	}

	return
}

func byteExp(b byte, exp float64) float64 {
	return float64(b) * math.Exp2(exp)
}

func checkPattern(b byte) string {
	for key, val := range patterns {
		if val == string(b) {
			return string(key)
		}
	}
	return "unknown ('" + string(b) + "')"
}

func (d *device) GetStats() (bS BERStats, err error) {
	var b []byte
	b, err = d.writeRead("r", 40, 10*time.Millisecond)
	if err != nil {
		return
	}

	if len(b) >= 27 {
		if b[26] != 0x00 {
			err = fmt.Errorf("unexpected termination (0x%2.2x != 0x00)", b[26])
			return
		}

		bS.Datetime = time.Now().UTC().Format(time.RFC3339Nano)
		if stopTime {
			bS.Duration = stopDuration
		} else {
			bS.Duration = time.Since(timeStart)
		}

		bS.RateGpbs = float32(byteExp(b[0], 24)+byteExp(b[1], 16)+byteExp(b[2], 8)+float64(b[3])) / 1e8
		bS.Pattern = checkPattern(b[4])

		bS.SFPrxPow = (32768. - (float32(b[5])*256. + float32(b[6]))) / 100.
		bS.SFPtxPow = (32768. - (float32(b[7])*256. + float32(b[8]))) / 100.
		// waveLen := float32(byteExp(b[9], 16)+byteExp(b[10], 8)+float64(b[11])) / 100.
		bS.SFPtemp = (32768. - (float32(b[12])*256. + float32(b[13]))) / 100.

		currSfpStatus := b[14]
		switch currSfpStatus & 0x3F {
		case 0:
			bS.Status = "SFP not in use"
		case 1:
			bS.Status = "SFP no signal"
		case 2:
			bS.Status = "ok" // "SFP signal and synch OK"
		case 3:
			bS.Status = "SFP signal but no lock"
		default:
			err = fmt.Errorf("SFP wtf status (0x%2.2X)", currSfpStatus)
			return
		}
		if currSfpStatus&0x40 == 0 {
			bS.Status = "no SFP detected"
		}
		if currSfpStatus&0x80 == 0x80 {
			bS.Status = "new SFP inserted"
		}

		bS.BitCnt = (byteExp(b[15], 16) + byteExp(b[16], 8) + float64(b[17])) * math.Exp2(float64(b[18])-24.)
		bS.ErrCnt = (byteExp(b[19], 16) + byteExp(b[20], 8) + float64(b[21])) * math.Exp2(float64(b[22])-24.)
		if (!stopTime && bS.BitCnt <= prevBitCount) || bS.BitCnt == 0 {
			bS.BER = -1.
		} else {
			bS.BER = bS.ErrCnt / bS.BitCnt
		}
		prevBitCount = bS.BitCnt

		bS.EyeHorzUI = float32(b[23]) / 32.
		bS.EyeVertmV = float32(b[24]) * 3.125
	} else {
		err = ErrReply
	}

	return
}

func (d *device) ResetStats() (err error) {
	prevBitCount = -1.
	timeStart = time.Now()
	stopDuration = 0
	_, err = d.writeRead("reset", 0, 0)

	return
}

func (d *device) SetWaveLength(wl float32) (err error) {
	_, err = d.writeRead("setwl "+fmt.Sprintf("%f", wl), 0, 0)

	return
}

func (d *device) SetDataRate(dr float32) (err error) {
	_, err = d.writeRead("setrate "+fmt.Sprintf("%d", int(dr*1e6)), 0, 0) // kbps
	if err != nil {
		return
	}
	// this will effectively restart the test, so let's make it obvious
	return d.ResetStats()
}

func (d *device) SetPattern(p Pattern) (err error) {
	if strV, ok := patterns[p]; ok {
		_, err = d.writeRead("setpat "+strV, 0, 0)
	} else {
		err = fmt.Errorf("invalid pattern: " + string(p))
	}
	if err != nil {
		return
	}
	// this will effectively restart the test, so let's make it obvious
	return d.ResetStats()
}

func (d *device) SetSFPtxEnable(on bool) (err error) {
	cmdV := "0"
	if on {
		cmdV = "1"
	}
	_, err = d.writeRead("tx "+cmdV, 0, 0)

	return
}

func (d *device) StartTest() (err error) {
	// no actual command do so, but there is a way around it
	stopTime = false
	if err := d.SetSFPtxEnable(true); err != nil {
		return err
	}
	err = d.ResetStats()

	return
}

func (d *device) StopTest() (err error) {
	// no actual command do so, but there is a way around it
	stopTime = true
	stopDuration = time.Since(timeStart)
	err = d.SetSFPtxEnable(false)

	return
}

func (d *device) Close() error {
	return d.conn.Close()
}

var (
	ErrSerPort     = errors.New("no serial port provided")
	ErrTesterStuck = errors.New("tester seems stuck! please power cycle it")
	ErrNoSFP       = errors.New("no SFP detected")
	ErrReply       = errors.New("un-expected reply from tester")
)

// for validation only
var _ Serial = &serial.Port{}
var _ BERTDriver = &device{}
