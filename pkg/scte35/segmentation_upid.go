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
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bamiaux/iobit"
)

const (
	// SegmentationUPIDTypeNotUsed is the segmentation_upid_type for Not Used.
	SegmentationUPIDTypeNotUsed = 0x00
	// SegmentationUPIDTypeUserDefined is the segmentation_upid_type for User
	// Defined.
	SegmentationUPIDTypeUserDefined = 0x01
	// SegmentationUPIDTypeISCI is the segmentation_upid_type for ISCI
	SegmentationUPIDTypeISCI = 0x02
	// SegmentationUPIDTypeAdID is the segmentation_upid_type for Ad-ID
	SegmentationUPIDTypeAdID = 0x03
	// SegmentationUPIDTypeUMID is the segmentation_upid_type for UMID
	SegmentationUPIDTypeUMID = 0x04
	// SegmentationUPIDTypeISANDeprecated is the segmentation_upid_type for
	// ISAN Deprecated.
	SegmentationUPIDTypeISANDeprecated = 0x05
	// SegmentationUPIDTypeISAN is the segmentation_upid_type for ISAN.
	SegmentationUPIDTypeISAN = 0x06
	// SegmentationUPIDTypeTID is the segmentation_upid_type for TID.
	SegmentationUPIDTypeTID = 0x07
	// SegmentationUPIDTypeTI is the segmentation_upid_type for TI.
	SegmentationUPIDTypeTI = 0x08
	// SegmentationUPIDTypeADI is the segmentation_upid_type for ADI.
	SegmentationUPIDTypeADI = 0x09
	// SegmentationUPIDTypeEIDR is the segmentation_upid_type for EIDR.
	SegmentationUPIDTypeEIDR = 0x0a
	// SegmentationUPIDTypeATSC is the segmentation_upid_type for ATSC Content
	// Identifier.
	SegmentationUPIDTypeATSC = 0x0b
	// SegmentationUPIDTypeMPU is the segmentation_upid_type for MPU().
	SegmentationUPIDTypeMPU = 0x0c
	// SegmentationUPIDTypeMID is the segmentation_upid_type for MID().
	SegmentationUPIDTypeMID = 0x0d
	// SegmentationUPIDTypeADS is the segmentation_upid_type for ADS Information.
	SegmentationUPIDTypeADS = 0x0e
	// SegmentationUPIDTypeURI is the segmentation_upid_type for URI.
	SegmentationUPIDTypeURI = 0x0f
	// SegmentationUPIDTypeUUID is the segmentation_upid_type for UUID.
	SegmentationUPIDTypeUUID = 0x10
)

// NewSegmentationUPID construct a new SegmentationUPID
func NewSegmentationUPID(upidType uint32, buf []byte) SegmentationUPID {
	r := iobit.NewReader(buf)

	switch upidType {
	case SegmentationUPIDTypeEIDR:
		return SegmentationUPID{
			Type:   upidType,
			Format: "text",
			Value:  canonicalEIDR(r.LeftBytes()),
		}
	case SegmentationUPIDTypeISAN, SegmentationUPIDTypeISANDeprecated:
		return SegmentationUPID{
			Type:   upidType,
			Format: "base-64",
			Value:  base64.StdEncoding.EncodeToString(r.LeftBytes()),
		}
	case SegmentationUPIDTypeMPU:
		fi := r.Uint32(32)
		return SegmentationUPID{
			Type:             upidType,
			Format:           "base-64",
			FormatIdentifier: &fi,
			Value:            base64.StdEncoding.EncodeToString(r.LeftBytes()),
		}
	case SegmentationUPIDTypeTI:
		return SegmentationUPID{
			Type:   upidType,
			Format: "text",
			Value:  fmt.Sprintf("%d", r.Uint64(r.LeftBits())),
		}
	default:
		return SegmentationUPID{
			Type:   upidType,
			Format: "text",
			Value:  string(r.LeftBytes()),
		}
	}
}

// SegmentationUPID is used to express a UPID in an XML document.
type SegmentationUPID struct {
	Type             uint32  `xml:"segmentationUpidType,attr" json:"segmentationUpidType"`
	FormatIdentifier *uint32 `xml:"formatIdentifier,attr" json:"formatIdentifier"`
	Format           string  `xml:"format,attr" json:"format"`
	Value            string  `xml:",chardata" json:"value"`
}

// Name returns the name for the segmentation_upid_type.
func (upid *SegmentationUPID) Name() string {
	switch upid.Type {
	case SegmentationUPIDTypeNotUsed:
		return "Not Used"
	case SegmentationUPIDTypeUserDefined:
		return "User Defined"
	case SegmentationUPIDTypeISCI:
		return "ISCI"
	case SegmentationUPIDTypeAdID:
		return "Ad-ID"
	case SegmentationUPIDTypeUMID:
		return "UMID"
	case SegmentationUPIDTypeISANDeprecated:
		return "ISAN (Deprecated)"
	case SegmentationUPIDTypeISAN:
		return "ISAN"
	case SegmentationUPIDTypeTID:
		return "TID"
	case SegmentationUPIDTypeTI:
		return "TI"
	case SegmentationUPIDTypeADI:
		return "ADI"
	case SegmentationUPIDTypeEIDR:
		return "EIDR: " + upid.eidrTypeName()
	case SegmentationUPIDTypeATSC:
		return "ATSC Content Identifier"
	case SegmentationUPIDTypeMPU:
		return "MPU()"
	case SegmentationUPIDTypeMID:
		return "MID()"
	case SegmentationUPIDTypeADS:
		return "ADS Information"
	case SegmentationUPIDTypeURI:
		return "URI"
	case SegmentationUPIDTypeUUID:
		return "UUID"
	default:
		return "Unknown"
	}
}

// ASCIIValue returns the UPID value as an ASCII string. Bytes outside ASCII
// range are represented by a dot (.).
func (upid *SegmentationUPID) ASCIIValue() string {
	b := upid.valueBytes()
	var stringsValues []string
	for _, b := range b {
		var stringVal string
		if utf8.Valid([]byte{b}) {
			stringVal = string(b)
		} else {
			// non valid ascii byte element, it is represented as a dot
			stringVal = "."
		}
		stringsValues = append(stringsValues, stringVal)
	}
	return strings.Join(stringsValues, "")
}

// compressEIRD returns a compressed EIDR.
func (upid *SegmentationUPID) compressEIDR(s string) []byte {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '.' || r == '/'
	})

	if len(parts) != 3 {
		Logger.Printf("non-canonical EIDR string: %s", s)
		return []byte(s)
	}

	i, err := strconv.Atoi(parts[1])
	if err != nil {
		Logger.Printf("non-canonical EIDR string: %s", s)
		return []byte(s)
	}

	b := make([]byte, 12)
	iow := iobit.NewWriter(b)
	iow.PutUint32(16, uint32(i))

	h, err := hex.DecodeString(strings.ReplaceAll(parts[2], "-", ""))
	if err != nil {
		Logger.Printf("non-canonical EIDR string: %s", s)
		return []byte(s)
	}

	_, _ = iow.Write(h)
	_ = iow.Flush()

	return b
}

// eidrTypeName returns the EIDR type name.
func (upid *SegmentationUPID) eidrTypeName() string {
	if strings.HasPrefix(upid.Value, "10.5237") {
		return "Party ID"
	}
	if strings.HasPrefix(upid.Value, "10.5238") {
		return "User ID"
	}
	if strings.HasPrefix(upid.Value, "10.5239") {
		return "Service ID"
	}
	if strings.HasPrefix(upid.Value, "10.5240") {
		return "Content ID"
	}
	return ""
}

// formatIdentifierString returns the format identifier as a string
func (upid *SegmentationUPID) formatIdentifierString() string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, *upid.FormatIdentifier)
	return string(b)
}

// valueBytes returns the value as a byte array.
func (upid *SegmentationUPID) valueBytes() []byte {
	switch upid.Type {
	case SegmentationUPIDTypeMPU:
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, *upid.FormatIdentifier)
		v, _ := base64.StdEncoding.DecodeString(upid.Value)
		b = append(b, v...)
		return b
	case SegmentationUPIDTypeEIDR:
		return upid.compressEIDR(upid.Value)
	case SegmentationUPIDTypeTI:
		b := make([]byte, 8)
		if i, err := strconv.Atoi(upid.Value); err == nil {
			binary.BigEndian.PutUint64(b, uint64(i))
		}
		return b
	default:
		switch upid.Format {
		case "hexbinary":
			b, _ := hex.DecodeString(upid.Value)
			return b
		case "base-64":
			b, _ := base64.StdEncoding.DecodeString(upid.Value)
			return b
		default:
			return []byte(upid.Value)
		}
	}
}

// canonicalEIDR returns a canonical EIDR.
func canonicalEIDR(b []byte) string {
	// already canonical
	if bytes.Contains(b, []byte("/")) {
		return string(b)
	}

	// dunno what this is
	if len(b) != 12 {
		Logger.Printf("unexpected eidr value received: %s", b)
		return ""
	}

	i := int(binary.BigEndian.Uint16(b[:2]))
	return fmt.Sprintf("10.%d/%X-%X-%X-%X-%X", i, b[2:4], b[4:6], b[6:8], b[8:10], b[10:12])
}
