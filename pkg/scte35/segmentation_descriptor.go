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
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/bamiaux/iobit"
)

const (
	// SegmentationDescriptorTag is the splice_descriptor_tag for
	// segmentation_descriptor
	SegmentationDescriptorTag = 0x02
	// PO Start.
	SegmentationTypeProviderPOStart = 0x34
	// Distributor PO Start.
	SegmentationTypeDistributorPOStart = 0x36
	// SegmentationTypeDistributorPOEnd is the segmentation_type_id for

)

// SegmentationDescriptor is an implementation of a splice_descriptor(). It
// provides an optional extension to the time_signal() and splice_insert()
// commands that allows for segmentation messages to be sent in a time/video
// accurate method. This descriptor shall only be used with the time_signal(),
// splice_insert() and the splice_null() commands. The time_signal() or
// splice_insert() message should be sent at least once a minimum of 4 seconds
// in advance of the signaled splice_time() to permit the insertion device to
// place the splice_info_section( ) accurately.
type SegmentationDescriptor struct {
	XMLName                          xml.Name                          `xml:"http://www.scte.org/schemas/35 SegmentationDescriptor" json:"-"`
	JSONType                         uint32                            `xml:"-" json:"type"`
	DeliveryRestrictions             *DeliveryRestrictions             `xml:"http://www.scte.org/schemas/35 DeliveryRestrictions" json:"deliveryRestrictions"`
	SegmentationUPIDs                []SegmentationUPID                `xml:"http://www.scte.org/schemas/35 SegmentationUpid" json:"segmentationUpids"`
	Components                       []SegmentationDescriptorComponent `xml:"http://www.scte.org/schemas/35 Component" json:"components"`
	SegmentationEventID              uint32                            `xml:"segmentationEventId,attr" json:"segmentationEventId"`
	SegmentationEventCancelIndicator bool                              `xml:"segmentationEventCancelIndicator,attr" json:"segmentationEventCancelIndicator"`
	SegmentationDuration             *uint64                           `xml:"segmentationDuration,attr" json:"segmentationDuration"`
	SegmentationTypeID               uint32                            `xml:"segmentationTypeId,attr" json:"segmentationTypeId"`
	SegmentNum                       uint32                            `xml:"segmentNum,attr" json:"segmentNum"`
	SegmentsExpected                 uint32                            `xml:"segmentsExpected,attr" json:"segmentsExpected"`
	SubSegmentNum                    *uint32                           `xml:"subSegmentNum,attr" json:"subSegmentNum"`
	SubSegmentsExpected              *uint32                           `xml:"subSegmentsExpected,attr" json:"subSegmentsExpected"`
}

// Name returns the human readable string for the segmentation_type_id.
func (sd *SegmentationDescriptor) Name() string {

	var table22 = map[uint32]string{
		0x00: "Not Indicated",
		0x01: "Content Identification",
		0x10: "Program Start",
		0x11: "Program End",
		0x12: "Program Early Termination",
		0x13: "Program Breakaway",
		0x14: "Program Resumption",
		0x15: "Program Runover Planned",
		0x16: "Program RunoverUnplanned",
		0x17: "Program Overlap Start",
		0x18: "Program Blackout Override",
		0x19: "Program Start ??? In Progress",
		0x20: "Chapter Start",
		0x21: "Chapter End",
		0x22: "Break Start",
		0x23: "Break End",
		0x24: "Opening Credit Start",
		0x25: "Opening Credit End",
		0x26: "Closing Credit Start",
		0x27: "Closing Credit End",
		0x30: "Provider Advertisement Start",
		0x31: "Provider Advertisement End",
		0x32: "Distributor Advertisement Start",
		0x33: "Distributor Advertisement End",
		0x34: "Provider Placement Opportunity Start",
		0x35: "Provider Placement Opportunity End",
		0x36: "Distributor Placement Opportunity Start",
		0x37: "Distributor Placement Opportunity End",
		0x38: "Provider Overlay Placement Opportunity Start",
		0x39: "Provider Overlay Placement Opportunity End",
		0x3A: "Distributor Overlay Placement Opportunity Start",
		0x3B: "Distributor Overlay Placement Opportunity End",
		0x40: "Unscheduled Event Start",
		0x41: "Unscheduled Event End",
		0x50: "Network Start",
		0x51: "Network End",
	}
	mesg, ok := table22[sd.SegmentationTypeID]
	if ok {
		return mesg
	}
	return "Unknown"
}

// Tag returns the splice_descriptor_tag.
func (sd *SegmentationDescriptor) Tag() uint32 {
	// ensure JSONType is set
	sd.JSONType = SegmentationDescriptorTag
	return SegmentationDescriptorTag
}

// DeliveryNotRestrictedFlag returns the delivery_not_restricted_flag.
func (sd *SegmentationDescriptor) DeliveryNotRestrictedFlag() bool {
	return sd.DeliveryRestrictions == nil
}

// ProgramSegmentationFlag returns the program_segmentation_flag.
func (sd *SegmentationDescriptor) ProgramSegmentationFlag() bool {
	return len(sd.Components) == 0
}

// SegmentationDurationFlag returns the segmentation_duration_flag.
func (sd *SegmentationDescriptor) SegmentationDurationFlag() bool {
	return sd.SegmentationDuration != nil
}

// SegmentationUpidLength return the segmentation_upid_length
func (sd *SegmentationDescriptor) SegmentationUpidLength() int {
	length := 0
	if len(sd.SegmentationUPIDs) == 1 {
		length += len(sd.SegmentationUPIDs[0].valueBytes()) * 8 // segmentation_upid() (bytes -> bits)
	} else if len(sd.SegmentationUPIDs) > 1 {
		// for MID, include type & length with each contained upid
		for _, upid := range sd.SegmentationUPIDs {
			length += 8                          // segmentation_upid_type
			length += 8                          // segmentation_upid_length
			length += len(upid.valueBytes()) * 8 // segmentation_upid (bytes -> bits)
		}
	}
	return length / 8
}

// decode updates this splice_descriptor from binary.
func (sd *SegmentationDescriptor) decode(b []byte) error {
	r := iobit.NewReader(b)
	r.Skip(8)  // splice_descriptor_tag
	r.Skip(8)  // descriptor_length
	r.Skip(32) // identifier
	sd.SegmentationEventID = r.Uint32(32)
	sd.SegmentationEventCancelIndicator = r.Bit()
	r.Skip(7) // reserved

	if !sd.SegmentationEventCancelIndicator {
		programSegmentationFlag := r.Bit()
		segmentationDurationFlag := r.Bit()
		deliveryNotRestrictedFlag := r.Bit()

		if !deliveryNotRestrictedFlag {
			sd.DeliveryRestrictions = &DeliveryRestrictions{}
			sd.DeliveryRestrictions.WebDeliveryAllowedFlag = r.Bit()
			sd.DeliveryRestrictions.NoRegionalBlackoutFlag = r.Bit()
			sd.DeliveryRestrictions.ArchiveAllowedFlag = r.Bit()
			sd.DeliveryRestrictions.DeviceRestrictions = r.Uint32(2)
		} else {
			r.Skip(5) // reserved
		}

		if !programSegmentationFlag {
			componentCount := int(r.Uint32(8))
			sd.Components = make([]SegmentationDescriptorComponent, componentCount)
			for i := 0; i < componentCount; i++ {
				c := SegmentationDescriptorComponent{}
				c.Tag = r.Uint32(8)
				r.Skip(7) // reserved
				c.PTSOffset = r.Uint64(33)
				sd.Components[i] = c
			}
		}

		if segmentationDurationFlag {
			dur := r.Uint64(40)
			sd.SegmentationDuration = &dur
		}

		segmentationUpidType := r.Uint32(8)
		segmentationUpidLength := int(r.Uint32(8))
		if segmentationUpidLength > 0 {
			segmentationUpidValue := r.Bytes(segmentationUpidLength)

			if segmentationUpidType == SegmentationUPIDTypeMID {
				upidr := iobit.NewReader(segmentationUpidValue)
				sd.SegmentationUPIDs = []SegmentationUPID{}
				for upidr.LeftBits() > 0 {
					upidType := upidr.Uint32(8)
					upidLength := int(upidr.Uint32(8))
					upidValue := upidr.Bytes(upidLength)
					sd.SegmentationUPIDs = append(
						sd.SegmentationUPIDs,
						NewSegmentationUPID(upidType, upidValue),
					)
				}
			} else {
				sd.SegmentationUPIDs = []SegmentationUPID{
					NewSegmentationUPID(segmentationUpidType, segmentationUpidValue),
				}
			}
		}

		sd.SegmentationTypeID = r.Uint32(8)
		sd.SegmentNum = r.Uint32(8)
		sd.SegmentsExpected = r.Uint32(8)

		// these fields are new in 2016 so we need a secondary check whether they were actually included
		// in the binary payload
		if sd.SegmentationTypeID == SegmentationTypeProviderPOStart || sd.SegmentationTypeID == SegmentationTypeDistributorPOStart {
			if r.LeftBits() == 16 {
				n := r.Uint32(8)
				e := r.Uint32(8)
				sd.SubSegmentNum = &n
				sd.SubSegmentsExpected = &e
			}
		}
	}

	if err := readerError(r); err != nil {
		return fmt.Errorf("segmentation_descriptor: %w", err)
	}
	return nil
}

// encode this splice_descriptor to binary.
func (sd *SegmentationDescriptor) encode() ([]byte, error) {
	length := sd.length()

	// add 2 bytes to contain splice_descriptor_tag & descriptor_length
	buf := make([]byte, length+2)
	iow := iobit.NewWriter(buf)
	iow.PutUint32(8, SegmentationDescriptorTag)
	iow.PutUint32(8, uint32(length))
	iow.PutUint32(32, CUEIdentifier)
	iow.PutUint32(32, sd.SegmentationEventID)
	iow.PutBit(sd.SegmentationEventCancelIndicator)
	iow.PutUint32(7, Reserved)

	if !sd.SegmentationEventCancelIndicator {
		iow.PutBit(sd.ProgramSegmentationFlag())
		iow.PutBit(sd.SegmentationDurationFlag())

		iow.PutBit(sd.DeliveryNotRestrictedFlag())
		if sd.DeliveryRestrictions != nil {
			iow.PutBit(sd.DeliveryRestrictions.WebDeliveryAllowedFlag)
			iow.PutBit(sd.DeliveryRestrictions.NoRegionalBlackoutFlag)
			iow.PutBit(sd.DeliveryRestrictions.ArchiveAllowedFlag)
			iow.PutUint32(2, sd.DeliveryRestrictions.DeviceRestrictions)
		} else {
			iow.PutUint32(5, Reserved)
		}

		if !sd.ProgramSegmentationFlag() {
			iow.PutUint32(8, uint32(len(sd.Components)))
			for _, c := range sd.Components {
				iow.PutUint32(8, c.Tag)
				iow.PutUint32(7, Reserved)
				iow.PutUint64(33, c.PTSOffset)
			}
		}

		if sd.SegmentationDurationFlag() {
			iow.PutUint64(40, *sd.SegmentationDuration)
		}

		if len(sd.SegmentationUPIDs) == 0 {
			iow.PutUint32(8, 0x00) // segmentation_upid_type
			iow.PutUint32(8, 0x00) // segmentation_upid_length
		} else if len(sd.SegmentationUPIDs) == 1 {
			vb := sd.SegmentationUPIDs[0].valueBytes()
			iow.PutUint32(8, sd.SegmentationUPIDs[0].Type)
			iow.PutUint32(8, uint32(len(vb)))
			_, _ = iow.Write(vb)
		} else {
			iow.PutUint32(8, SegmentationUPIDTypeMID)
			iow.PutUint32(8, uint32(sd.SegmentationUpidLength()))
			for _, upid := range sd.SegmentationUPIDs {
				vb := upid.valueBytes()
				iow.PutUint32(8, upid.Type)
				iow.PutUint32(8, uint32(len(vb)))
				_, _ = iow.Write(vb)
			}
		}

		iow.PutUint32(8, sd.SegmentationTypeID)
		iow.PutUint32(8, sd.SegmentNum)
		iow.PutUint32(8, sd.SegmentsExpected)

		if sd.SubSegmentNum != nil {
			iow.PutUint32(8, *sd.SubSegmentNum)
		}
		if sd.SubSegmentsExpected != nil {
			iow.PutUint32(8, *sd.SubSegmentsExpected)
		}
	}

	err := iow.Flush()
	return buf, err
}

// descriptorLength returns the descriptor_length
func (sd *SegmentationDescriptor) length() int {
	length := 32 // identifier
	length += 32 // segmentation_event_id
	length++     // segmentation_event_cancel_indicator
	length += 7  // reserved

	// if segmentation_event_cancel_indicator == 0
	if !sd.SegmentationEventCancelIndicator {
		length++    // program_segmentation_flag
		length++    // segmentation_duration_flag
		length++    // delivery_not_restricted_flag
		length += 5 // delivery restriction flags or reserved

		// if program_segmentation_flag == 0
		if !sd.ProgramSegmentationFlag() {
			length += 8 // component_count

			// for i=0 to component_count
			for range sd.Components {
				length += 8  // component_tag
				length += 7  // reserved
				length += 33 // pts_offset
			}
		}
		if sd.SegmentationDurationFlag() {
			length += 40 // segmentation_duration
		}
		length += 8                               // segmentation_upid_type
		length += 8                               // segmentation_upid_length
		length += sd.SegmentationUpidLength() * 8 // segmentation_upid() (bytes -> bits)
		length += 8                               // segmentation_type_id
		length += 8                               // segment_num
		length += 8                               // segments_expected

		if sd.SubSegmentNum != nil {
			length += 8 // sub_segment_num
		}
		if sd.SubSegmentsExpected != nil {
			length += 8 // sub_segments_expected
		}
	}

	return length / 8
}

// table returns the tabular description of this SegmentationDescriptor.
func (sd *SegmentationDescriptor) table(prefix, indent string) string {
	var b bytes.Buffer
	_, _ = fmt.Fprintf(&b, prefix+"avail_descriptor() {\n")
	_, _ = fmt.Fprintf(&b, prefix+indent+"splice_descriptor_tag: %#02x\n", sd.Tag())
	_, _ = fmt.Fprintf(&b, prefix+indent+"descriptor_length: %d bytes\n", sd.length())
	_, _ = fmt.Fprintf(&b, prefix+indent+"identifier: %s\n", CUEIASCII)
	_, _ = fmt.Fprintf(&b, prefix+indent+"segmentation_event_id: %d\n", sd.SegmentationEventID)
	_, _ = fmt.Fprintf(&b, prefix+indent+"segmentation_event_cancel_indicator: %#v\n", sd.SegmentationEventCancelIndicator)
	if !sd.SegmentationEventCancelIndicator {
		_, _ = fmt.Fprintf(&b, prefix+indent+"program_segmentation_flag: %#v\n", sd.ProgramSegmentationFlag())
		_, _ = fmt.Fprintf(&b, prefix+indent+"segmentation_duration_flag: %#v\n", sd.SegmentationDurationFlag())
		_, _ = fmt.Fprintf(&b, prefix+indent+"delivery_not_restricted_flag: %#v\n", sd.DeliveryNotRestrictedFlag())
		if sd.DeliveryRestrictions != nil {
			_, _ = fmt.Fprintf(&b, prefix+indent+"web_delivery_allowed_flag: %#v\n", sd.DeliveryRestrictions.WebDeliveryAllowedFlag)
			_, _ = fmt.Fprintf(&b, prefix+indent+"no_regional_blackout_flag: %#v\n", sd.DeliveryRestrictions.NoRegionalBlackoutFlag)
			_, _ = fmt.Fprintf(&b, prefix+indent+"archive_allowed_flag: %#v\n", sd.DeliveryRestrictions.ArchiveAllowedFlag)
			_, _ = fmt.Fprintf(&b, prefix+indent+"device_restrictions: %s\n", sd.DeliveryRestrictions.deviceRestrictionsName())
		}
		if len(sd.Components) > 0 {
			_, _ = fmt.Fprintf(&b, prefix+indent+"component_count: %d\n", len(sd.Components))
			for i, c := range sd.Components {
				_, _ = fmt.Fprintf(&b, prefix+indent+"component[%d] {\n", i)
				_, _ = fmt.Fprintf(&b, prefix+indent+indent+"component_tag %d\n", c.Tag)
				_, _ = fmt.Fprintf(&b, prefix+indent+indent+"pts_offset %d ticks (%s)\n", c.PTSOffset, TicksToDuration(c.PTSOffset))
				_, _ = fmt.Fprintf(&b, prefix+indent+"}\n")
			}
		}
		if sd.SegmentationDurationFlag() {
			_, _ = fmt.Fprintf(&b, prefix+indent+"segmentation_duration: %d ticks (%s)\n", *sd.SegmentationDuration, TicksToDuration(*sd.SegmentationDuration))
		}

		_, _ = fmt.Fprintf(&b, prefix+indent+"segmentation_upid_length: %d bytes\n", sd.SegmentationUpidLength())
		for i, u := range sd.SegmentationUPIDs {
			_, _ = fmt.Fprintf(&b, prefix+indent+"segmentation_upid[%d] {\n", i)
			_, _ = fmt.Fprintf(&b, prefix+indent+indent+"segmentation_upid_type: %s (%#02x)\n", u.Name(), u.Type)
			if u.Type == SegmentationUPIDTypeMPU {
				_, _ = fmt.Fprintf(&b, prefix+indent+indent+"format_identifier: %s\n", u.formatIdentifierString())
			}
			if u.Format == "text" {
				_, _ = fmt.Fprintf(&b, prefix+indent+indent+"segmentation_upid: %s\n", u.Value)
			} else {
				_, _ = fmt.Fprintf(&b, prefix+indent+indent+"segmentation_upid: %#x\n", u.valueBytes())
			}
			_, _ = fmt.Fprintf(&b, prefix+indent+"}\n")
		}
	}

	_, _ = fmt.Fprintf(&b, prefix+indent+"segmentation_type_id: %s (%#02x)\n", sd.Name(), sd.SegmentationTypeID)
	_, _ = fmt.Fprintf(&b, prefix+indent+"segment_num: %d\n", sd.SegmentNum)
	_, _ = fmt.Fprintf(&b, prefix+indent+"segments_expected: %d\n", sd.SegmentsExpected)
	if sd.SubSegmentNum != nil {
		_, _ = fmt.Fprintf(&b, prefix+indent+"sub_segment_num: %d\n", *sd.SubSegmentNum)
	}
	if sd.SubSegmentsExpected != nil {
		_, _ = fmt.Fprintf(&b, prefix+indent+"sub_segments_expected: %d\n", *sd.SubSegmentsExpected)
	}
	_, _ = fmt.Fprintf(&b, prefix+"}\n")

	return b.String()
}

// SegmentationDescriptorComponent describes the Component element contained
// within the SegmentationDescriptorType XML schema definition.
type SegmentationDescriptorComponent struct {
	Tag       uint32 `xml:"componentTag,attr" json:"componentTag"`
	PTSOffset uint64 `xml:"ptsOffset,attr" json:"ptsOffset"`
}
