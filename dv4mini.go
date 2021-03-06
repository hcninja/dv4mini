/*
Copyright 2016 Hacker Cat Ninja

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dv4mini

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

var (
	// Byte 1-4 Preamble; Byte 4 Command; Byte 5 Length of params; Byte 6-n Params
	CmdPreamble = []byte{0x71, 0xfe, 0x39, 0x1d}
	debug       = false
)

// Dongle commands
const (
	SETADFQRG       = 0x01
	SETADFMODE      = 0x02
	FLUSHTXBUF      = 0x03
	ADFWRITE        = 0x04
	ADFWATCHDOG     = 0x05
	ADFGETDATA      = 0x07
	ADFGREENLED     = 0x08
	ADFSETPOWER     = 0x09
	ADFFLASHMODESET = 0x0b // captured from USB with dv4mini official software
	ADFFLASHMODE_0c = 0x0c // captured from USB with dv4mini official software
	ADFFLASHMODE_0d = 0x0d // captured from USB with dv4mini official software
	ADFFLASHMODE_0e = 0x0e // captured from USB with dv4mini official software
	ADFFLASHMODE_0f = 0x0f // captured from USB with dv4mini official software
	ADFDEBUG        = 0x10
	COMMAND_11      = 0x11 // captured from USB with dv4mini official software
	STRFWVERSION    = 0x12 // captured from USB with dv4mini official software
	COMMAND_13      = 0x13 // captured from USB with dv4mini official software
	COMMAND_14      = 0x14 // captured from USB with dv4mini official software
	ADFSETSEED      = 0x17
	ADFVERSION      = 0x18
	ADFSETTXBUF     = 0x19
)

// Dongle mode
const (
	MODE_DSTAR = 0x44
	MODE_C4FM  = 0x46
	MODE_DMR   = 0x4d
	MODE_DPRM  = 0x4d
	MODE_P25   = 0x4d
)

// TX power constants
const (
	_ = uint8(iota)
	POWER_MIN
	POWER_1
	POWER_2
	POWER_3
	POWER_4
	POWER_5
	POWER_6
	POWER_7
	POWER_8
	POWER_MAX
)

type DV4Mini struct {
	Port      *serial.Port
	RSSIMSB   uint8
	RSSILSB   uint8
	RSSI      int8
	FWVersion string
	DongleID  string
	// RXChan       chan bool
	// WatchdogChan chan bool
}

// Connect opens the DV4mini serial device
func Connect(device string, dbg bool) (*DV4Mini, error) {
	var err error
	d := &DV4Mini{}

	debug = dbg

	// Set up options.
	options := serial.Config{
		Name:        device,
		Baud:        115200,
		Parity:      serial.ParityNone,
		Size:        8,
		StopBits:    serial.Stop1,
		ReadTimeout: time.Millisecond * 250,
	}

	// Open the port. io.ReadWriteCloser
	d.Port, err = serial.OpenPort(&options)
	if err != nil {
		log.Println(err.Error())
		return d, fmt.Errorf("Serial device %s not found", options.Name)
	}

	return d, nil
}

// Close closes serial connection
func (d *DV4Mini) Close() {
	d.FlushTXBuffer()
	d.Port.Flush()
	d.Port.Close()
}

// Flush all the data available in the serial buffer
func (d *DV4Mini) FlushSerial() error {
	err := d.Port.Flush()
	if err != nil {
		return err
	}

	return nil
}

// FlushTXBuffer
func (d *DV4Mini) FlushTXBuffer() {
	// []byte{0x03}
	d.sendCmd([]byte{FLUSHTXBUF})
}

// GreenLedOn sets the green LED on
func (d *DV4Mini) GreenLedOn() {
	// []byte{ADFGREENLED, 0x01, 0x01}
	d.sendCmd([]byte{ADFGREENLED, 0x01})
}

// GreenLedOff sets the green LED off
func (d *DV4Mini) GreenLedOff() {
	// []byte{ADFGREENLED, 0x01, 0x00}
	d.sendCmd([]byte{ADFGREENLED, 0x00})
}

// Watchdog The DV4mini returns a ADFWATCHDOG message upon receiving this message.
func (d *DV4Mini) Watchdog() ([]byte, error) {
	// []byte{0x05}
	d.sendCmd([]byte{ADFWATCHDOG})

	bHead := make([]byte, 6)
	_, err := d.Port.Read(bHead)
	if err != nil {
		return bHead, err
	}

	if debug {
		log.Printf("\t[*] serial.read: %#v (len: %d)\n", bHead, len(bHead))
	}

	bParams := make([]byte, int(bHead[5]))
	_, err = d.Port.Read(bParams)
	if err != nil {
		return bParams, err
	}

	if debug {
		log.Printf("\t[*] serial.read: %#v (len: %d)\n", bParams, len(bParams))
	}

	b := bHead
	b = append(b, bParams...)

	// Flush serial buffer before receiving more data
	d.Port.Flush()

	return b, nil
}

// Version returns the firmware version and dongle ID byte array.
func (d *DV4Mini) Version() ([]byte, error) {
	// []byte{0x18}
	d.sendCmd([]byte{ADFVERSION})

	bHead := make([]byte, 6)
	_, err := d.Port.Read(bHead)
	if err != nil {
		return bHead, err
	}

	if debug {
		log.Printf("\t[*] serial.read: %#v (len: %d)\n", bHead, len(bHead))
	}

	bParams := make([]byte, int(bHead[5]))
	_, err = d.Port.Read(bParams)
	if err != nil {
		return bParams, err
	}

	if debug {
		log.Printf("\t[*] serial.read: %#v (len: %d)\n", bParams, len(bParams))
	}

	b := bHead
	b = append(b, bParams...)

	// Flush serial buffer before receiving more data
	d.Port.Flush()

	return b, nil
}

// SetOperatingMode
func (d *DV4Mini) SetOperatingMode(mode byte) {
	// []byte{0x02, mode}
	d.sendCmd([]byte{SETADFMODE, mode})
}

// SetInitialSeed
func (d *DV4Mini) SetInitialSeed() error {
	// []byte{0x11, seed}
	b, err := GetRandBytes(4)
	if err != nil {
		return err
	}

	cmd := []byte{ADFSETSEED}
	cmd = append(cmd, b...)

	d.sendCmd(cmd)

	return nil
}

// SetFrequency
func (d *DV4Mini) SetFrequency(txqrg, rxqrg []byte) {
	// []byte{0x01, txqrg, rxqrg}
	cmd := []byte{SETADFQRG}
	cmd = append(cmd, txqrg...)
	cmd = append(cmd, rxqrg...)

	d.sendCmd(cmd)
}

// SetTXPower
func (d *DV4Mini) SetTXPower(pwr uint8) {
	// []byte{0x09, pwr}
	d.sendCmd([]byte{ADFSETPOWER, pwr})
}

// SetTXBuffer
func (d *DV4Mini) SetTXBuffer(size int) error {
	if size >= 1 || size <= 15 {
		cmd := []byte{ADFSETTXBUF}
		cmd = append(cmd, byte(size))
		d.sendCmd(cmd)

		return nil
	}

	return errors.New("Buffer size must be >=1 || <= 15 (100ms to 1500ms)")
}

// GetRXBufferData
func (d *DV4Mini) GetRXBufferData() ([]byte, error) {
	d.sendCmd([]byte{ADFGETDATA})

	bHead := make([]byte, 6)
	_, err := d.Port.Read(bHead)
	if err != nil {
		return bHead, err
	}

	if debug {
		log.Printf("\t[*] serial.read: %#v (len: %d)\n", bHead, len(bHead))
	}

	bParams := make([]byte, int(bHead[5]))
	_, err = d.Port.Read(bParams)
	if err != nil {
		return bParams, err
	}

	if debug {
		log.Printf("\t[*] serial.read: %#v (len: %d)\n", bParams, len(bParams))
	}

	b := bHead
	b = append(b, bParams...)

	// Flush serial buffer before receiving more data
	d.Port.Flush()

	return b, nil
}

// WriteTXBufferData writes to transmission buffer, the PTT (W/ red LED) is
// triggered automatically
func (d *DV4Mini) WriteTXBufferData(data []byte) {
	var packetSize int = 34 // full 36
	// var counter int = 0
	// var crcValue byte

	// []byte{0x04, data}
	cmd := []byte{ADFWRITE}

	for i := 0; i < len(data); i += packetSize {
		time.Sleep(time.Millisecond * 30)

		batch := data[i:min(i+packetSize, len(data))]

		// if (counter % 2) == 0 {
		// 	batch = append([]byte{0x91}, batch...)
		// } else {
		// 	batch = append([]byte{0x23}, batch...)
		// }

		// counter++
		log.Printf("[>>>] \n%s", hex.Dump(batch))

		fullPacket := cmd
		fullPacket = append(fullPacket, batch...)

		d.sendCmd(fullPacket)
	}

	// d.FlushTXBuffer()
	d.Port.Flush()
}

// sendCmd Sends command to the dv4mini.
func (d *DV4Mini) sendCmd(data []byte) {
	b := CmdPreamble
	params := data[1:]
	cmd := data[0]

	// Set command
	b = append(b, cmd)
	// Set param length
	b = append(b, byte(len(params)))
	// Set params
	b = append(b, params...)

	if debug {
		log.Printf("\t[*] serial.write: %#v (len: %d)\n", b, len(b))
	}

	_, err := d.Port.Write(b)
	if err != nil {
		fmt.Println(err)
	}
}

/*
WRaw sends a raw command to the dongle, useful for debugging and
testing purposes. The full packed must be crafted by hand with the
format []byte{preamble, command, length, params}.

An example for setting the TX and RX frequency:
	dv, err := dv4mini.Connect(DEVICE, true)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure to close it later.
	defer dv.Close()

	c := []byte{
		0x71, 0xfe, 0x39, 0x1d, // Preamble
		0x01,                   // Command
		0x08,                   // Length
		0x19, 0xfc, 0xd3, 0x70, // TX freq.
		0x19, 0xfc, 0xd3, 0x70, // RX freq.
	}
	dv.WRaw(c)
*/
func (d *DV4Mini) WRaw(data []byte) {
	if debug {
		log.Printf("\t[*] serial.write: %#v (len: %d)\n", data, len(data))
	}

	_, err := d.Port.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	// log.Printf("\t[*] Wroten: %d bytes.\n", n)
}

/*
RWRaw sends a raw command to the dongle and reads out the response,
useful for debugging purposes.  The full packed must be crafted by hand with the
format []byte{preamble, command, length, params}.

An example for getting dongle's version:
	dv, err := dv4mini.Connect(DEVICE, true)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure to close it later.
	defer dv.Close()

	c := []byte{
		0x71, 0xfe, 0x39, 0x1d, // Preamble
		0x18,                   // Command
		0x00,                   // Length
	}
	buff, err dv.RWRaw(c)
	[...]
*/
func (d *DV4Mini) RWRaw(data []byte) {
	if debug {
		log.Printf("\t[*]> serial.write: %#v (len: %d)\n", data, len(data))
	}

	_, err := d.Port.Write(data)
	if err != nil {
		fmt.Println(err)
	}

	// b := make([]byte, n)
	// _, err := d.Port.Read(b)
	// if err != nil {
	// 	log.Printf("%#v", err)
	// }

	// if debug {
	// 	log.Printf("\t[*] serial.read: %#v (len: %d)\n", b, len(b))
	// }

	// log.Printf("%#v", b)

	bHead := make([]byte, 6)
	_, err = d.Port.Read(bHead)
	if err != nil {
		log.Printf("%#v", err)
	}

	if debug {
		log.Printf("\t[*] serial.read: %#v (len: %d)\n", bHead, len(bHead))
	}

	bParams := make([]byte, int(bHead[5]))
	_, err = d.Port.Read(bParams)
	if err != nil {
		log.Printf("%#v", err)
	}

	if debug {
		log.Printf("\t[*] serial.read: %#v (len: %d)\n", bParams, len(bParams))
	}

	b := bHead
	b = append(b, bParams...)

	// Flush serial buffer before receiving more data
	// d.Port.Flush()

	log.Printf("%#v", b)
}

// ReadSerial
func (d *DV4Mini) ReadSerial(bufferSize int) ([]byte, error) {
	b := make([]byte, bufferSize)

	_, err := d.Port.Read(b)
	if err != nil {
		return b, err
	}

	return b, nil
}

// ===========
// = Helpers =
// ===========

func GetRandBytes(size int) (b []byte, err error) {
	b = make([]byte, size)

	_, err = rand.Read(b)
	if err != nil {
		return
	}

	return
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

//TODO: G_8(x)=x^8+x^2+x+1
func crc8() {
	return
}

// G(x) = x^9+x^6+x^4+x^3+1
func crc9(crc *uint16, b uint8, bits int) {
	var v uint8 = 0x80
	for i := 0; i < 8-bits; i++ {
		v >>= 1
	}
	for i := 0; i < 8; i++ {
		xor := (*crc)&0x0100 > 0
		(*crc) <<= 1
		// Limit the number of shift registers to 9.
		*crc &= 0x01ff
		if b&v > 0 {
			(*crc)++
		}
		if xor {
			(*crc) ^= 0x0059
		}
		v >>= 1
	}
}

// func TestCRC9(t *testing.T) {
// 	tests := map[uint16][]byte{
// 		0x0000: []byte{},
// 		0x0100: []byte{0x00, 0x01},
// 		0x0179: []byte("hello world"),
// 	}

// 	for want, test := range tests {
// 		var crc uint16
// 		for _, b := range test {
// 			crc9(&crc, b, 8)
// 		}
// 		crc9end(&crc, 8)
// 		if crc != want {
// 			t.Fatalf("crc9 %v failed: %#04x != %#04x", test, crc, want)
// 		}
// 	}
// }
