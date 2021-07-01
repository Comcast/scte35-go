// Copyright 2021 Comcast Cable Communications Management, LLC
//
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
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/bamiaux/iobit"
)

const (
	// SpliceInsertType is the splice_command_type for splice_insert()
	SpliceInsertType = 0x05
)

// SpliceInsert is a  command shall be sent at least once for every splice
// event.
type SpliceInsert struct {
	XMLName                    xml.Name                `xml:"http://www.scte.org/schemas/35 SpliceInsert" json:"-"`
	JSONType                   uint32                  `xml:"-" json:"type"`
	Program                    *SpliceInsertProgram    `xml:"http://www.scte.org/schemas/35 Program" json:"program,omitempty"`
	Components                 []SpliceInsertComponent `xml:"http://www.scte.org/schemas/35 Component" json:"components,omitempty"`
	BreakDuration              *BreakDuration          `xml:"http://www.scte.org/schemas/35 BreakDuration" json:"breakDuration,omitempty"`
	SpliceEventID              uint32                  `xml:"spliceEventId,attr" json:"spliceEventId,omitempty"`
	SpliceEventCancelIndicator bool                    `xml:"spliceEventCancelIndicator,attr" json:"spliceEventCancelIndicator"`
	SpliceImmediateFlag        bool                    `xml:"spliceImmediateFlag,attr" json:"spliceImmediateFlag"`
	OutOfNetworkIndicator      bool                    `xml:"outOfNetworkIndicator,attr" json:"outOfNetworkIndicator"`
	UniqueProgramID            uint32                  `xml:"uniqueProgramId,attr" json:"uniqueProgramId,omitempty"`
	AvailNum                   uint32                  `xml:"availNum,attr" json:"availNum,omitempty"`
	AvailsExpected             uint32                  `xml:"availsExpected,attr" json:"availsExpected,omitempty"`
}

// Type returns the splice_command_type.
func (cmd *SpliceInsert) Type() uint32 {
	cmd.JSONType = SpliceInsertType
	return SpliceInsertType
}

// String returns the table description of this splice_insert.
func (cmd *SpliceInsert) String() string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("splice_insert() {\n"))
	buf.WriteString(fmt.Sprintf("    splice_event_id: %d\n", cmd.SpliceEventID))
	buf.WriteString(fmt.Sprintf("    splice_event_cancel_indicator: %v\n", cmd.SpliceEventCancelIndicator))
	if !cmd.SpliceEventCancelIndicator {
		buf.WriteString(fmt.Sprintf("    out_of_network_indicator: %v\n", cmd.OutOfNetworkIndicator))
		buf.WriteString(fmt.Sprintf("    program_splice_flag: %v\n", cmd.programSpliceFlag()))
		buf.WriteString(fmt.Sprintf("    duration_flag: %v\n", cmd.durationFlag()))
		buf.WriteString(fmt.Sprintf("    splice_immediate_flag: %v\n", cmd.SpliceImmediateFlag))
		if cmd.programSpliceFlag() && !cmd.SpliceImmediateFlag {
			buf.WriteString(fmt.Sprintf("    time_specified_flag: %v\n", cmd.timeSpecifiedFlag()))
			if cmd.timeSpecifiedFlag() {
				buf.WriteString(fmt.Sprintf("    pts_time: %d ticks (%s)\n", *cmd.Program.SpliceTime.PTSTime, TicksToDuration(*cmd.Program.SpliceTime.PTSTime)))
			}
		}
		if !cmd.programSpliceFlag() {
			buf.WriteString(fmt.Sprintf("    component_count: %d\n", len(cmd.Components)))
			for j, c := range cmd.Components {
				buf.WriteString(fmt.Sprintf("    component[%d] {\n", j))
				buf.WriteString(fmt.Sprintf("        component_tag: %d\n", c.Tag))
				if !cmd.SpliceImmediateFlag {
					buf.WriteString(fmt.Sprintf("        time_specified_flag: %v\n", c.timeSpecifiedFlag()))
					if c.timeSpecifiedFlag() {
						buf.WriteString(fmt.Sprintf("        pts_time: %d ticks (%s)\n", *c.SpliceTime.PTSTime, TicksToDuration(*c.SpliceTime.PTSTime)))
					}
				}
				buf.WriteString(fmt.Sprintf("    }\n"))
			}
		}
		if cmd.durationFlag() {
			buf.WriteString(fmt.Sprintf("    auto_return: %v\n", cmd.BreakDuration.AutoReturn))
			buf.WriteString(fmt.Sprintf("    duration: %d ticks (%s)\n", cmd.BreakDuration.Duration, TicksToDuration(cmd.BreakDuration.Duration)))
		}
		buf.WriteString(fmt.Sprintf("    unique_program_id: %d\n", cmd.UniqueProgramID))
		buf.WriteString(fmt.Sprintf("    avail_num: %d\n", cmd.AvailNum))
		buf.WriteString(fmt.Sprintf("    avails_expected: %d\n", cmd.AvailsExpected))
		buf.WriteString(fmt.Sprintf("}\n"))
	}

	return buf.String()
}

// decode a binary splice_insert.
func (cmd *SpliceInsert) decode(b []byte) error {
	r := iobit.NewReader(b)

	cmd.SpliceEventID = r.Uint32(32)
	cmd.SpliceEventCancelIndicator = r.Bit()
	r.Skip(7) // reserved
	if !cmd.SpliceEventCancelIndicator {
		cmd.OutOfNetworkIndicator = r.Bit()
		programSpliceFlag := r.Bit()
		durationFlag := r.Bit()
		cmd.SpliceImmediateFlag = r.Bit()
		r.Skip(4) // reserved
		if programSpliceFlag {
			cmd.Program = &SpliceInsertProgram{}
			if !cmd.SpliceImmediateFlag {
				timeSpecifiedFlag := r.Bit()
				if timeSpecifiedFlag {
					r.Skip(6) // reserved
					ptsTime := r.Uint64(33)
					cmd.Program.SpliceTime.PTSTime = &ptsTime
				} else {
					r.Skip(7) // reserved
				}
			}
		} else {
			componentCount := int(r.Uint32(8))
			cmd.Components = make([]SpliceInsertComponent, componentCount)
			for i := 0; i < componentCount; i++ {
				c := SpliceInsertComponent{}
				c.Tag = r.Uint32(8)
				if !cmd.SpliceImmediateFlag {
					timeSpecifiedFlag := r.Bit()
					if timeSpecifiedFlag {
						r.Skip(6) // reserved
						ptsTime := r.Uint64(33)
						c.SpliceTime = &SpliceTime{
							PTSTime: &ptsTime,
						}
					} else {
						r.Skip(7) // reserved
					}
				}
				cmd.Components[i] = c
			}
		}
		if durationFlag {
			cmd.BreakDuration = &BreakDuration{}
			cmd.BreakDuration.AutoReturn = r.Bit()
			r.Skip(6) // reserved
			cmd.BreakDuration.Duration = r.Uint64(33)
		}
	}
	cmd.UniqueProgramID = r.Uint32(16)
	cmd.AvailNum = r.Uint32(8)
	cmd.AvailsExpected = r.Uint32(8)

	if err := readerError(r); err != nil {
		return fmt.Errorf("splice_insert: %w", err)
	}
	return nil
}

// encode this splice_insert to binary.
func (cmd *SpliceInsert) encode() ([]byte, error) {
	buf := make([]byte, cmd.length())

	iow := iobit.NewWriter(buf)
	iow.PutUint32(32, cmd.SpliceEventID)
	iow.PutBit(cmd.SpliceEventCancelIndicator)
	iow.PutUint32(7, Reserved)
	if !cmd.SpliceEventCancelIndicator {
		iow.PutBit(cmd.OutOfNetworkIndicator)
		iow.PutBit(cmd.programSpliceFlag())
		iow.PutBit(cmd.durationFlag())
		iow.PutBit(cmd.SpliceImmediateFlag)
		iow.PutUint32(4, Reserved)
		if cmd.programSpliceFlag() && !cmd.SpliceImmediateFlag {
			if cmd.Program.timeSpecifiedFlag() {
				iow.PutBit(true)
				iow.PutUint32(6, Reserved)
				iow.PutUint64(33, *cmd.Program.SpliceTime.PTSTime)
			} else {
				iow.PutBit(false)
				iow.PutUint32(7, Reserved)
			}
		}
		if !cmd.programSpliceFlag() {
			iow.PutUint32(8, uint32(len(cmd.Components)))
			for _, c := range cmd.Components {
				iow.PutUint32(8, c.Tag)
				if !cmd.SpliceImmediateFlag {
					if c.timeSpecifiedFlag() {
						iow.PutBit(true)
						iow.PutUint32(6, Reserved)
						iow.PutUint64(33, *c.SpliceTime.PTSTime)
					} else {
						iow.PutBit(false)
						iow.PutUint32(7, Reserved)
					}
				}
			}
		}
		if cmd.durationFlag() {
			iow.PutBit(cmd.BreakDuration.AutoReturn)
			iow.PutUint32(6, Reserved)
			iow.PutUint64(33, cmd.BreakDuration.Duration)
		}
		iow.PutUint32(16, cmd.UniqueProgramID)
		iow.PutUint32(8, cmd.AvailNum)
		iow.PutUint32(8, cmd.AvailsExpected)
	}

	err := iow.Flush()
	return buf, err
}

// length returns the splice_command_length.
func (cmd SpliceInsert) length() int {
	length := 32 // splice_event_id
	length++     // splice_event_cancel_indicator
	length += 7  // reserved

	// if splice_event_cancel_indicator == 0
	if !cmd.SpliceEventCancelIndicator {
		length++    // out_of_network_indicator
		length++    // program_splice_flag
		length++    // duration_flag
		length++    // splice_immediate_flag
		length += 4 // reserved

		// if program_splice_flag == 1 && splice_immediate_flag == 0
		if cmd.programSpliceFlag() && !cmd.SpliceImmediateFlag {
			length++ // time_specified_flag

			// if time_specified_flag == 1
			if cmd.Program.timeSpecifiedFlag() {
				length += 6  // reserved
				length += 33 // pts_time
			} else {
				length += 7 // reserved
			}
		}

		// if program_splice_flag == 0
		if !cmd.programSpliceFlag() {
			length += 8 // component_count

			// for i = 0 to component_count
			for _, c := range cmd.Components {
				length += 8 // component_tag

				// if splice_immediate_flag == 0
				if !cmd.SpliceImmediateFlag {
					length++ // time_specified_flag

					// if time_specified_flag == 1
					if c.timeSpecifiedFlag() {
						length += 6  // reserved
						length += 33 // pts_time
					} else {
						length += 7 // reserved
					}
				}
			}
		}

		// if duration_flag == 1
		if cmd.durationFlag() {
			length++     // auto_return
			length += 6  // reserved
			length += 33 // duration
		}

		length += 16 // unique_program_id
		length += 8  // avail_num
		length += 8  // avails_expected
	}

	return length / 8
}

// durationFlag returns the duration_flag.
func (cmd *SpliceInsert) durationFlag() bool {
	return cmd.BreakDuration != nil
}

// programSpliceFlag returns the program_splice_flag.
func (cmd *SpliceInsert) programSpliceFlag() bool {
	return cmd.Program != nil
}

// timeSpecifiedFlag returns the time_specified_flag
func (cmd *SpliceInsert) timeSpecifiedFlag() bool {
	return cmd.Program == nil && cmd.Program.SpliceTime.PTSTime != nil
}

// SpliceInsertComponent contains the Splice Point in Component Splice Mode.
type SpliceInsertComponent struct {
	Tag        uint32      `xml:"componentTag,attr" json:"componentTag,omitempty"`
	SpliceTime *SpliceTime `xml:"http://www.scte.org/schemas/35 SpliceTime" json:"spliceTime,omitempty"`
}

// timeSpecifiedFlag returns the time_specified_flag.
func (c *SpliceInsertComponent) timeSpecifiedFlag() bool {
	return c != nil && c.SpliceTime != nil && c.SpliceTime.PTSTime != nil
}

// NewSpliceInsertProgram returns a SpliceInsertProgram with the given ptsTime.
func NewSpliceInsertProgram(ptsTime uint64) *SpliceInsertProgram {
	return &SpliceInsertProgram{
		SpliceTime: SpliceTime{
			PTSTime: &ptsTime,
		},
	}
}

// SpliceInsertProgram contains the Splice Point in Program Splice Mode.
type SpliceInsertProgram struct {
	SpliceTime SpliceTime `xml:"http://www.scte.org/schemas/35 SpliceTime" json:"spliceTime"`
}

// timeSpecifiedFlag returns the time_specified_flag.
func (p *SpliceInsertProgram) timeSpecifiedFlag() bool {
	return p != nil && p.SpliceTime.PTSTime != nil
}
