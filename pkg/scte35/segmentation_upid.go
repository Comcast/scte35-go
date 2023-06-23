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
	"io/ioutil"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/bamiaux/iobit"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
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
	// SegmentationUPIDTypeSCR is the segmentation_upid_type for SCR.
	SegmentationUPIDTypeSCR = 0x11
)

// NewSegmentationUPID construct a new SegmentationUPID
func NewSegmentationUPID(upidType uint32, buf []byte) SegmentationUPID {
	r := iobit.NewReader(buf)

	switch upidType {
	// EIDR - custom
	case SegmentationUPIDTypeEIDR:
		return SegmentationUPID{
			Type:  upidType,
			Value: canonicalEIDR(r.LeftBytes()),
		}
	// ISAN - base64
	case SegmentationUPIDTypeISAN, SegmentationUPIDTypeISANDeprecated:
		return SegmentationUPID{
			Type:  upidType,
			Value: base64.StdEncoding.EncodeToString(r.LeftBytes()),
		}
	// MPU - custom
	case SegmentationUPIDTypeMPU:
		fi := r.Uint32(32)
		return SegmentationUPID{
			Type:             upidType,
			FormatIdentifier: &fi,
			Value:            base64.StdEncoding.EncodeToString(r.LeftBytes()),
		}
	// TI - unsigned int
	case SegmentationUPIDTypeTI:
		return SegmentationUPID{
			Type:  upidType,
			Value: strconv.FormatUint(r.Uint64(r.LeftBits()), 10),
		}
	// everything else - plain text
	default:
		return SegmentationUPID{
			Type:  upidType,
			Value: DecodeASCII(r.LeftBytes()),
		}
	}
}

// DecodeASCII assumes the byte array holds an ASCII-encoded string and converts it to a UTF-8
// compatible string.
func DecodeASCII(b []byte) string {
	// This shouldn't fail...
	e, _ := ianaindex.MIME.Encoding("US-ASCII")

	r := transform.NewReader(bytes.NewReader(b), e.NewDecoder())
	result, err := ioutil.ReadAll(r)
	if err != nil {
		// Note: NewSegmentationUPID, which calls this function, doesn't return errors.
		// Not sure how best to account for this?
		return "invalid encoding"
	}

	return string(result)
}

// SegmentationUPID is used to express a UPID in an XML document.
type SegmentationUPID struct {
	Type             uint32  `xml:"segmentationUpidType,attr" json:"segmentationUpidType"`
	FormatIdentifier *uint32 `xml:"formatIdentifier,attr,omitempty" json:"formatIdentifier,omitempty"`
	Value            string  `xml:",chardata" json:"value"`
	// Deprecated: no longer used and will be removed in a future release
	Format string `xml:"-" json:"-"`
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
	case SegmentationUPIDTypeSCR:
		return "SCR"
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
		Logger.Printf("EIDR string contains too many parts: %s", s)
		return []byte(s)
	}

	i, err := strconv.Atoi(parts[1])
	if err != nil {
		Logger.Printf("Non-canonical EIDR prefix: '%s'", s)
		return []byte(s)
	}

	b := make([]byte, 12)
	iow := iobit.NewWriter(b)
	iow.PutUint32(16, uint32(i))

	h, err := hex.DecodeString(strings.ReplaceAll(parts[2], "-", ""))
	if err != nil {
		Logger.Printf("Non-canonical EIDR suffix: '%s'", s)
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
	upid.Value = strings.TrimSpace(upid.Value)

	// this switch should align with the constructor above
	switch upid.Type {
	// EIDR - custom
	case SegmentationUPIDTypeEIDR:
		return upid.compressEIDR(upid.Value)
	// ISAN - base64
	case SegmentationUPIDTypeISAN, SegmentationUPIDTypeISANDeprecated:
		b, err := base64.StdEncoding.DecodeString(upid.Value)
		if err != nil {
			Logger.Fatalf("Error parsing UPID value: %s", err)
			return b
		}
		return b
	// MPU - custom
	case SegmentationUPIDTypeMPU:
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, *upid.FormatIdentifier)
		v, err := base64.StdEncoding.DecodeString(upid.Value)
		if err != nil {
			Logger.Fatalf("Error parsing UPID value: %s", err)
			return b
		}
		b = append(b, v...)
		return b
	// TI - unsigned int
	case SegmentationUPIDTypeTI:
		b := make([]byte, 8)
		i, err := strconv.ParseUint(strings.TrimSpace(upid.Value), 10, 64)
		if err != nil {
			Logger.Fatalf("Error parsing UPID value: %s", err)
			return b
		}
		binary.BigEndian.PutUint64(b, i)
		return b
	// everything else - plain text
	default:
		return []byte(upid.Value)
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
		Logger.Printf("Unexpected eidr value received: %s", b)
		return ""
	}

	i := int(binary.BigEndian.Uint16(b[:2]))
	return fmt.Sprintf("10.%d/%X-%X-%X-%X-%X", i, b[2:4], b[4:6], b[6:8], b[8:10], b[10:12])
}
