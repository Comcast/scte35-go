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

package scte35_test

import (
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Comcast/scte35-go/pkg/scte35"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	scte35.Logger.SetOutput(os.Stderr)
	os.Exit(m.Run())
}

// combined Testcases for decodeBase64, decodeHex and decodeBytes
//
// when adding tests that contain multiple splice descriptors, care must be
// taken to ensure they are in the order specified in the custom UnmarshalXML
// implementation, otherwise misleading error may occur
var commonDecodeTestcases = map[string]struct {
	base64   string
	hex      string
	bytes    []byte
	err      error
	expected scte35.SpliceInfoSection
	legacy   bool
}{
	"Sample 14.1 time_signal - Placement Opportunity Start": {
		base64: "/DA0AAAAAAAA///wBQb+cr0AUAAeAhxDVUVJSAAAjn/PAAGlmbAICAAAAAAsoKGKNAIAmsnRfg==",
		hex:    "0xFC3034000000000000FFFFF00506FE72BD0050001E021C435545494800008E7FCF0001A599B00808000000002CA0A18A3402009AC9D17E",
		bytes: []byte{
			0xFC, 0x30, 0x34, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xF0, 0x05, 0x06, 0xFE, 0x72,
			0xBD, 0x00, 0x50, 0x00, 0x1E, 0x02, 0x1C, 0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00, 0x8E, 0x7F,
			0xCF, 0x00, 0x01, 0xA5, 0x99, 0xB0, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x2C, 0xA0, 0xA1, 0x8A,
			0x34, 0x02, 0x00, 0x9A, 0xC9, 0xD1, 0x7E,
		},

		expected: scte35.SpliceInfoSection{
			EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
			SpliceCommand: &scte35.TimeSignal{
				SpliceTime: scte35.SpliceTime{
					PTSTime: ptr(uint64(0x072bd0050)),
				},
			},
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
					},
					SegmentationEventID:  uint32(0x4800008e),
					SegmentationTypeID:   scte35.SegmentationTypeProviderPOStart,
					SegmentationDuration: ptr(uint64(0x0001a599b0)),
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002ca0a18a)),
					},
					SegmentNum: 2,
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Sample 14.2 splice_insert": {
		base64: "/DAvAAAAAAAA///wFAVIAACPf+/+c2nALv4AUsz1AAAAAAAKAAhDVUVJAAABNWLbowo=",
		hex:    "0xFC302F000000000000FFFFF014054800008F7FEFFE7369C02EFE0052CCF500000000000A0008435545490000013562DBA30A",
		bytes: []byte{
			0xFC, 0x30, 0x2F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xF0, 0x14, 0x05, 0x48, 0x00,
			0x00, 0x8F, 0x7F, 0xEF, 0xFE, 0x73, 0x69, 0xC0, 0x2E, 0xFE, 0x00, 0x52, 0xCC, 0xF5, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x0A, 0x00, 0x08, 0x43, 0x55, 0x45, 0x49, 0x00, 0x00, 0x01, 0x35, 0x62, 0xDB,
			0xA3, 0x0A,
		},
		expected: scte35.SpliceInfoSection{
			EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
			SpliceCommand: &scte35.SpliceInsert{
				BreakDuration: &scte35.BreakDuration{
					AutoReturn: true,
					Duration:   uint64(0x00052ccf5),
				},
				SpliceEventID:         uint32(0x4800008f),
				OutOfNetworkIndicator: true,
				Program: &scte35.SpliceInsertProgram{
					SpliceTime: scte35.SpliceTime{
						PTSTime: ptr(uint64(0x07369c02e)),
					},
				},
			},
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.AvailDescriptor{
					ProviderAvailID: 0x00000135,
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Sample 14.3 time_signal - Placement Opportunity End": {
		base64: "/DAvAAAAAAAA///wBQb+dGKQoAAZAhdDVUVJSAAAjn+fCAgAAAAALKChijUCAKnMZ1g=",
		hex:    "0xFC302F000000000000FFFFF00506FE746290A000190217435545494800008E7F9F0808000000002CA0A18A350200A9CC6758",
		bytes: []byte{
			0xFC, 0x30, 0x2F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xF0, 0x05, 0x06, 0xFE, 0x74,
			0x62, 0x90, 0xA0, 0x00, 0x19, 0x02, 0x17, 0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00, 0x8E, 0x7F,
			0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x2C, 0xA0, 0xA1, 0x8A, 0x35, 0x02, 0x00, 0xA9, 0xCC,
			0x67, 0x58,
		},
		expected: scte35.SpliceInfoSection{
			EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
			SpliceCommand: &scte35.TimeSignal{
				SpliceTime: scte35.SpliceTime{
					PTSTime: ptr(uint64(0x0746290a0)),
				},
			},
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
						WebDeliveryAllowedFlag: true,
					},
					SegmentationEventID: 0x4800008e,
					SegmentationTypeID:  scte35.SegmentationTypeProviderPOEnd,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002ca0a18a)),
					},
					SegmentNum: 2,
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Sample 14.4 time_signal - Program Start/End": {
		base64: "/DBIAAAAAAAA///wBQb+ek2ItgAyAhdDVUVJSAAAGH+fCAgAAAAALMvDRBEAAAIXQ1VFSUgAABl/nwgIAAAAACyk26AQAACZcuND",
		hex:    "0xFC3048000000000000FFFFF00506FE7A4D88B60032021743554549480000187F9F0808000000002CCBC344110000021743554549480000197F9F0808000000002CA4DBA01000009972E343",
		bytes: []byte{
			0xFC, 0x30, 0x48, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xF0, 0x05, 0x06, 0xFE, 0x7A,
			0x4D, 0x88, 0xB6, 0x00, 0x32, 0x02, 0x17, 0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00, 0x18, 0x7F,
			0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x2C, 0xCB, 0xC3, 0x44, 0x11, 0x00, 0x00, 0x02, 0x17,
			0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00, 0x19, 0x7F, 0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00,
			0x2C, 0xA4, 0xDB, 0xA0, 0x10, 0x00, 0x00, 0x99, 0x72, 0xE3, 0x43,
		},
		expected: scte35.SpliceInfoSection{
			EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
			SpliceCommand: &scte35.TimeSignal{
				SpliceTime: scte35.SpliceTime{
					PTSTime: ptr(uint64(0x07a4d88b6)),
				},
			},
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
						WebDeliveryAllowedFlag: true,
					},
					SegmentationEventID: 0x48000018,
					SegmentationTypeID:  scte35.SegmentationTypeProgramEnd,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002ccbc344)),
					},
				},
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
						WebDeliveryAllowedFlag: true,
					},
					SegmentationEventID: 0x48000019,
					SegmentationTypeID:  scte35.SegmentationTypeProgramStart,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002ca4dba0)),
					},
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Sample 14.5 time_signal - Program Overlap Start": {
		base64: "/DAvAAAAAAAA///wBQb+rr//ZAAZAhdDVUVJSAAACH+fCAgAAAAALKVs9RcAAJUdsKg=",
		hex:    "0xFC302F000000000000FFFFF00506FEAEBFFF640019021743554549480000087F9F0808000000002CA56CF5170000951DB0A8",
		bytes: []byte{
			0xFC, 0x30, 0x2F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xF0, 0x05, 0x06, 0xFE, 0xAE,
			0xBF, 0xFF, 0x64, 0x00, 0x19, 0x02, 0x17, 0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00, 0x08, 0x7F,
			0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x2C, 0xA5, 0x6C, 0xF5, 0x17, 0x00, 0x00, 0x95, 0x1D,
			0xB0, 0xA8,
		},
		expected: scte35.SpliceInfoSection{
			EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
			SpliceCommand: &scte35.TimeSignal{
				SpliceTime: scte35.SpliceTime{
					PTSTime: ptr(uint64(0x0aebfff64)),
				},
			},
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
						WebDeliveryAllowedFlag: true,
					},
					SegmentationEventID: 0x48000008,
					SegmentationTypeID:  scte35.SegmentationTypeProgramOverlapStart,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002ca56cf5)),
					},
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Sample 14.6 time_signal - Program Blackout Override / Program End": {
		base64: "/DBIAAAAAAAA///wBQb+ky44CwAyAhdDVUVJSAAACn+fCAgAAAAALKCh4xgAAAIXQ1VFSUgAAAl/nwgIAAAAACygoYoRAAC0IX6w",
		hex:    "0xFC3048000000000000FFFFF00506FE932E380B00320217435545494800000A7F9F0808000000002CA0A1E3180000021743554549480000097F9F0808000000002CA0A18A110000B4217EB0",
		bytes: []byte{
			0xFC, 0x30, 0x48, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xF0, 0x05, 0x06, 0xFE, 0x93,
			0x2E, 0x38, 0x0B, 0x00, 0x32, 0x02, 0x17, 0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00, 0x0A, 0x7F,
			0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x2C, 0xA0, 0xA1, 0xE3, 0x18, 0x00, 0x00, 0x02, 0x17,
			0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00, 0x09, 0x7F, 0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00,
			0x2C, 0xA0, 0xA1, 0x8A, 0x11, 0x00, 0x00, 0xB4, 0x21, 0x7E, 0xB0,
		},
		expected: scte35.SpliceInfoSection{
			EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
			SpliceCommand: &scte35.TimeSignal{
				SpliceTime: scte35.SpliceTime{
					PTSTime: ptr(uint64(0x0932e380b)),
				},
			},
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
						WebDeliveryAllowedFlag: true,
					},
					SegmentationEventID: 0x4800000a,
					SegmentationTypeID:  scte35.SegmentationTypeProgramBlackoutOverride,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002ca0a1e3)),
					},
				},
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
						WebDeliveryAllowedFlag: true,
					},
					SegmentationEventID: 0x48000009,
					SegmentationTypeID:  scte35.SegmentationTypeProgramEnd,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002ca0a18a)),
					},
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Sample 14.7 time_signal - Program End": {
		base64: "/DAvAAAAAAAA///wBQb+rvF8TAAZAhdDVUVJSAAAB3+fCAgAAAAALKVslxEAAMSHai4=",
		hex:    "0xFC302F000000000000FFFFF00506FEAEF17C4C0019021743554549480000077F9F0808000000002CA56C97110000C4876A2E",
		bytes: []byte{
			0xFC, 0x30, 0x2F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xF0, 0x05, 0x06, 0xFE, 0xAE,
			0xF1, 0x7C, 0x4C, 0x00, 0x19, 0x02, 0x17, 0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00, 0x07, 0x7F,
			0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x2C, 0xA5, 0x6C, 0x97, 0x11, 0x00, 0x00, 0xC4, 0x87,
			0x6A, 0x2E,
		},
		expected: scte35.SpliceInfoSection{
			EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
			SpliceCommand: &scte35.TimeSignal{
				SpliceTime: scte35.SpliceTime{
					PTSTime: ptr(uint64(0x0aef17c4c)),
				},
			},
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
						WebDeliveryAllowedFlag: true,
					},
					SegmentationEventID: 0x48000007,
					SegmentationTypeID:  scte35.SegmentationTypeProgramEnd,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002ca56c97)),
					},
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Sample 14.8 time_signal - Program Start/End - Placement Opportunity End": {
		base64: "/DBhAAAAAAAA///wBQb+qM1E7QBLAhdDVUVJSAAArX+fCAgAAAAALLLXnTUCAAIXQ1VFSUgAACZ/nwgIAAAAACyy150RAAACF0NVRUlIAAAnf58ICAAAAAAsstezEAAAihiGnw==",
		hex:    "0xFC3061000000000000FFFFF00506FEA8CD44ED004B021743554549480000AD7F9F0808000000002CB2D79D350200021743554549480000267F9F0808000000002CB2D79D110000021743554549480000277F9F0808000000002CB2D7B31000008A18869F",
		bytes: []byte{
			0xFC, 0x30, 0x61, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xF0, 0x05, 0x06, 0xFE, 0xA8,
			0xCD, 0x44, 0xED, 0x00, 0x4B, 0x02, 0x17, 0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00, 0xAD, 0x7F,
			0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x2C, 0xB2, 0xD7, 0x9D, 0x35, 0x02, 0x00, 0x02, 0x17,
			0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00, 0x26, 0x7F, 0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00,
			0x2C, 0xB2, 0xD7, 0x9D, 0x11, 0x00, 0x00, 0x02, 0x17, 0x43, 0x55, 0x45, 0x49, 0x48, 0x00, 0x00,
			0x27, 0x7F, 0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x2C, 0xB2, 0xD7, 0xB3, 0x10, 0x00, 0x00,
			0x8A, 0x18, 0x86, 0x9F,
		},
		expected: scte35.SpliceInfoSection{
			EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
			SpliceCommand: &scte35.TimeSignal{
				SpliceTime: scte35.SpliceTime{
					PTSTime: ptr(uint64(0x0a8cd44ed)),
				},
			},
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
						WebDeliveryAllowedFlag: true,
					},
					SegmentationEventID: 0x480000ad,
					SegmentationTypeID:  scte35.SegmentationTypeProviderPOEnd,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002cb2d79d)),
					},
					SegmentNum: 2,
				},
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
						WebDeliveryAllowedFlag: true,
					},
					SegmentationEventID: 0x48000026,
					SegmentationTypeID:  scte35.SegmentationTypeProgramEnd,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002cb2d79d)),
					},
				},
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						NoRegionalBlackoutFlag: true,
						ArchiveAllowedFlag:     true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
						WebDeliveryAllowedFlag: true,
					},
					SegmentationEventID: 0x48000027,
					SegmentationTypeID:  scte35.SegmentationTypeProgramStart,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, toBytes(0x000000002cb2d7b3)),
					},
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"SpliceInsert With DTMF": {
		base64: "/DAxAAAAAAAAAP/wFAVAAIeuf+/+0AWRK/4AUmXAAC0AfwAMAQpDVUVJUJ81MTkqo5/+gA==",
		hex:    "0xFC303100000000000000FFF01405400087AE7FEFFED005912BFE005265C0002D007F000C010A43554549509F3531392AA39FFE80",
		bytes: []byte{
			0xFC, 0x30, 0x31, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xF0, 0x14, 0x05, 0x40, 0x00,
			0x87, 0xAE, 0x7F, 0xEF, 0xFE, 0xD0, 0x05, 0x91, 0x2B, 0xFE, 0x00, 0x52, 0x65, 0xC0, 0x00, 0x2D,
			0x00, 0x7F, 0x00, 0x0C, 0x01, 0x0A, 0x43, 0x55, 0x45, 0x49, 0x50, 0x9F, 0x35, 0x31, 0x39, 0x2A,
			0xA3, 0x9F, 0xFE, 0x80,
		},
		expected: scte35.SpliceInfoSection{
			SpliceCommand: &scte35.SpliceInsert{
				BreakDuration:              &scte35.BreakDuration{AutoReturn: true, Duration: 5400000},
				Program:                    scte35.NewSpliceInsertProgram(3490025771),
				SpliceEventID:              1073776558,
				SpliceEventCancelIndicator: false,
				SpliceImmediateFlag:        false,
				OutOfNetworkIndicator:      true,
				UniqueProgramID:            45,
				AvailNum:                   0,
				AvailsExpected:             127,
			},
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.DTMFDescriptor{
					Preroll:   80,
					DTMFChars: "519*",
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Time Signal with Segmentation Descriptors": {
		base64: "/DBIAAAAAAAAAP/wBQb/tB67hgAyAhdDVUVJQAABEn+fCAgAAAAALzE8BTUAAAIXQ1VFSUAAAEV/nwgIAAAAAC8xPN4jAAAfiOPE",
		hex:    "0xFC304800000000000000FFF00506FFB41EBB860032021743554549400001127F9F0808000000002F313C05350000021743554549400000457F9F0808000000002F313CDE2300001F88E3C4",
		bytes: []byte{
			0xFC, 0x30, 0x48, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xF0, 0x05, 0x06, 0xFF, 0xB4,
			0x1E, 0xBB, 0x86, 0x00, 0x32, 0x02, 0x17, 0x43, 0x55, 0x45, 0x49, 0x40, 0x00, 0x01, 0x12, 0x7F,
			0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x2F, 0x31, 0x3C, 0x05, 0x35, 0x00, 0x00, 0x02, 0x17,
			0x43, 0x55, 0x45, 0x49, 0x40, 0x00, 0x00, 0x45, 0x7F, 0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00,
			0x2F, 0x31, 0x3C, 0xDE, 0x23, 0x00, 0x00, 0x1F, 0x88, 0xE3, 0xC4,
		},
		expected: scte35.SpliceInfoSection{
			SpliceCommand: scte35.NewTimeSignal(7316880262),
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						ArchiveAllowedFlag:     true,
						WebDeliveryAllowedFlag: true,
						NoRegionalBlackoutFlag: true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
					},
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{
							Type:   8,
							Format: scte35.SegmentationUPIDFormatText,
							Value:  "791755781",
						},
					},
					SegmentationTypeID:  53,
					SegmentationEventID: 1073742098,
				},
				&scte35.SegmentationDescriptor{
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						ArchiveAllowedFlag:     true,
						WebDeliveryAllowedFlag: true,
						NoRegionalBlackoutFlag: true,
						DeviceRestrictions:     scte35.DeviceRestrictionsNone,
					},
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{
							Type:   8,
							Format: scte35.SegmentationUPIDFormatText,
							Value:  "791755998",
						},
					},
					SegmentationTypeID:  35,
					SegmentationEventID: 1073741893,
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Empty Data": {
		base64: "",
		hex:    "",
		bytes:  []byte{},
		err:    scte35.ErrBufferOverflow,
	},
	"Invalid Encoding": {
		base64: "/DBaf%^",
		hex:    "0xBROKEN",
		err:    scte35.ErrUnsupportedEncoding,
	},
	"Splice Insert with Avail Descriptor": {
		base64: "/DAqAAAAAAAAAP/wDwUAAHn+f8/+QubGOQAAAAAACgAIQ1VFSQAAAADizteX",
		hex:    "0xFC302A00000000000000FFF00F05000079FE7FCFFE42E6C63900000000000A00084355454900000000E2CED797",
		bytes: []byte{
			0xFC, 0x30, 0x2A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xF0, 0x0F, 0x05, 0x00, 0x00,
			0x79, 0xFE, 0x7F, 0xCF, 0xFE, 0x42, 0xE6, 0xC6, 0x39, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x00,
			0x08, 0x43, 0x55, 0x45, 0x49, 0x00, 0x00, 0x00, 0x00, 0xE2, 0xCE, 0xD7, 0x97,
		},
		expected: scte35.SpliceInfoSection{
			SpliceCommand: &scte35.SpliceInsert{
				Program:               scte35.NewSpliceInsertProgram(1122420281),
				SpliceEventID:         31230,
				OutOfNetworkIndicator: true,
			},
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.AvailDescriptor{
					ProviderAvailID: 0,
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Multiple SegmentationUPIDs": {
		base64: "/DBrAAAAAAAAAP/wBQb/AAAAAABVAlNDVUVJAAAAAn+/DUQKDBR3i+Xj9gAAAAAAAAoMFHeL5eP2AAAAAAAACSZTSUdOQUw6THk5RU1HeEtSMGhGWlV0cE1IZENVVlpuUlVGblp6MTcBA6QTOe8=",
		hex:    "0xFC306B00000000000000FFF00506FF000000000055025343554549000000027FBF0D440A0C14778BE5E3F60000000000000A0C14778BE5E3F600000000000009265349474E414C3A4C7939454D47784B523068465A5574704D48644355565A6E5255466E5A7A31370103A41339EF",
		bytes: []byte{
			0xFC, 0x30, 0x6B, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xF0, 0x05, 0x06, 0xFF, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x55, 0x02, 0x53, 0x43, 0x55, 0x45, 0x49, 0x00, 0x00, 0x00, 0x02, 0x7F,
			0xBF, 0x0D, 0x44, 0x0A, 0x0C, 0x14, 0x77, 0x8B, 0xE5, 0xE3, 0xF6, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x0A, 0x0C, 0x14, 0x77, 0x8B, 0xE5, 0xE3, 0xF6, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09,
			0x26, 0x53, 0x49, 0x47, 0x4E, 0x41, 0x4C, 0x3A, 0x4C, 0x79, 0x39, 0x45, 0x4D, 0x47, 0x78, 0x4B,
			0x52, 0x30, 0x68, 0x46, 0x5A, 0x55, 0x74, 0x70, 0x4D, 0x48, 0x64, 0x43, 0x55, 0x56, 0x5A, 0x6E,
			0x52, 0x55, 0x46, 0x6E, 0x5A, 0x7A, 0x31, 0x37, 0x01, 0x03, 0xA4, 0x13, 0x39, 0xEF,
		},
		expected: scte35.SpliceInfoSection{
			SpliceCommand: scte35.NewTimeSignal(4294967296),
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.SegmentationDescriptor{
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{
							Type:   scte35.SegmentationUPIDTypeEIDR,
							Format: scte35.SegmentationUPIDFormatText,
							Value:  "10.5239/8BE5-E3F6-0000-0000-0000",
						},
						{
							Type:   scte35.SegmentationUPIDTypeEIDR,
							Format: scte35.SegmentationUPIDFormatText,
							Value:  "10.5239/8BE5-E3F6-0000-0000-0000",
						},
						{
							Type:   scte35.SegmentationUPIDTypeADI,
							Format: scte35.SegmentationUPIDFormatText,
							Value:  "SIGNAL:Ly9EMGxKR0hFZUtpMHdCUVZnRUFnZz1",
						},
					},
					SegmentationEventID: 2,
					SegmentationTypeID:  scte35.SegmentationTypeDistributorPOEnd,
					SegmentNum:          1,
					SegmentsExpected:    3,
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Legacy splice_command_length: 0xFFF": {
		base64: "/DA8AAAAAAAAAP///wb+06ACpQAmAiRDVUVJAACcHX//AACky4AMEERJU0NZTVdGMDQ1MjAwMEgxAQEMm4c0",
		hex:    "0xFC303C00000000000000FFFFFF06FED3A002A5002602244355454900009C1D7FFF0000A4CB800C1044495343594D574630343532303030483101010C9B8734",
		bytes: []byte{
			0xFC, 0x30, 0x3C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x06, 0xFE, 0xD3,
			0xA0, 0x02, 0xA5, 0x00, 0x26, 0x02, 0x24, 0x43, 0x55, 0x45, 0x49, 0x00, 0x00, 0x9C, 0x1D, 0x7F,
			0xFF, 0x00, 0x00, 0xA4, 0xCB, 0x80, 0x0C, 0x10, 0x44, 0x49, 0x53, 0x43, 0x59, 0x4D, 0x57, 0x46,
			0x30, 0x34, 0x35, 0x32, 0x30, 0x30, 0x30, 0x48, 0x31, 0x01, 0x01, 0x0C, 0x9B, 0x87, 0x34,
		},
		expected: scte35.SpliceInfoSection{
			SpliceCommand: scte35.NewTimeSignal(3550479013),
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.SegmentationDescriptor{
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{
							Type:             scte35.SegmentationUPIDTypeMPU,
							Format:           scte35.SegmentationUPIDFormatBase64,
							FormatIdentifier: ptr(uint32(1145656131)),
							Value:            "WU1XRjA0NTIwMDBI",
						},
					},
					SegmentationDuration: ptr(uint64(10800000)),
					SegmentationEventID:  39965,
					SegmentationTypeID:   scte35.SegmentationTypeProviderAdEnd,
					SegmentNum:           1,
					SegmentsExpected:     1,
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
		legacy: true,
	},
	"SpliceInsert Time Specified With Invalid Component Count": {
		base64: "FkC1lwP3uTQD0VvxHwVBEH89G6B7VjzaZ9eNuyUF9q8pYAIXsRM9ZpDCczBeDbytQhXkssQstGJVGcvjZ3tiIMULiA4BpRHlzLGFa0q6aVMtzk8ZRUeLcxtKibgVOKBBnkCbOQyhSflFiDkrAAIp+Fk+VRsByTSkPN3RvyK+lWcjHElhwa9hNFcAy4dm3DdeRXnrD3I2mISNc7DkgS0ReotPyp94FV77xMHT4D7SYL48XU20UM4bgg==",
		hex:    "0x1640B59703F7B93403D15BF11F0541107F3D1BA07B563CDA67D78DBB2505F6AF29600217B1133D6690C273305E0DBCAD4215E4B2C42CB4625519CBE3677B6220C50B880E01A511E5CCB1856B4ABA69532DCE4F1945478B731B4A89B81538A0419E409B390CA149F94588392B000229F8593E551B01C934A43CDDD1BF22BE9567231C4961C1AF61345700CB8766DC375E4579EB0F723698848D73B0E4812D117A8B4FCA9F78155EFBC4C1D3E03ED260BE3C5D4DB450CE1B82",
		bytes: []byte{
			0x16, 0x40, 0xB5, 0x97, 0x03, 0xF7, 0xB9, 0x34, 0x03, 0xD1, 0x5B, 0xF1, 0x1F, 0x05, 0x41, 0x10,
			0x7F, 0x3D, 0x1B, 0xA0, 0x7B, 0x56, 0x3C, 0xDA, 0x67, 0xD7, 0x8D, 0xBB, 0x25, 0x05, 0xF6, 0xAF,
			0x29, 0x60, 0x02, 0x17, 0xB1, 0x13, 0x3D, 0x66, 0x90, 0xC2, 0x73, 0x30, 0x5E, 0x0D, 0xBC, 0xAD,
			0x42, 0x15, 0xE4, 0xB2, 0xC4, 0x2C, 0xB4, 0x62, 0x55, 0x19, 0xCB, 0xE3, 0x67, 0x7B, 0x62, 0x20,
			0xC5, 0x0B, 0x88, 0x0E, 0x01, 0xA5, 0x11, 0xE5, 0xCC, 0xB1, 0x85, 0x6B, 0x4A, 0xBA, 0x69, 0x53,
			0x2D, 0xCE, 0x4F, 0x19, 0x45, 0x47, 0x8B, 0x73, 0x1B, 0x4A, 0x89, 0xB8, 0x15, 0x38, 0xA0, 0x41,
			0x9E, 0x40, 0x9B, 0x39, 0x0C, 0xA1, 0x49, 0xF9, 0x45, 0x88, 0x39, 0x2B, 0x00, 0x02, 0x29, 0xF8,
			0x59, 0x3E, 0x55, 0x1B, 0x01, 0xC9, 0x34, 0xA4, 0x3C, 0xDD, 0xD1, 0xBF, 0x22, 0xBE, 0x95, 0x67,
			0x23, 0x1C, 0x49, 0x61, 0xC1, 0xAF, 0x61, 0x34, 0x57, 0x00, 0xCB, 0x87, 0x66, 0xDC, 0x37, 0x5E,
			0x45, 0x79, 0xEB, 0x0F, 0x72, 0x36, 0x98, 0x84, 0x8D, 0x73, 0xB0, 0xE4, 0x81, 0x2D, 0x11, 0x7A,
			0x8B, 0x4F, 0xCA, 0x9F, 0x78, 0x15, 0x5E, 0xFB, 0xC4, 0xC1, 0xD3, 0xE0, 0x3E, 0xD2, 0x60, 0xBE,
			0x3C, 0x5D, 0x4D, 0xB4, 0x50, 0xCE, 0x1B, 0x82,
		},
		err: scte35.ErrBufferOverflow,
	},
	"Signal with non-CUEI descriptor": {
		base64: "/DBPAAAAAAAAAP/wBQb/Gq9LggA5AAVTQVBTCwIwQ1VFSf////9//wAAFI4PDxx1cm46bmJjdW5pLmNvbTpicmM6NDk5ODY2NDM0MQoBbM98zw==",
		hex:    "0xFC304F00000000000000FFF00506FF1AAF4B8200390005534150530B023043554549FFFFFFFF7FFF0000148E0F0F1C75726E3A6E6263756E692E636F6D3A6272633A343939383636343334310A016CCF7CCF",
		bytes: []byte{
			0xFC, 0x30, 0x4F, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xF0, 0x05, 0x06, 0xFF, 0x1A,
			0xAF, 0x4B, 0x82, 0x00, 0x39, 0x00, 0x05, 0x53, 0x41, 0x50, 0x53, 0x0B, 0x02, 0x30, 0x43, 0x55,
			0x45, 0x49, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F, 0xFF, 0x00, 0x00, 0x14, 0x8E, 0x0F, 0x0F, 0x1C, 0x75,
			0x72, 0x6E, 0x3A, 0x6E, 0x62, 0x63, 0x75, 0x6E, 0x69, 0x2E, 0x63, 0x6F, 0x6D, 0x3A, 0x62, 0x72,
			0x63, 0x3A, 0x34, 0x39, 0x39, 0x38, 0x36, 0x36, 0x34, 0x33, 0x34, 0x31, 0x0A, 0x01, 0x6C, 0xCF,
			0x7C, 0xCF,
		},
		expected: scte35.SpliceInfoSection{
			SpliceCommand: scte35.NewTimeSignal(4742663042),
			SpliceDescriptors: []scte35.SpliceDescriptor{
				&scte35.PrivateDescriptor{
					Identifier:   1396789331,
					PrivateBytes: []byte{11},
				},
				&scte35.SegmentationDescriptor{
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{
							Type:   scte35.SegmentationUPIDTypeURI,
							Format: scte35.SegmentationUPIDFormatText,
							Value:  "urn:nbcuni.com:brc:499866434",
						},
					},
					SegmentationDuration: ptr(uint64(1347087)),
					SegmentationEventID:  4294967295,
					SegmentationTypeID:   scte35.SegmentationTypeProviderAdEnd,
					SegmentNum:           10,
					SegmentsExpected:     1,
				},
			},
			Tier:    4095,
			SAPType: 3,
		},
	},
	"Splice Null - Heartbeat": {
		base64: "/DARAAAAAAAAAP/wAAAAAHpPv/8=",
		hex:    "0xFC301100000000000000FFF0000000007A4FBFFF",
		bytes: []byte{
			0xFC, 0x30, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xF0, 0x00, 0x00, 0x00, 0x00,
			0x7A, 0x4F, 0xBF, 0xFF,
		},
		expected: scte35.SpliceInfoSection{
			SpliceCommand: &scte35.SpliceNull{},
			Tier:          4095,
			SAPType:       3,
		},
	},
	"Invalid CRC_32": {
		base64: "/DA4AAAAAAAAAP/wFAUABDEAf+//mWEhzP4Azf5gAQAAAAATAhFDVUVJAAAAAX+/AQIwNAEAAKeYO3Q=",
		hex:    "0xFC303800000000000000FFF01405000431007FEFFF996121CCFE00CDFE60010000000013021143554549000000017FBF01023034010000A7983B74",
		bytes: []byte{
			0xFC, 0x30, 0x38, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xF0, 0x14, 0x05, 0x00, 0x04,
			0x31, 0x00, 0x7F, 0xEF, 0xFF, 0x99, 0x61, 0x21, 0xCC, 0xFE, 0x00, 0xCD, 0xFE, 0x60, 0x01, 0x00,
			0x00, 0x00, 0x00, 0x13, 0x02, 0x11, 0x43, 0x55, 0x45, 0x49, 0x00, 0x00, 0x00, 0x01, 0x7F, 0xBF,
			0x01, 0x02, 0x30, 0x34, 0x01, 0x00, 0x00, 0xA7, 0x98, 0x3B, 0x74,
		},
		err: scte35.ErrCRC32Invalid,
	},
	"Alignment Stuffing without Encryption": {
		base64: "/DAeAAAAAAAAAP///wViAA/nf18ACQAAAAAskJv+YPtE",
		hex:    "0xFC301E00000000000000FFFFFF0562000FE77F5F0009000000002C909BFE60FB44",
		bytes: []byte{
			0xFC, 0x30, 0x1E, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x05, 0x62, 0x00,
			0x0F, 0xE7, 0x7F, 0x5F, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00, 0x2C, 0x90, 0x9B, 0xFE, 0x60, 0xFB,
			0x44,
		},
		expected: scte35.SpliceInfoSection{
			SpliceCommand: &scte35.SpliceInsert{
				SpliceEventID:       1644171239,
				Program:             &scte35.SpliceInsertProgram{},
				SpliceImmediateFlag: true,
				UniqueProgramID:     9,
			},
			Tier:    4095,
			SAPType: 3,
		},
		legacy: true, // binary wont match because of stuffing
	},
	"Unused Subsegments Included": {
		base64: "/DCRAAAAAAAAAP/wBQb/9peOEAB7AjhDVUVJAAAAnH+/DilhdmFpbGlkPTkxNDg2NjA2NSZiaXRtYXA9JmluYWN0aXZpdHk9MzEyMDEHCgI/Q1VFSQAAAJ1//wAANu6ADilhdmFpbGlkPTkxMDkwMTM4OSZiaXRtYXA9JmluYWN0aXZpdHk9MzEyMDAICgAAoJMeaA==",
		hex:    "0xFC309100000000000000FFF00506FFF6978E10007B0238435545490000009C7FBF0E29617661696C69643D393134383636303635266269746D61703D26696E61637469766974793D3331323031070A023F435545490000009D7FFF000036EE800E29617661696C69643D393130393031333839266269746D61703D26696E61637469766974793D3331323030080A0000A0931E68",
		bytes: []byte{
			0xFC, 0x30, 0x91, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xF0, 0x05, 0x06, 0xFF, 0xF6,
			0x97, 0x8E, 0x10, 0x00, 0x7B, 0x02, 0x38, 0x43, 0x55, 0x45, 0x49, 0x00, 0x00, 0x00, 0x9C, 0x7F,
			0xBF, 0x0E, 0x29, 0x61, 0x76, 0x61, 0x69, 0x6C, 0x69, 0x64, 0x3D, 0x39, 0x31, 0x34, 0x38, 0x36,
			0x36, 0x30, 0x36, 0x35, 0x26, 0x62, 0x69, 0x74, 0x6D, 0x61, 0x70, 0x3D, 0x26, 0x69, 0x6E, 0x61,
			0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x3D, 0x33, 0x31, 0x32, 0x30, 0x31, 0x07, 0x0A, 0x02,
			0x3F, 0x43, 0x55, 0x45, 0x49, 0x00, 0x00, 0x00, 0x9D, 0x7F, 0xFF, 0x00, 0x00, 0x36, 0xEE, 0x80,
			0x0E, 0x29, 0x61, 0x76, 0x61, 0x69, 0x6C, 0x69, 0x64, 0x3D, 0x39, 0x31, 0x30, 0x39, 0x30, 0x31,
			0x33, 0x38, 0x39, 0x26, 0x62, 0x69, 0x74, 0x6D, 0x61, 0x70, 0x3D, 0x26, 0x69, 0x6E, 0x61, 0x63,
			0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x3D, 0x33, 0x31, 0x32, 0x30, 0x30, 0x08, 0x0A, 0x00, 0x00,
			0xA0, 0x93, 0x1E, 0x68,
		},
		expected: scte35.SpliceInfoSection{
			SAPType: scte35.SAPTypeNotSpecified,
			Tier:    4095,
			SpliceCommand: &scte35.TimeSignal{
				SpliceTime: scte35.SpliceTime{
					PTSTime: ptr(uint64(8432094736)),
				},
			},
			SpliceDescriptors: scte35.SpliceDescriptors{
				&scte35.SegmentationDescriptor{
					SegmentationEventID: uint32(156),
					SegmentationTypeID:  scte35.SegmentationTypeProviderAdEnd,
					SegmentNum:          7,
					SegmentsExpected:    10,
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{Type: scte35.SegmentationUPIDTypeADS, Format: scte35.SegmentationUPIDFormatText, Value: "availid=914866065\u0026bitmap=\u0026inactivity=3120"},
					},
				},
				&scte35.SegmentationDescriptor{
					SegmentationEventID:  uint32(157),
					SegmentationTypeID:   scte35.SegmentationTypeProviderAdStart,
					SegmentationDuration: ptr(uint64(3600000)),
					SegmentNum:           8,
					SegmentsExpected:     10,
					SubSegmentNum:        ptr(uint32(0)),
					SubSegmentsExpected:  ptr(uint32(0)),
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{Type: scte35.SegmentationUPIDTypeADS, Format: scte35.SegmentationUPIDFormatText, Value: "availid=910901389\u0026bitmap=\u0026inactivity=3120"},
					},
				},
			},
		},
	},
	"UPID with Valid ASCII Invalid UTF8": {
		base64: "/DDHAAAAABc0AP/wBQb/tVo+agCxAhdDVUVJQA4hwH+fCAgAAAAAPj6IcCMAAAIXQ1VFSUAOI1x/nwgIAAAAAD4+iHARAAACF0NVRUlADiHgf58ICAAAAAA+Poi2EAAAAhxDVUVJQA4hyn/fAABSlKwICAAAAAA+Poi2IgAAAkZDVUVJQA4h1n/PAABSlKwNMgoMFHf5uXs0AAAAAAAADhh0eXBlPUxBJmR1cj02MDAwMCZ0aWVy/DDHAAAAAAAAAP8ABQb/HPCt2w==",
		hex:    "0xFC30C700000000173400FFF00506FFB55A3E6A00B1021743554549400E21C07F9F0808000000003E3E8870230000021743554549400E235C7F9F0808000000003E3E8870110000021743554549400E21E07F9F0808000000003E3E88B6100000021C43554549400E21CA7FDF00005294AC0808000000003E3E88B6220000024643554549400E21D67FCF00005294AC0D320A0C1477F9B97B340000000000000E18747970653D4C41266475723D36303030302674696572FC30C700000000000000FF000506FF1CF0ADDB",
		bytes: []byte{
			0xFC, 0x30, 0xC7, 0x00, 0x00, 0x00, 0x00, 0x17, 0x34, 0x00, 0xFF, 0xF0, 0x05, 0x06, 0xFF, 0xB5,
			0x5A, 0x3E, 0x6A, 0x00, 0xB1, 0x02, 0x17, 0x43, 0x55, 0x45, 0x49, 0x40, 0x0E, 0x21, 0xC0, 0x7F,
			0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x3E, 0x3E, 0x88, 0x70, 0x23, 0x00, 0x00, 0x02, 0x17,
			0x43, 0x55, 0x45, 0x49, 0x40, 0x0E, 0x23, 0x5C, 0x7F, 0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00,
			0x3E, 0x3E, 0x88, 0x70, 0x11, 0x00, 0x00, 0x02, 0x17, 0x43, 0x55, 0x45, 0x49, 0x40, 0x0E, 0x21,
			0xE0, 0x7F, 0x9F, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x3E, 0x3E, 0x88, 0xB6, 0x10, 0x00, 0x00,
			0x02, 0x1C, 0x43, 0x55, 0x45, 0x49, 0x40, 0x0E, 0x21, 0xCA, 0x7F, 0xDF, 0x00, 0x00, 0x52, 0x94,
			0xAC, 0x08, 0x08, 0x00, 0x00, 0x00, 0x00, 0x3E, 0x3E, 0x88, 0xB6, 0x22, 0x00, 0x00, 0x02, 0x46,
			0x43, 0x55, 0x45, 0x49, 0x40, 0x0E, 0x21, 0xD6, 0x7F, 0xCF, 0x00, 0x00, 0x52, 0x94, 0xAC, 0x0D,
			0x32, 0x0A, 0x0C, 0x14, 0x77, 0xF9, 0xB9, 0x7B, 0x34, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0E,
			0x18, 0x74, 0x79, 0x70, 0x65, 0x3D, 0x4C, 0x41, 0x26, 0x64, 0x75, 0x72, 0x3D, 0x36, 0x30, 0x30,
			0x30, 0x30, 0x26, 0x74, 0x69, 0x65, 0x72, 0xFC, 0x30, 0xC7, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0xFF, 0x00, 0x05, 0x06, 0xFF, 0x1C, 0xF0, 0xAD, 0xDB,
		},
		expected: scte35.SpliceInfoSection{
			SpliceCommand: &scte35.TimeSignal{
				SpliceTime: scte35.SpliceTime{
					PTSTime: ptr(uint64(7337557610)),
				},
			},
			SpliceDescriptors: scte35.SpliceDescriptors{
				&scte35.SegmentationDescriptor{
					SegmentationEventID: uint32(1074667968),
					SegmentationTypeID:  scte35.SegmentationTypeBreakEnd,
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						ArchiveAllowedFlag:     true,
						WebDeliveryAllowedFlag: true,
						NoRegionalBlackoutFlag: true,
						DeviceRestrictions:     3,
					},
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{Type: scte35.SegmentationUPIDTypeTI, Format: scte35.SegmentationUPIDFormatText, Value: "1044285552"},
					},
				},
				&scte35.SegmentationDescriptor{
					SegmentationEventID: uint32(1074668380),
					SegmentationTypeID:  scte35.SegmentationTypeProgramEnd,
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						ArchiveAllowedFlag:     true,
						WebDeliveryAllowedFlag: true,
						NoRegionalBlackoutFlag: true,
						DeviceRestrictions:     3,
					},
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{Type: scte35.SegmentationUPIDTypeTI, Format: scte35.SegmentationUPIDFormatText, Value: "1044285552"},
					},
				},
				&scte35.SegmentationDescriptor{
					SegmentationEventID: uint32(1074668000),
					SegmentationTypeID:  scte35.SegmentationTypeProgramStart,
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						ArchiveAllowedFlag:     true,
						WebDeliveryAllowedFlag: true,
						NoRegionalBlackoutFlag: true,
						DeviceRestrictions:     3,
					},
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{Type: scte35.SegmentationUPIDTypeTI, Format: scte35.SegmentationUPIDFormatText, Value: "1044285622"},
					},
				},
				&scte35.SegmentationDescriptor{
					SegmentationEventID:  uint32(1074667978),
					SegmentationDuration: ptr(uint64(5412012)),
					SegmentationTypeID:   scte35.SegmentationTypeBreakStart,
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						ArchiveAllowedFlag:     true,
						WebDeliveryAllowedFlag: true,
						NoRegionalBlackoutFlag: true,
						DeviceRestrictions:     3,
					},
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{Type: scte35.SegmentationUPIDTypeTI, Format: scte35.SegmentationUPIDFormatText, Value: "1044285622"},
					},
				},
				&scte35.SegmentationDescriptor{
					SegmentationEventID:  uint32(1074667990),
					SegmentationTypeID:   0x05,
					SegmentationDuration: ptr(uint64(5412012)),
					SegmentNum:           6,
					SegmentsExpected:     255,
					DeliveryRestrictions: &scte35.DeliveryRestrictions{
						ArchiveAllowedFlag:     true,
						WebDeliveryAllowedFlag: false,
						NoRegionalBlackoutFlag: true,
						DeviceRestrictions:     3,
					},
					SegmentationUPIDs: []scte35.SegmentationUPID{
						{Type: scte35.SegmentationUPIDTypeEIDR, Format: scte35.SegmentationUPIDFormatText, Value: "10.5239/F9B9-7B34-0000-0000-0000"},
						{Type: scte35.SegmentationUPIDTypeADS, Format: scte35.SegmentationUPIDFormatText, Value: "type=LA&dur=60000&tierü0"},
						{Type: uint32(199), Format: scte35.SegmentationUPIDFormatText},
						{Type: uint32(0), Format: scte35.SegmentationUPIDFormatText},
						{Type: uint32(0), Format: scte35.SegmentationUPIDFormatText},
						{Type: uint32(0), Format: scte35.SegmentationUPIDFormatText},
						{Type: uint32(255), Format: scte35.SegmentationUPIDFormatText},
					},
				},
			},
			PTSAdjustment: uint64(5940),
			Tier:          4095,
			SAPType:       3,
		},
	},
}

func TestDecodeBase64(t *testing.T) {
	scte35.Logger.SetOutput(os.Stderr)
	defer scte35.Logger.SetOutput(io.Discard)

	for k, c := range commonDecodeTestcases {
		t.Run(k, func(t *testing.T) {
			// decode the binary
			sis, err := scte35.DecodeBase64(c.base64)
			require.ErrorIs(t, err, c.err)
			if err != nil {
				return
			}

			// test encode/decode XML
			encodedXML := toXML(sis)
			assert.Equal(t, toXML(&c.expected), encodedXML)
			decodedXML := scte35.SpliceInfoSection{}
			assert.NoError(t, xml.Unmarshal([]byte(encodedXML), &decodedXML))

			// legacy 35's produce an "updated" binary so will not match
			if !c.legacy {
				assert.Equal(t, c.base64, decodedXML.Base64())
			}

			// test encode/decode JSON
			encodedJSON := toJSON(sis)
			assert.Equal(t, toJSON(&c.expected), encodedJSON)
			decodedJSON := scte35.SpliceInfoSection{}
			require.NoError(t, json.Unmarshal([]byte(encodedJSON), &decodedJSON))

			// legacy 35's produce an "updated" binary so will not match
			if !c.legacy {
				assert.Equal(t, c.base64, decodedJSON.Base64())
			}

			// uncomment this to verify the output as text
			// scte35.Logger.Printf("\n%s", sis.Table("", "\t"))
		})
	}
}

func TestDecodeHex(t *testing.T) {
	for k, c := range commonDecodeTestcases {
		t.Run(k, func(t *testing.T) {
			// decode the binary
			sis, err := scte35.DecodeHex(c.hex)
			require.ErrorIs(t, err, c.err)
			if err != nil {
				return
			}

			// test encode/decode XML
			encodedXML := toXML(sis)
			assert.Equal(t, toXML(&c.expected), encodedXML)
			decodedXML := scte35.SpliceInfoSection{}
			assert.NoError(t, xml.Unmarshal([]byte(encodedXML), &decodedXML))

			// test encode/decode JSON
			encodedJSON := toJSON(sis)
			assert.Equal(t, toJSON(&c.expected), encodedJSON)
			decodedJSON := scte35.SpliceInfoSection{}
			require.NoError(t, json.Unmarshal([]byte(encodedJSON), &decodedJSON))
		})
	}
}

func TestDecodeBytes(t *testing.T) {
	for k, c := range commonDecodeTestcases {
		t.Run(k, func(t *testing.T) {
			if len(c.bytes) == 0 {
				t.Skip("skipping test, does not apply to DecodeBytes")
			}

			// decode the binary
			sis, err := scte35.DecodeBytes(c.bytes)
			require.ErrorIs(t, err, c.err)
			if err != nil {
				return
			}

			// test encode/decode XML
			encodedXML := toXML(sis)
			assert.Equal(t, toXML(&c.expected), encodedXML)
			decodedXML := scte35.SpliceInfoSection{}
			assert.NoError(t, xml.Unmarshal([]byte(encodedXML), &decodedXML))

			// test encode/decode JSON
			encodedJSON := toJSON(sis)
			assert.Equal(t, toJSON(&c.expected), encodedJSON)
			decodedJSON := scte35.SpliceInfoSection{}
			require.NoError(t, json.Unmarshal([]byte(encodedJSON), &decodedJSON))
		})
	}
}

func TestEncodeWithAlignmentStuffing(t *testing.T) {
	cases := map[string]struct {
		name   string
		binary string
	}{
		"SpliceInsert Program Out Point with 3 bytes alignment stuffing": {
			binary: "/DA0AABS2+YAAACgFAUALJGCf+/+MSwPcX4AUmXAAAAAAAAMAQpDVUVJRp8xMjEq3pnIPCi6lw==",
		},
		"SpliceInsert Program In Point with 3 bytes alignment stuffing": {
			binary: "/DAvAABS2+YAAACgDwUALJGEf0/+MX7z3AAAAAAADAEKQ1VFSQCfMTIxI6SMuQkzWQI=",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sis, err := scte35.DecodeBase64(c.binary)
			require.NoError(t, err)
			require.Equal(t, c.binary, sis.Base64())
		})
	}
}

func TestTicksToDuration(t *testing.T) {
	// test a wide range of tick values
	min := 29 * scte35.TicksPerSecond
	max := 61 * scte35.TicksPerSecond
	for i := min; i < max; i++ {
		d := scte35.TicksToDuration(uint64(i))
		require.Equal(t, i, int(scte35.DurationToTicks(d)))

		// detect for rounding errors. If the tick is a multiple of 90,
		// we should get a value that is a multiple of full milliseconds
		if i%90 == 0 {
			require.Zero(t, d.Nanoseconds()%int64(time.Millisecond))
			require.Equal(t, int64(i/90), d.Milliseconds())
		}
	}
}

func TestDurationToTicks(t *testing.T) {
	// Test for usual framerates
	frameRates := []float64{
		24, 25, 30000.0 / 1001, 30, 50, 60,
	}
	frameTicks := []int{
		3750, 3600, 3003, 3000, 1800, 1500,
	}

	for idx, fps := range frameRates {
		timePerFrame := time.Duration(float64(time.Second) / fps)
		ticksPerFrame := frameTicks[idx]

		// try for durations up to 90 seconds
		for framenum := 0; framenum < int(fps*90); framenum++ {
			ticks := scte35.DurationToTicks(timePerFrame * time.Duration(framenum))
			require.Equal(t, framenum*ticksPerFrame, int(ticks))
		}
	}
}

var commonGPSTimeTestCases = map[string]struct {
	gpsSeconds    uint32
	utcSpliceTime scte35.UTCSpliceTime
}{
	"GPS Second 0": {
		gpsSeconds: 0,
		utcSpliceTime: scte35.UTCSpliceTime{
			Time: time.Date(1980, time.January, 6, 0, 0, 0, 0, time.UTC),
		},
	},
	// last value before a uint32 based calculation would overflow
	"GPS Second 3979002495": {
		gpsSeconds: 3979002495,
		utcSpliceTime: scte35.UTCSpliceTime{
			Time: time.Date(2106, time.February, 7, 6, 28, 15, 0, time.UTC),
		},
	},
	// in a uint32 based calculation this would overflow
	"GPS Second 3979002496": {
		gpsSeconds: 3979002496,
		utcSpliceTime: scte35.UTCSpliceTime{
			Time: time.Date(2106, time.February, 7, 6, 28, 16, 0, time.UTC),
		},
	},
	// highest possible value
	"GPS Second 4294967295": {
		gpsSeconds: 4294967295,
		utcSpliceTime: scte35.UTCSpliceTime{
			Time: time.Date(2116, time.February, 12, 6, 28, 15, 0, time.UTC),
		},
	},
}

func TestNewUTCSpliceTime(t *testing.T) {
	for k, c := range commonGPSTimeTestCases {
		t.Run(k, func(t *testing.T) {
			utcSpliceTime := scte35.NewUTCSpliceTime(c.gpsSeconds)
			assert.Equal(t, c.utcSpliceTime.Time, utcSpliceTime.Time)
		})
	}
}

func TestUTCSpliceTimeGPSSeconds(t *testing.T) {
	for k, c := range commonGPSTimeTestCases {
		t.Run(k, func(t *testing.T) {
			gpsSeconds := c.utcSpliceTime.GPSSeconds()
			assert.Equal(t, c.gpsSeconds, gpsSeconds)
		})
	}
}

// helper func to make test life a bit easier

func toBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

func toJSON(sis *scte35.SpliceInfoSection) string {
	b, _ := json.MarshalIndent(sis, "", "\t")
	return string(b)
}

func toXML(sis *scte35.SpliceInfoSection) string {
	b, _ := xml.MarshalIndent(sis, "", "\t")
	return string(b)
}

func ptr[T any](i T) *T {
	return &i
}
