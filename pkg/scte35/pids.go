package scte35

// PIDs holds collections of PIDs by type for scte35.Stream.
type PIDs struct {
	PMTPIDs    []uint16
	PCRPIDs    []uint16
	SCTE35PIDs []uint16
}

func (p *PIDs) isPMTPID(pid uint16) bool {
	return isIn16(p.PMTPIDs, pid)
}

func (p *PIDs) addPMTPID(pid uint16) {
	if !p.isPMTPID(pid) {
		p.PMTPIDs = append(p.PMTPIDs, pid)
	}
}

func (p *PIDs) isPCRPID(pid uint16) bool {
	return isIn16(p.PCRPIDs, pid)
}

func (p *PIDs) addPCRPID(pid uint16) {
	if !p.isPCRPID(pid) {
		p.PCRPIDs = append(p.PCRPIDs, pid)
	}
}

func (p *PIDs) isSCTE35PID(pid uint16) bool {
	return isIn16(p.SCTE35PIDs, pid)
}

func (p *PIDs) addSCTE35PID(pid uint16) {
	if !(p.isSCTE35PID(pid)) {
		p.SCTE35PIDs = append(p.SCTE35PIDs, pid)
	}
}
func (p *PIDs) delSCTE35PID(pid uint16) {
	n := 0
	for _, val := range p.SCTE35PIDs {
		if val != pid {
			p.SCTE35PIDs[n] = val
			n++
		}
	}

	p.SCTE35PIDs = p.SCTE35PIDs[:n]
}
