// Copyright © 2018 Nathan Caldwell <saintdev@gmail.com>
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
