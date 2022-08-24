// Copyright 2022 Adrian of Doom
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or   implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package scte35

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const (
	// packetSize is the size of an MPEG-TS packet in bytes.
	packetSize = 188

	// readSize is the size of a read when parsing files.
	readSize = 13000 * packetSize
)

type PacketData struct {
	PacketNumber int     `json:",omitempty"`
	PID          uint16  `json:",omitempty"`
	Program      uint16  `json:",omitempty"`
	PCR          float64 `json:",omitempty"`
	PTS          float64 `json:",omitempty"`
}

// Stream for parsing MPEGTS for SCTE-35
type Stream struct {
	pktNum       int // packet count.
	programs     []uint16
	pidToProgram map[uint16]uint16 //lookup table for pid to program
	programToPCR map[uint16]uint64 //lookup table for program to pcr
	programToPTS map[uint16]uint64 //lookup table for program to pts
	partial      map[uint16][]byte // partial manages tables spread across multiple packets by pid
	last         map[uint16][]byte // last compares current packet payload to last packet payload by pid
	PIDs
}

func (st *Stream) mkMaps() {
	st.pidToProgram = make(map[uint16]uint16)
	st.last = make(map[uint16][]byte)
	st.partial = make(map[uint16][]byte)
	st.programToPCR = make(map[uint16]uint64)
	st.programToPTS = make(map[uint16]uint64)
}

// Decode fname (a file name) for SCTE-35
func (st *Stream) Decode(fname string) {
	st.mkMaps()
	st.pktNum = 0
	file, err := os.Open(fname)
	check(err)
	defer file.Close()
	buffer := make([]byte, readSize)
	for {
		bytesread, err := file.Read(buffer)
		if err != nil {
			break
		}
		for i := 1; i <= (bytesread / packetSize); i++ {
			end := i * packetSize
			start := end - packetSize
			p := buffer[start:end]
			pkt := &p
			st.pktNum++
			st.parse(*pkt)
		}
	}
}

func (st *Stream) makePCR(prgm uint16) float64 {
	pcrb := st.programToPCR[prgm]
	return make90K(pcrb)
}

func (st *Stream) makePTS(prgm uint16) float64 {
	pts := st.programToPTS[prgm]
	return make90K(pts)
}

func (st *Stream) parsePUSI(pkt []byte) bool {
	if (pkt[1]>>6)&1 == 1 {
		if pkt[6]&1 == 1 {
			return true
		}
	}
	return false
}

func (st *Stream) parsePTS(pkt []byte, pid uint16) {
	if st.parsePUSI(pkt) {
		prgm, ok := st.pidToProgram[pid]
		if ok {
			pts := (uint64(pkt[13]) >> 1 & 7) << 30
			pts |= uint64(pkt[14]) << 22
			pts |= (uint64(pkt[15]) >> 1) << 15
			pts |= uint64(pkt[16]) << 7
			pts |= uint64(pkt[17]) >> 1
			st.programToPTS[prgm] = pts
		}
	}
}

//
func (st *Stream) parsePCR(pkt []byte, pid uint16) {
	if (pkt[3]>>5)&1 == 1 {
		if (pkt[5]>>4)&1 == 1 {
			pcr := (uint64(pkt[6]) << 25)
			pcr |= (uint64(pkt[7]) << 17)
			pcr |= (uint64(pkt[8]) << 9)
			pcr |= (uint64(pkt[9]) << 1)
			pcr |= uint64(pkt[10]) >> 7
			prgm := st.pidToProgram[pid]
			st.programToPCR[prgm] = pcr
		}
	}
}

// parsePayload packet payload starts after header and afc (if present)
func (st *Stream) parsePayload(pkt []byte) []byte {
	head := 4
	hasafc := (pkt[3] >> 5) & 1
	if hasafc == 1 {
		afl := int(pkt[4])
		head += afl + 1
	}
	if head > packetSize {
		head = packetSize
	}
	return pkt[head:]
}

// checkPartial appends the current packet payload to partial table by pid.
func (st *Stream) checkPartial(pay []byte, pid uint16, sep []byte) []byte {
	val, ok := st.partial[pid]
	if ok {
		pay = append(val, pay...)
	}
	return splitByIndex(pay, sep)
}

// sameAsLast compares the current packet to the last packet by pid.
func (st *Stream) sameAsLast(pay []byte, pid uint16) bool {
	val, ok := st.last[pid]
	if ok {
		if bytes.Compare(pay, val) == 0 {
			return true
		}
	}
	st.last[pid] = pay
	return false
}

// sectionDone aggregates partial tables by pid until the section is complete.
func (st *Stream) sectionDone(pay []byte, pid uint16, seclen uint16) bool {
	if seclen+3 > uint16(len(pay)) {
		st.partial[pid] = pay
		return false
	}
	delete(st.partial, pid)
	return true
}

// parse parses an MPEGTS packet based on the pid.
func (st *Stream) parse(pkt []byte) {
	p := parsePID(pkt[1], pkt[2])
	pid := &p
	pl := st.parsePayload(pkt)
	pay := &pl

	if *pid == 0 {
		st.parsePAT(*pay, *pid)
	}
	if st.isPMTPID(*pid) {
		st.parsePMT(*pay, *pid)
	}
	if st.isPCRPID(*pid) {
		st.parsePCR(pkt, *pid)
	} else {
		st.parsePTS(pkt, *pid)
	}
	if st.isSCTE35PID(*pid) {
		st.parseScte35(*pay, *pid)
	}
}

func (st *Stream) parsePAT(pay []byte, pid uint16) {
	if st.sameAsLast(pay, pid) {
		return
	}
	pay = st.checkPartial(pay, pid, []byte("\x00\x00"))
	if len(pay) < 1 {
		return
	}
	seclen := parseLength(pay[2], pay[3])
	if st.sectionDone(pay, pid, seclen) {
		seclen -= 5 // pay bytes 4,5,6,7,8
		idx := uint16(9)
		end := idx + seclen - 4 //  4 bytes for crc
		chunksize := uint16(4)
		for idx < end {
			prgm := parseProgram(pay[idx], pay[idx+1])
			if prgm > 0 {
				if !isIn16(st.programs, prgm) {
					st.programs = append(st.programs, prgm)
				}
				pmtpid := parsePID(pay[idx+2], pay[idx+3])
				st.addPMTPID(pmtpid)
			}
			idx += chunksize
		}
	}
}

func (st *Stream) parsePMT(pay []byte, pid uint16) {
	if st.sameAsLast(pay, pid) {
		return
	}
	pay = st.checkPartial(pay, pid, []byte("\x02"))
	if len(pay) < 1 {
		return
	}
	secinfolen := parseLength(pay[1], pay[2])
	if st.sectionDone(pay, pid, secinfolen) {
		prgm := parseProgram(pay[3], pay[4])
		pcrpid := parsePID(pay[8], pay[9])
		st.addPCRPID(pcrpid)
		proginfolen := parseLength(pay[10], pay[11])
		idx := uint16(12)
		idx += proginfolen
		silen := secinfolen - 9
		silen -= proginfolen
		st.parseStreams(silen, pay, idx, prgm)
	}
}

func (st *Stream) parseStreams(silen uint16, pay []byte, idx uint16, prgm uint16) {
	chunksize := uint16(5)
	endidx := (idx + silen) - chunksize
	for idx < endidx {
		streamtype := pay[idx]
		elpid := parsePID(pay[idx+1], pay[idx+2])
		eilen := parseLength(pay[idx+3], pay[idx+4])
		idx += chunksize
		idx += eilen
		st.pidToProgram[elpid] = prgm
		st.verifyStreamType(elpid, streamtype)
	}
}

func (st *Stream) verifyStreamType(pid uint16, streamtype uint8) {
	if streamtype == 6 || streamtype == 134 {
		st.addSCTE35PID(pid)
	}
}

func (st *Stream) parseScte35(pay []byte, pid uint16) {
	pay = st.checkPartial(pay, pid, []byte("\xfc0"))
	if len(pay) == 0 {
		st.PIDs.delSCTE35PID(pid)
		return
	}
	seclen := parseLength(pay[1], pay[2])
	if st.sectionDone(pay, pid, seclen) {
		sis := st.makeSpliceInfoSection(pid)
		sis.Decode(pay)
		b, _ := json.MarshalIndent(sis, "", "\t")
		_, _ = fmt.Fprintf(os.Stdout, "\nSplice Info Section: \n%s\n", b)
	}
}

func (st *Stream) makeSpliceInfoSection(pid uint16) *SpliceInfoSection {
	sis := &SpliceInfoSection{}
	p := st.pidToProgram[pid]
	prgm := &p
	var packet PacketData
	packet.PID = pid
	packet.Program = *prgm
	packet.PCR = st.makePCR(*prgm)
	packet.PTS = st.makePTS(*prgm)
	packet.PacketNumber = st.pktNum
	pkt, _ := json.MarshalIndent(packet, "", "\t")
	_, _ = fmt.Fprintf(os.Stdout, "\nPacket Data: \n%s\n", pkt)
	return sis
}

// isIn16 is a test for slice membership
func isIn16(slice []uint16, val uint16) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// isIn8 is a test for slice membership
func isIn8(slice []uint8, val uint8) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func make90K(raw uint64) float64 {
	nk := float64(raw) / 90000.0
	return float64(uint64(nk*1000000)) / 1000000
}

func parseLength(byte1 byte, byte2 byte) uint16 {
	return uint16(byte1&0xf)<<8 | uint16(byte2)
}

func parsePID(byte1 byte, byte2 byte) uint16 {
	return uint16(byte1&0x1f)<<8 | uint16(byte2)
}

func parseProgram(byte1 byte, byte2 byte) uint16 {
	return uint16(byte1)<<8 | uint16(byte2)
}

func splitByIndex(payload, sep []byte) []byte {
	idx := bytes.Index(payload, sep)
	if idx == -1 {
		return []byte("")
	}
	return payload[idx:]
}

// check generic catchall error checking
func check(e error) {
	if e != nil {
		fmt.Println(e)

	}
}
