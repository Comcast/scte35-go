package scte35_test

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/Comcast/scte35-go/pkg/scte35"
	"github.com/stretchr/testify/require"
)

func TestSpliceInfoSection_UnmarshalXML(t *testing.T) {
	cases := map[string]struct {
		xml      string
		expected *scte35.SpliceInfoSection
	}{
		"SAPType Missing": {
			xml: `
				<SpliceInfoSection xmlns="http://www.scte.org/schemas/35" tier="4095">
					<EncryptedPacket xmlns="http://www.scte.org/schemas/35" cwIndex="255"></EncryptedPacket>
					<TimeSignal xmlns="http://www.scte.org/schemas/35">
						<SpliceTime xmlns="http://www.scte.org/schemas/35" ptsTime="1924989008"></SpliceTime>
					</TimeSignal>
					<SegmentationDescriptor xmlns="http://www.scte.org/schemas/35" segmentationEventId="1207959694" segmentationDuration="27630000" segmentationTypeId="52" segmentNum="2">
						<DeliveryRestrictions xmlns="http://www.scte.org/schemas/35" archiveAllowedFlag="true" webDeliveryAllowedFlag="false" noRegionalBlackoutFlag="true" deviceRestrictions="3"></DeliveryRestrictions>
						<SegmentationUpid xmlns="http://www.scte.org/schemas/35" segmentationUpidType="8">748724618</SegmentationUpid>
					</SegmentationDescriptor>
				</SpliceInfoSection>`,
			expected: &scte35.SpliceInfoSection{
				Tier:            uint32(4095),
				SAPType:         scte35.SAPTypeNotSpecified,
				EncryptedPacket: scte35.EncryptedPacket{CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{PTSTime: ptr(uint64(1924989008))},
				},
				SpliceDescriptors: scte35.SpliceDescriptors{
					&scte35.SegmentationDescriptor{
						DeliveryRestrictions: &scte35.DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: false,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []scte35.SegmentationUPID{
							{
								Type:  scte35.SegmentationUPIDTypeTI,
								Value: "748724618",
							},
						},
						SegmentationEventID:  uint32(1207959694),
						SegmentationDuration: ptr(uint64(27630000)),
						SegmentationTypeID:   scte35.SegmentationTypeProviderPOStart,
						SegmentNum:           2,
					},
				},
			},
		},
		"SAPType Specified": {
			xml: `
				<SpliceInfoSection xmlns="http://www.scte.org/schemas/35" tier="4095" sapType="0">
					<EncryptedPacket xmlns="http://www.scte.org/schemas/35" cwIndex="255"></EncryptedPacket>
					<TimeSignal xmlns="http://www.scte.org/schemas/35">
						<SpliceTime xmlns="http://www.scte.org/schemas/35" ptsTime="1924989008"></SpliceTime>
					</TimeSignal>
					<SegmentationDescriptor xmlns="http://www.scte.org/schemas/35" segmentationEventId="1207959694" segmentationDuration="27630000" segmentationTypeId="52" segmentNum="2">
						<DeliveryRestrictions xmlns="http://www.scte.org/schemas/35" archiveAllowedFlag="true" webDeliveryAllowedFlag="false" noRegionalBlackoutFlag="true" deviceRestrictions="3"></DeliveryRestrictions>
						<SegmentationUpid xmlns="http://www.scte.org/schemas/35" segmentationUpidType="8">748724618</SegmentationUpid>
					</SegmentationDescriptor>
				</SpliceInfoSection>`,
			expected: &scte35.SpliceInfoSection{
				Tier:            uint32(4095),
				SAPType:         scte35.SAPType1,
				EncryptedPacket: scte35.EncryptedPacket{CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{PTSTime: ptr(uint64(1924989008))},
				},
				SpliceDescriptors: scte35.SpliceDescriptors{
					&scte35.SegmentationDescriptor{
						DeliveryRestrictions: &scte35.DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: false,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []scte35.SegmentationUPID{
							{
								Type:  scte35.SegmentationUPIDTypeTI,
								Value: "748724618",
							},
						},
						SegmentationEventID:  uint32(1207959694),
						SegmentationDuration: ptr(uint64(27630000)),
						SegmentationTypeID:   scte35.SegmentationTypeProviderPOStart,
						SegmentNum:           2,
					},
				},
			},
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var sis scte35.SpliceInfoSection
			require.NoError(t, xml.Unmarshal([]byte(c.xml), &sis))
			require.Equal(t, toXML(c.expected), toXML(&sis))
		})
	}
}

func TestSpliceInfoSection_UnmarshalJSON(t *testing.T) {
	cases := map[string]struct {
		json     string
		expected *scte35.SpliceInfoSection
	}{
		"SAPType Missing": {
			json: `{
				"encryptedPacket": {
					"cwIndex": 255
				},
				"spliceCommand": {
					"type": 6,
					"spliceTime": {
						"ptsTime": 1924989008
					}
				},
				"spliceDescriptors": [
					{
						"type": 2,
						"deliveryRestrictions": {
							"archiveAllowedFlag": true,
							"webDeliveryAllowedFlag": false,
							"noRegionalBlackoutFlag": true,
							"deviceRestrictions": 3
						},
						"segmentationUpids": [
							{
								"segmentationUpidType": 8,
								"format": "text",
								"value": "748724618"
							}
						],
						"segmentationEventId": 1207959694,
						"segmentationDuration": 27630000,
						"segmentationTypeId": 52,
						"segmentNum": 2
					}
				],
				"tier": 4095
			}`,
			expected: &scte35.SpliceInfoSection{
				Tier:            uint32(4095),
				SAPType:         scte35.SAPTypeNotSpecified,
				EncryptedPacket: scte35.EncryptedPacket{CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{PTSTime: ptr(uint64(1924989008))},
				},
				SpliceDescriptors: scte35.SpliceDescriptors{
					&scte35.SegmentationDescriptor{
						DeliveryRestrictions: &scte35.DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: false,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []scte35.SegmentationUPID{
							{
								Type:  scte35.SegmentationUPIDTypeTI,
								Value: "748724618",
							},
						},
						SegmentationEventID:  uint32(1207959694),
						SegmentationDuration: ptr(uint64(27630000)),
						SegmentationTypeID:   scte35.SegmentationTypeProviderPOStart,
						SegmentNum:           2,
					},
				},
			},
		},
		"SAPType Specified": {
			json: `{
				"encryptedPacket": {
					"cwIndex": 255
				},
				"sapType": 0,
				"spliceCommand": {
					"type": 6,
					"spliceTime": {
						"ptsTime": 1924989008
					}
				},
				"spliceDescriptors": [
					{
						"type": 2,
						"deliveryRestrictions": {
							"archiveAllowedFlag": true,
							"webDeliveryAllowedFlag": false,
							"noRegionalBlackoutFlag": true,
							"deviceRestrictions": 3
						},
						"segmentationUpids": [
							{
								"segmentationUpidType": 8,
								"format": "text",
								"value": "748724618"
							}
						],
						"segmentationEventId": 1207959694,
						"segmentationDuration": 27630000,
						"segmentationTypeId": 52,
						"segmentNum": 2
					}
				],
				"tier": 4095
			}`,
			expected: &scte35.SpliceInfoSection{
				Tier:            uint32(4095),
				SAPType:         scte35.SAPType1,
				EncryptedPacket: scte35.EncryptedPacket{CWIndex: 255},
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{PTSTime: ptr(uint64(1924989008))},
				},
				SpliceDescriptors: scte35.SpliceDescriptors{
					&scte35.SegmentationDescriptor{
						DeliveryRestrictions: &scte35.DeliveryRestrictions{
							ArchiveAllowedFlag:     true,
							WebDeliveryAllowedFlag: false,
							NoRegionalBlackoutFlag: true,
							DeviceRestrictions:     3,
						},
						SegmentationUPIDs: []scte35.SegmentationUPID{
							{
								Type:  scte35.SegmentationUPIDTypeTI,
								Value: "748724618",
							},
						},
						SegmentationEventID:  uint32(1207959694),
						SegmentationDuration: ptr(uint64(27630000)),
						SegmentationTypeID:   scte35.SegmentationTypeProviderPOStart,
						SegmentNum:           2,
					},
				},
			},
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var sis scte35.SpliceInfoSection
			require.NoError(t, json.Unmarshal([]byte(c.json), &sis))
			require.Equal(t, toXML(c.expected), toXML(&sis))
		})
	}
}

type testDurationExpectation struct {
	durationType          scte35.Duration
	ticks                 uint64
	outOfNetworkIndicator bool
	segmentationTypeID    uint32
}

func TestDurations(t *testing.T) {
	cases := map[string]struct {
		sis      scte35.SpliceInfoSection
		expected []testDurationExpectation
	}{
		"SimpleSpliceInsertCueOut": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{
					OutOfNetworkIndicator: true,
					BreakDuration: &scte35.BreakDuration{
						Duration: 2700000,
					},
				},
			},
			expected: []testDurationExpectation{
				{
					durationType:          &scte35.SpliceInsertDuration{},
					ticks:                 2700000,
					outOfNetworkIndicator: true,
				},
			},
		},
		"SimpleSpliceInsertCueIn": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{
					OutOfNetworkIndicator: false,
					BreakDuration: &scte35.BreakDuration{
						Duration: 2700000,
					},
				},
			},
			expected: []testDurationExpectation{
				{
					durationType:          &scte35.SpliceInsertDuration{},
					ticks:                 2700000,
					outOfNetworkIndicator: false,
				},
			},
		},
		"SpliceInsertCueOutWithoutDuration": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{
					OutOfNetworkIndicator: true,
				},
			},
			expected: []testDurationExpectation{},
		},
		"SpliceInsertCueOutPlusSegmentationDescriptors": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{
					OutOfNetworkIndicator: true,
					BreakDuration: &scte35.BreakDuration{
						Duration: 2700000,
					},
				},
				SpliceDescriptors: scte35.SpliceDescriptors{
					&scte35.SegmentationDescriptor{
						SegmentationTypeID:   scte35.SegmentationTypeProviderPOStart,
						SegmentationDuration: ptr(uint64(2790000)),
					},
					&scte35.SegmentationDescriptor{
						SegmentationTypeID:   scte35.SegmentationTypeDistributorPOStart,
						SegmentationDuration: ptr(uint64(2880000)),
					},
				},
			},
			expected: []testDurationExpectation{
				{
					durationType:          &scte35.SpliceInsertDuration{},
					ticks:                 2700000,
					outOfNetworkIndicator: true,
				},
				{
					durationType:       &scte35.SegmentationDuration{},
					ticks:              2790000,
					segmentationTypeID: scte35.SegmentationTypeProviderPOStart,
				},
				{
					durationType:       &scte35.SegmentationDuration{},
					ticks:              2880000,
					segmentationTypeID: scte35.SegmentationTypeDistributorPOStart,
				},
			},
		},
		"TimeSignal": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.TimeSignal{},
				SpliceDescriptors: scte35.SpliceDescriptors{
					&scte35.SegmentationDescriptor{
						SegmentationTypeID:   scte35.SegmentationTypeProviderPOStart,
						SegmentationDuration: ptr(uint64(2790000)),
					},
					&scte35.SegmentationDescriptor{
						SegmentationTypeID:   scte35.SegmentationTypeDistributorPOStart,
						SegmentationDuration: ptr(uint64(2880000)),
					},
				},
			},
			expected: []testDurationExpectation{
				{
					durationType:       &scte35.SegmentationDuration{},
					ticks:              2790000,
					segmentationTypeID: scte35.SegmentationTypeProviderPOStart,
				},
				{
					durationType:       &scte35.SegmentationDuration{},
					ticks:              2880000,
					segmentationTypeID: scte35.SegmentationTypeDistributorPOStart,
				},
			},
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			durations := c.sis.Durations()

			require.Equal(t, len(c.expected), len(durations))

			for i, d := range durations {
				require.IsType(t, c.expected[i].durationType, d)
				if sd, ok := d.(*scte35.SegmentationDuration); ok {
					require.Equal(t, c.expected[i].ticks, sd.Ticks())
					require.Equal(t, c.expected[i].segmentationTypeID, sd.SegmentationTypeID)
				} else if si, ok := d.(*scte35.SpliceInsertDuration); ok {
					require.Equal(t, c.expected[i].ticks, si.Ticks())
					require.Equal(t, c.expected[i].outOfNetworkIndicator, si.OutOfNetworkIndicator)
				}
			}
		})
	}
}

func Test_TimeSpecifiedFlag(t *testing.T) {
	cases := map[string]struct {
		sis      scte35.SpliceInfoSection
		expected bool
	}{
		"SpliceInsert": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{
					Program: &scte35.SpliceInsertProgram{
						SpliceTime: scte35.SpliceTime{
							PTSTime: ptr(uint64(90000)),
						},
					},
				},
			},
			expected: true,
		},
		"SpliceInsertNoProgramSplice": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{},
			},
			expected: false,
		},
		"SpliceInsertNoSpliceTime": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{
					Program: &scte35.SpliceInsertProgram{},
				},
			},
			expected: false,
		},
		"TimeSignal": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{PTSTime: ptr(uint64(90000))},
				},
			},
			expected: true,
		},
		"TimeSignalNoSpliceTime": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{},
				},
			},
			expected: false,
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			require.Equal(t, c.expected, c.sis.TimeSpecifiedFlag())
		})
	}
}

func Test_SpliceTimePTS(t *testing.T) {
	cases := map[string]struct {
		sis      scte35.SpliceInfoSection
		expected uint64
	}{
		"SpliceInsert": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{
					Program: &scte35.SpliceInsertProgram{
						SpliceTime: scte35.SpliceTime{
							PTSTime: ptr(uint64(90000)),
						},
					},
				},
			},
			expected: 90000,
		},
		"SpliceInsertPtsAdjustment": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{
					Program: &scte35.SpliceInsertProgram{
						SpliceTime: scte35.SpliceTime{
							PTSTime: ptr(uint64(90000)),
						},
					},
				},
				PTSAdjustment: 90000,
			},
			expected: 180000,
		},
		"SpliceInsertPtsAdjustmentWrapping": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{
					Program: &scte35.SpliceInsertProgram{
						SpliceTime: scte35.SpliceTime{
							PTSTime: ptr(uint64(8589844592)),
						},
					},
				},
				PTSAdjustment: 180000,
			},
			expected: 90000,
		},
		"SpliceInsertNoProgramSplice": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{},
			},
			expected: 0,
		},
		"SpliceInsertNoSpliceTime": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.SpliceInsert{
					Program: &scte35.SpliceInsertProgram{},
				},
			},
			expected: 0,
		},
		"TimeSignal": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{PTSTime: ptr(uint64(90000))},
				},
			},
			expected: 90000,
		},
		"TimeSingalPtsAdjustment": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{
						PTSTime: ptr(uint64(90000)),
					},
				},
				PTSAdjustment: 90000,
			},
			expected: 180000,
		},
		"TimeSignalPtsAdjustmentWrapping": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{
						PTSTime: ptr(uint64(8589844592)),
					},
				},
				PTSAdjustment: 180000,
			},
			expected: 90000,
		},
		"TimeSignalNoSpliceTime": {
			sis: scte35.SpliceInfoSection{
				SpliceCommand: &scte35.TimeSignal{
					SpliceTime: scte35.SpliceTime{},
				},
			},
			expected: 0,
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			require.Equal(t, c.expected, c.sis.SpliceTimePTS())
		})
	}
}
