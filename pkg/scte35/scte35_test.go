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
	"fmt"
	"io"
	"os"
	"testing"
	"unicode/utf8"

	"github.com/Comcast/scte35-go/pkg/scte35"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeBase64(t *testing.T) {
	scte35.Logger.SetOutput(os.Stderr)
	defer scte35.Logger.SetOutput(io.Discard)

	// when adding tests that contain multiple splice descriptors, care must be
	// taken to ensure they are in the order specified in the custom UnmarshalXML
	// implementation, otherwise misleading error may occur
	cases := map[string]struct {
		binary   string
		err      error
		expected scte35.SpliceInfoSection
		legacy   bool
	}{
		"Sample 14.1 time_signal - Placement Opportunity Start": {
			binary: "/DA0AAAAAAAA///wBQb+cr0AUAAeAhxDVUVJSAAAjn/PAAGlmbAICAAAAAAsoKGKNAIAmsnRfg==",
			expected: scte35.SpliceInfoSection{
				EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{
						PTSTime: uint64ptr(0x072bd0050),
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
						SegmentationDuration: uint64ptr(0x0001a599b0),
						SegmentationUPIDs: []scte35.SegmentationUPID{
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002ca0a18a)),
						},
						SegmentNum: 2,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.2 splice_insert": {
			binary: "/DAvAAAAAAAA///wFAVIAACPf+/+c2nALv4AUsz1AAAAAAAKAAhDVUVJAAABNWLbowo=",
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
							PTSTime: uint64ptr(0x07369c02e),
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
			binary: "/DAvAAAAAAAA///wBQb+dGKQoAAZAhdDVUVJSAAAjn+fCAgAAAAALKChijUCAKnMZ1g=",
			expected: scte35.SpliceInfoSection{
				EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{
						PTSTime: uint64ptr(0x0746290a0),
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
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002ca0a18a)),
						},
						SegmentNum: 2,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.4 time_signal - Program Start/End": {
			binary: "/DBIAAAAAAAA///wBQb+ek2ItgAyAhdDVUVJSAAAGH+fCAgAAAAALMvDRBEAAAIXQ1VFSUgAABl/nwgIAAAAACyk26AQAACZcuND",
			expected: scte35.SpliceInfoSection{
				EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{
						PTSTime: uint64ptr(0x07a4d88b6),
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
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002ccbc344)),
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
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002ca4dba0)),
						},
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.5 time_signal - Program Overlap Start": {
			binary: "/DAvAAAAAAAA///wBQb+rr//ZAAZAhdDVUVJSAAACH+fCAgAAAAALKVs9RcAAJUdsKg=",
			expected: scte35.SpliceInfoSection{
				EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{
						PTSTime: uint64ptr(0x0aebfff64),
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
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002ca56cf5)),
						},
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.6 time_signal - Program Blackout Override / Program End": {
			binary: "/DBIAAAAAAAA///wBQb+ky44CwAyAhdDVUVJSAAACn+fCAgAAAAALKCh4xgAAAIXQ1VFSUgAAAl/nwgIAAAAACygoYoRAAC0IX6w",
			expected: scte35.SpliceInfoSection{
				EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{
						PTSTime: uint64ptr(0x0932e380b),
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
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002ca0a1e3)),
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
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002ca0a18a)),
						},
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.7 time_signal - Program End": {
			binary: "/DAvAAAAAAAA///wBQb+rvF8TAAZAhdDVUVJSAAAB3+fCAgAAAAALKVslxEAAMSHai4=",
			expected: scte35.SpliceInfoSection{
				EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{
						PTSTime: uint64ptr(0x0aef17c4c),
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
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002ca56c97)),
						},
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.8 time_signal - Program Start/End - Placement Opportunity End": {
			binary: "/DBhAAAAAAAA///wBQb+qM1E7QBLAhdDVUVJSAAArX+fCAgAAAAALLLXnTUCAAIXQ1VFSUgAACZ/nwgIAAAAACyy150RAAACF0NVRUlIAAAnf58ICAAAAAAsstezEAAAihiGnw==",
			expected: scte35.SpliceInfoSection{
				EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{
						PTSTime: uint64ptr(0x0a8cd44ed),
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
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002cb2d79d)),
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
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002cb2d79d)),
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
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002cb2d7b3)),
						},
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"SpliceInsert With DTMF": {
			binary: "/DAxAAAAAAAAAP/wFAVAAIeuf+/+0AWRK/4AUmXAAC0AfwAMAQpDVUVJUJ81MTkqo5/+gA==",
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
			binary: "/DBIAAAAAAAAAP/wBQb/tB67hgAyAhdDVUVJQAABEn+fCAgAAAAALzE8BTUAAAIXQ1VFSUAAAEV/nwgIAAAAAC8xPN4jAAAfiOPE",
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
								Type:  8,
								Value: "791755781",
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
								Type:  8,
								Value: "791755998",
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
		"Empty String": {
			binary: "",
			err:    fmt.Errorf("splice_info_section: %w", scte35.ErrBufferOverflow),
		},
		"Invalid Base64 Encoding": {
			binary: "/DBaf%^",
			err:    scte35.ErrUnsupportedEncoding,
		},
		"Splice Insert with Avail Descriptor": {
			binary: "/DAqAAAAAAAAAP/wDwUAAHn+f8/+QubGOQAAAAAACgAIQ1VFSQAAAADizteX",
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
			binary: "/DBrAAAAAAAAAP/wBQb/AAAAAABVAlNDVUVJAAAAAn+/DUQKDBR3i+Xj9gAAAAAAAAoMFHeL5eP2AAAAAAAACSZTSUdOQUw6THk5RU1HeEtSMGhGWlV0cE1IZENVVlpuUlVGblp6MTcBA6QTOe8=",
			expected: scte35.SpliceInfoSection{
				SpliceCommand: scte35.NewTimeSignal(4294967296),
				SpliceDescriptors: []scte35.SpliceDescriptor{
					&scte35.SegmentationDescriptor{
						SegmentationUPIDs: []scte35.SegmentationUPID{
							{
								Type:  scte35.SegmentationUPIDTypeEIDR,
								Value: "10.5239/8BE5-E3F6-0000-0000-0000",
							},
							{
								Type:  scte35.SegmentationUPIDTypeEIDR,
								Value: "10.5239/8BE5-E3F6-0000-0000-0000",
							},
							{
								Type:  scte35.SegmentationUPIDTypeADI,
								Value: "SIGNAL:Ly9EMGxKR0hFZUtpMHdCUVZnRUFnZz1",
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
			binary: "/DA8AAAAAAAAAP///wb+06ACpQAmAiRDVUVJAACcHX//AACky4AMEERJU0NZTVdGMDQ1MjAwMEgxAQEMm4c0",
			expected: scte35.SpliceInfoSection{
				SpliceCommand: scte35.NewTimeSignal(3550479013),
				SpliceDescriptors: []scte35.SpliceDescriptor{
					&scte35.SegmentationDescriptor{
						SegmentationUPIDs: []scte35.SegmentationUPID{
							{
								Type:             scte35.SegmentationUPIDTypeMPU,
								FormatIdentifier: uint32ptr(1145656131),
								Value:            "WU1XRjA0NTIwMDBI",
							},
						},
						SegmentationDuration: uint64ptr(10800000),
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
			binary: "FkC1lwP3uTQD0VvxHwVBEH89G6B7VjzaZ9eNuyUF9q8pYAIXsRM9ZpDCczBeDbytQhXkssQstGJVGcvjZ3tiIMULiA4BpRHlzLGFa0q6aVMtzk8ZRUeLcxtKibgVOKBBnkCbOQyhSflFiDkrAAIp+Fk+VRsByTSkPN3RvyK+lWcjHElhwa9hNFcAy4dm3DdeRXnrD3I2mISNc7DkgS0ReotPyp94FV77xMHT4D7SYL48XU20UM4bgg==",
			err:    fmt.Errorf("splice_insert: %w", scte35.ErrBufferOverflow),
		},
		"Signal with non-CUEI descriptor": {
			binary: "/DBPAAAAAAAAAP/wBQb/Gq9LggA5AAVTQVBTCwIwQ1VFSf////9//wAAFI4PDxx1cm46bmJjdW5pLmNvbTpicmM6NDk5ODY2NDM0MQoBbM98zw==",
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
								Type:  scte35.SegmentationUPIDTypeURI,
								Value: "urn:nbcuni.com:brc:499866434",
							},
						},
						SegmentationDuration: uint64ptr(1347087),
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
			binary: "/DARAAAAAAAAAP/wAAAAAHpPv/8=",
			expected: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceNull{},
				Tier:          4095,
				SAPType:       3,
			},
		},
		"Invalid CRC_32": {
			binary: "/DA4AAAAAAAAAP/wFAUABDEAf+//mWEhzP4Azf5gAQAAAAATAhFDVUVJAAAAAX+/AQIwNAEAAKeYO3Q=",
			err:    fmt.Errorf("splice_info_section: %w", scte35.ErrCRC32Invalid),
		},
		"Alignment Stuffing without Encryption": {
			binary: "/DAeAAAAAAAAAP///wViAA/nf18ACQAAAAAskJv+YPtE",
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
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			// decode the binary
			sis, err := scte35.DecodeBase64(c.binary)
			require.Equal(t, c.err, err)
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
				assert.Equal(t, c.binary, decodedXML.Base64())
			}

			// test encode/decode JSON
			encodedJSON := toJSON(sis)
			assert.Equal(t, toJSON(&c.expected), encodedJSON)
			decodedJSON := scte35.SpliceInfoSection{}
			require.NoError(t, json.Unmarshal([]byte(encodedJSON), &decodedJSON))

			// legacy 35's produce an "updated" binary so will not match
			if !c.legacy {
				assert.Equal(t, c.binary, decodedJSON.Base64())
			}

			// uncomment this to verify the output as text
			// scte35.Logger.Printf("\n%s", sis.Table("", "\t"))
		})
	}
}

func TestDecodeHex(t *testing.T) {
	// when adding tests that contain multiple splice descriptors, care must be
	// taken to ensure they are in the order specified in the custom UnmarshalXML
	// implementation, otherwise misleading error may occur
	cases := map[string]struct {
		hex      string
		err      error
		expected scte35.SpliceInfoSection
	}{
		"Sample 14.1 time_signal - Placement Opportunity Start": {
			hex: "0xFC3034000000000000FFFFF00506FE72BD0050001E021C435545494800008E7FCF0001A599B00808000000002CA0A18A3402009AC9D17E",
			expected: scte35.SpliceInfoSection{
				EncryptedPacket: scte35.EncryptedPacket{EncryptionAlgorithm: scte35.EncryptionAlgorithmNone, CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{
						PTSTime: uint64ptr(0x072bd0050),
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
						SegmentationDuration: uint64ptr(0x0001a599b0),
						SegmentationUPIDs: []scte35.SegmentationUPID{
							scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, bytes(0x000000002ca0a18a)),
						},
						SegmentNum: 2,
					},
				},
				Tier:    4095,
				SAPType: 3,
			},
		},
		"Sample 14.2 splice_insert (no prefix)": {
			hex: "FC302F000000000000FFFFF014054800008F7FEFFE7369C02EFE0052CCF500000000000A0008435545490000013562DBA30A",
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
							PTSTime: uint64ptr(0x07369c02e),
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
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			// decode the binary
			sis, err := scte35.DecodeHex(c.hex)
			require.Equal(t, c.err, err)
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

// Segmentation UPIDs are required to be US-ASCII, but Go uses UTF8 strings.  Make sure the decoder
// decodes the strings properly.
func TestASCIItoUTF8(t *testing.T) {
	cases := []struct {
		name   string
		binary string
	}{
		{
			name:   "Time Signal, multiple descriptors, valid ASCII but invalid UTF8 segmentation UPIDs",
			binary: "/DDHAAAAABc0AP/wBQb/tVo+agCxAhdDVUVJQA4hwH+fCAgAAAAAPj6IcCMAAAIXQ1VFSUAOI1x/nwgIAAAAAD4+iHARAAACF0NVRUlADiHgf58ICAAAAAA+Poi2EAAAAhxDVUVJQA4hyn/fAABSlKwICAAAAAA+Poi2IgAAAkZDVUVJQA4h1n/PAABSlKwNMgoMFHf5uXs0AAAAAAAADhh0eXBlPUxBJmR1cj02MDAwMCZ0aWVy/DDHAAAAAAAAAP/wBQb/dvhrwQ==",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sis, err := scte35.DecodeBase64(c.binary)
			require.NoError(t, err)

			for _, sd := range sis.SpliceDescriptors {
				if seg, ok := sd.(*scte35.SegmentationDescriptor); ok {
					for _, upid := range seg.SegmentationUPIDs {
						require.True(t, utf8.ValidString(upid.Value))
					}
				}
			}
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
	}
}

// helper func to make test life a bit easier

func bytes(i uint64) []byte {
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

func uint32ptr(i uint32) *uint32 {
	return &i
}

func uint64ptr(i uint64) *uint64 {
	return &i
}
