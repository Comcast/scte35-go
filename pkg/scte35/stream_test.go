package scte35

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func AreEqualJSON(s1, s2 []byte) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}

func MakeStream(arg string) Stream {
	var strm Stream
	strm.Silent = true
	strm.Decode(arg)

	return strm
}

// test for SCTE35 data in MPEGTS files
func TestStreamSCTE35(t *testing.T) {
	cases := []struct {
		name string
		want []byte
		arg  string
	}{

		{
			name: "Stream SCTE35 Test Splice Insert w/ DTMF Descriptor : test/test1.ts",
			want: []byte(`{"encryptedPacket":{},"protocolVersion":0,"ptsAdjustment":0,"sapType":3,"spliceCommand":{"type":5,"program":{"spliceTime":{"ptsTime":3438281293}},"spliceEventId":94,"spliceEventCancelIndicator":false,"spliceImmediateFlag":false,"outOfNetworkIndicator":false},"spliceDescriptors":[{"type":1,"preroll":177,"chars":"121#"}],"tier":4095}`),
			arg:  "test/test1.ts"},
		{
			name: "Stream SCTE35 Test Splice Insert: test/test2.ts",
			want: []byte(`{"encryptedPacket":{},"protocolVersion":0,"ptsAdjustment":183265,"sapType":3,"spliceCommand":{"type":5,"program":{"spliceTime":{}},"spliceEventId":6111,"spliceEventCancelIndicator":false,"spliceImmediateFlag":true,"outOfNetworkIndicator":false},"tier":4095}`),
			arg:  "test/test2.ts"},
		{
			name: "Stream SCTE35 Test Time Signal w/ Segmentation Descriptor : test/test3.ts",
			want: []byte(`{"encryptedPacket":{"cwIndex":255},"protocolVersion":0,"ptsAdjustment":0,"sapType":3,"spliceCommand":{"type":6,"spliceTime":{"ptsTime":8098574552}},"spliceDescriptors":[{"type":2,"deliveryRestrictions":{"archiveAllowedFlag":true,"webDeliveryAllowedFlag":true,"noRegionalBlackoutFlag":true,"deviceRestrictions":3},"segmentationUpids":[{"segmentationUpidType":8,"formatIdentifier":null,"format":"text","value":"749038837"}],"components":null,"segmentationEventId":1207959560,"segmentationEventCancelIndicator":false,"segmentationDuration":null,"segmentationTypeId":23,"segmentNum":0,"segmentsExpected":0,"subSegmentNum":null,"subSegmentsExpected":null}],"tier":4095}`),
			arg:  "test/test3.ts"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			strm := MakeStream(c.arg)
			if len(strm.Cues) > 0 {
				cue := &strm.Cues[0]
				got, _ := json.Marshal(cue)
				out, _ := AreEqualJSON(c.want, got)
				if !out {
					t.Errorf("\n%s\nWanted:\n %s\nGot:\n%s",c.name, c.want, got)
				}
			} else {
				t.Errorf("\nNo Cues in Stream %s", c.arg)
			}
		})
	}
}

func TestPacketData(t *testing.T) {
	cases := []struct {
		name string
		want PacketData
		arg  string
	}{
		{
			name: "Stream Packet Data Test:  test/test1.ts",
			want: PacketData{PacketNumber: 2283, PID: 515, Program: 51, PCR: 38199.918911, PTS: 38200.756444},
			arg:  "test/test1.ts"},
		{
			name: "Stream Packet Data Test:  test/test2.ts",
			want: PacketData{PacketNumber: 573, PID: 258, Program: 1, PCR: 595.9, PTS: 596.684688},
			arg:  "test/test2.ts"},

		{
			name: "Stream Packet Data Test:  test/test3.ts",
			want: PacketData{PacketNumber: 1672, PID: 309, Program: 1, PCR: 89975.5132, PTS: 89978.421666},
			arg:  "test/test3.ts"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			strm := MakeStream(c.arg)
			if len(strm.Cues) > 0 {
				got := strm.Cues[0].PacketData
				if got != c.want {
					t.Errorf("\n%s\nWanted:\n %v\nGot:\n%v",c.name, c.want, got)
				}
			} else {
				t.Errorf("\nNo Cues in Stream %s", c.arg)
			}
		})
	}
}
