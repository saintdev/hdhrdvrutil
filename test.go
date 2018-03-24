package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ziutek/dvb/ts"
)

func dumprecordingjson(filename string) {

	fmt.Printf("%v:\n", filename)

	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		log.Fatal("Error: Unable to open file", filename, err)
	}

	tsfile := ts.NewPktStreamReader(bufio.NewReader(file))

	var buf [ts.PktLen]byte
	var jsonBuf []byte

	for i := 0; i < 64; i++ {
		pkt := ts.AsPkt(buf[:])
		if err := tsfile.ReadPkt(pkt); err != nil {
			log.Print("Error: Unable to read TS packet", err)
			continue
		}
		if pkt.Pid() != 0x1FFA {
			break
		}
		payload := pkt.Payload()
		jsonBuf = append(jsonBuf, payload...)
	}

	jsonBuf = bytes.Trim(jsonBuf, "\xFF")

	var out bytes.Buffer
	json.Indent(&out, jsonBuf, "", "  ")
	out.WriteTo(os.Stdout)
	fmt.Println("")
}
