package scte35

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// pktSz is the size of an MPEG-TS packet in bytes.
const pktSz = 188

// bufSz is the size of a read when parsing files.
const bufSz = 13000 * pktSz

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

func mk90k(raw uint64) float64 {
	nk := float64(raw) / 90000.0
	return float64(uint64(nk*1000000)) / 1000000
}

func parseLen(byte1 byte, byte2 byte) uint16 {
	return uint16(byte1&0xf)<<8 | uint16(byte2)
}

func parsePid(byte1 byte, byte2 byte) uint16 {
	return uint16(byte1&0x1f)<<8 | uint16(byte2)
}

func parsePrgm(byte1 byte, byte2 byte) uint16 {
	return uint16(byte1)<<8 | uint16(byte2)
}

func splitByIdx(payload, sep []byte) []byte {
	idx := bytes.Index(payload, sep)
	if idx == -1 {
		return []byte("")
	}
	return payload[idx:]
}

//chk generic catchall error checking
func chk(e error) {
	if e != nil {
		fmt.Println(e)

	}
}

type PacketData struct {
	PacketNumber int     `json:",omitempty"`
	Pid          uint16  `json:",omitempty"`
	Program      uint16  `json:",omitempty"`
	Pcr          float64 `json:",omitempty"`
	Pts          float64 `json:",omitempty"`
}

//Pids holds collections of pids by type for threefive.Stream.
type Pids struct {
	PmtPids    []uint16
	PcrPids    []uint16
	Scte35Pids []uint16
}

func (pids *Pids) isPmtPid(pid uint16) bool {
	return isIn16(pids.PmtPids, pid)
}

func (pids *Pids) addPmtPid(pid uint16) {
	if !pids.isPmtPid(pid) {
		pids.PmtPids = append(pids.PmtPids, pid)
	}
}

func (pids *Pids) isPcrPid(pid uint16) bool {
	return isIn16(pids.PcrPids, pid)
}

func (pids *Pids) addPcrPid(pid uint16) {
	if !pids.isPcrPid(pid) {
		pids.PcrPids = append(pids.PcrPids, pid)
	}
}

func (pids *Pids) isScte35Pid(pid uint16) bool {
	return isIn16(pids.Scte35Pids, pid)
}

func (pids *Pids) addScte35Pid(pid uint16) {
	if !(pids.isScte35Pid(pid)) {
		pids.Scte35Pids = append(pids.Scte35Pids, pid)
	}
}
func (pids *Pids) delScte35Pid(pid uint16) {
	n := 0
	for _, val := range pids.Scte35Pids {
		if val != pid {
			pids.Scte35Pids[n] = val
			n++
		}
	}

	pids.Scte35Pids = pids.Scte35Pids[:n]
}

//Stream for parsing MPEGTS for SCTE-35
type Stream struct {
	pktNum   int // packet count.
	programs []uint16
	pid2Prgm map[uint16]uint16 //lookup table for pid to program
	prgm2pcr map[uint16]uint64 //lookup table for program to pcr
	prgm2pts map[uint16]uint64 //lookup table for program to pts
	partial  map[uint16][]byte // partial manages tables spread across multiple packets by pid
	last     map[uint16][]byte // last compares current packet payload to last packet payload by pid
	Pids
}

func (stream *Stream) mkMaps() {
	stream.pid2Prgm = make(map[uint16]uint16)
	stream.last = make(map[uint16][]byte)
	stream.partial = make(map[uint16][]byte)
	stream.prgm2pcr = make(map[uint16]uint64)
	stream.prgm2pts = make(map[uint16]uint64)
}

// Decode fname (a file name) for SCTE-35
func (stream *Stream) Decode(fname string) {
	stream.mkMaps()
	stream.pktNum = 0
	file, err := os.Open(fname)
	chk(err)
	defer file.Close()
	buffer := make([]byte, bufSz)
	for {
		bytesread, err := file.Read(buffer)
		if err != nil {
			break
		}
		for i := 1; i <= (bytesread / pktSz); i++ {
			end := i * pktSz
			start := end - pktSz
			p := buffer[start:end]
			pkt := &p
			stream.pktNum++
			stream.parse(*pkt)
		}
	}
}

func (stream *Stream) mkPcr(prgm uint16) float64 {
	pcrb := stream.prgm2pcr[prgm]
	return mk90k(pcrb)
}

func (stream *Stream) mkPts(prgm uint16) float64 {
	pts := stream.prgm2pts[prgm]
	return mk90k(pts)
}

func (stream *Stream) parsePusi(pkt []byte) bool {
	if (pkt[1]>>6)&1 == 1 {
		if pkt[6]&1 == 1 {
			return true
		}
	}
	return false
}

func (stream *Stream) parsePts(pkt []byte, pid uint16) {
	if stream.parsePusi(pkt) {
		prgm, ok := stream.pid2Prgm[pid]
		if ok {
			pts := (uint64(pkt[13]) >> 1 & 7) << 30
			pts |= uint64(pkt[14]) << 22
			pts |= (uint64(pkt[15]) >> 1) << 15
			pts |= uint64(pkt[16]) << 7
			pts |= uint64(pkt[17]) >> 1
			stream.prgm2pts[prgm] = pts
		}
	}
}

//
func (stream *Stream) parsePcr(pkt []byte, pid uint16) {
	if (pkt[3]>>5)&1 == 1 {
		if (pkt[5]>>4)&1 == 1 {
			pcr := (uint64(pkt[6]) << 25)
			pcr |= (uint64(pkt[7]) << 17)
			pcr |= (uint64(pkt[8]) << 9)
			pcr |= (uint64(pkt[9]) << 1)
			pcr |= uint64(pkt[10]) >> 7
			prgm := stream.pid2Prgm[pid]
			stream.prgm2pcr[prgm] = pcr
		}
	}
}

//parsePay packet payload starts after header and afc (if present)
func (stream *Stream) parsePayload(pkt []byte) []byte {
	head := 4
	hasafc := (pkt[3] >> 5) & 1
	if hasafc == 1 {
		afl := int(pkt[4])
		head += afl + 1
	}
	if head > pktSz {
		head = pktSz
	}
	return pkt[head:]
}

//chkPartial appends the current packet payload to partial table by pid.
func (stream *Stream) chkPartial(pay []byte, pid uint16, sep []byte) []byte {
	val, ok := stream.partial[pid]
	if ok {
		pay = append(val, pay...)
	}
	return splitByIdx(pay, sep)
}

// sameAsLast compares the current packet to the last packet by pid.
func (stream *Stream) sameAsLast(pay []byte, pid uint16) bool {
	val, ok := stream.last[pid]
	if ok {
		if bytes.Compare(pay, val) == 0 {
			return true
		}
	}
	stream.last[pid] = pay
	return false
}

//sectionDone aggregates partial tables by pid until the section is complete.
func (stream *Stream) sectionDone(pay []byte, pid uint16, seclen uint16) bool {
	if seclen+3 > uint16(len(pay)) {
		stream.partial[pid] = pay
		return false
	}
	delete(stream.partial, pid)
	return true
}

// parse is the parser method for Stream
func (stream *Stream) parse(pkt []byte) {
	p := parsePid(pkt[1], pkt[2])
	pid := &p
	pl := stream.parsePayload(pkt)
	pay := &pl

	if *pid == 0 {
		stream.parsePat(*pay, *pid)
	}
	if stream.isPmtPid(*pid) {
		stream.parsePmt(*pay, *pid)
	}
	if stream.isPcrPid(*pid) {
		stream.parsePcr(pkt, *pid)
	} else {
		stream.parsePts(pkt, *pid)
	}
	if stream.isScte35Pid(*pid) {
		stream.parseScte35(*pay, *pid)
	}
}

func (stream *Stream) parsePat(pay []byte, pid uint16) {
	if stream.sameAsLast(pay, pid) {
		return
	}
	pay = stream.chkPartial(pay, pid, []byte("\x00\x00"))
	if len(pay) < 1 {
		return
	}
	seclen := parseLen(pay[2], pay[3])
	if stream.sectionDone(pay, pid, seclen) {
		seclen -= 5 // pay bytes 4,5,6,7,8
		idx := uint16(9)
		end := idx + seclen - 4 //  4 bytes for crc
		chunksize := uint16(4)
		for idx < end {
			prgm := parsePrgm(pay[idx], pay[idx+1])
			if prgm > 0 {
				if !isIn16(stream.programs, prgm) {
					stream.programs = append(stream.programs, prgm)
				}
				pmtpid := parsePid(pay[idx+2], pay[idx+3])
				stream.addPmtPid(pmtpid)
			}
			idx += chunksize
		}
	}
}

func (stream *Stream) parsePmt(pay []byte, pid uint16) {
	if stream.sameAsLast(pay, pid) {
		return
	}
	pay = stream.chkPartial(pay, pid, []byte("\x02"))
	if len(pay) < 1 {
		return
	}
	secinfolen := parseLen(pay[1], pay[2])
	if stream.sectionDone(pay, pid, secinfolen) {
		prgm := parsePrgm(pay[3], pay[4])
		pcrpid := parsePid(pay[8], pay[9])
		stream.addPcrPid(pcrpid)
		proginfolen := parseLen(pay[10], pay[11])
		idx := uint16(12)
		idx += proginfolen
		silen := secinfolen - 9
		silen -= proginfolen
		stream.parseStreams(silen, pay, idx, prgm)
	}
}

func (stream *Stream) parseStreams(silen uint16, pay []byte, idx uint16, prgm uint16) {
	chunksize := uint16(5)
	endidx := (idx + silen) - chunksize
	for idx < endidx {
		streamtype := pay[idx]
		elpid := parsePid(pay[idx+1], pay[idx+2])
		eilen := parseLen(pay[idx+3], pay[idx+4])
		idx += chunksize
		idx += eilen
		stream.pid2Prgm[elpid] = prgm
		stream.vrfyStreamType(elpid, streamtype)
	}
}

func (stream *Stream) vrfyStreamType(pid uint16, streamtype uint8) {
	if streamtype == 6 || streamtype == 134 {
		stream.addScte35Pid(pid)
	}
}

func (stream *Stream) parseScte35(pay []byte, pid uint16) {
	pay = stream.chkPartial(pay, pid, []byte("\xfc0"))
	if len(pay) == 0 {
		stream.Pids.delScte35Pid(pid)
		return
	}
	seclen := parseLen(pay[1], pay[2])
	if stream.sectionDone(pay, pid, seclen) {
		sis := stream.mkSis(pid)
		sis.Decode(pay)
		b, _ := json.MarshalIndent(sis, "", "\t")
		_, _ = fmt.Fprintf(os.Stdout, "\nSplice Info Section: \n%s\n", b)
	}
}

func (stream *Stream) mkSis(pid uint16) *SpliceInfoSection {
	sis := &SpliceInfoSection{}
	p := stream.pid2Prgm[pid]
	prgm := &p
	var packet PacketData
	packet.Pid = pid
	packet.Program = *prgm
	packet.Pcr = stream.mkPcr(*prgm)
	packet.Pts = stream.mkPts(*prgm)
	packet.PacketNumber = stream.pktNum
	pkt, _ := json.MarshalIndent(packet, "", "\t")
	_, _ = fmt.Fprintf(os.Stdout, "\nPacket Data: \n%s\n", pkt)
	return sis
}


