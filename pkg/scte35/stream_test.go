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

func TestStreamSCTE35(t *testing.T) {
	cases := []struct {
		name string
		want []byte
		arg  string
	}{
		
		{
			name: "Stream SCTE35 Test test/test1.ts",
			want: []byte(`{"encryptedPacket":{},"protocolVersion":0,"ptsAdjustment":0,"sapType":3,"spliceCommand":{"type":5,"program":{"spliceTime":{"ptsTime":3438281293}},"spliceEventId":94,"spliceEventCancelIndicator":false,"spliceImmediateFlag":false,"outOfNetworkIndicator":false},"spliceDescriptors":[{"type":1,"preroll":177,"chars":"121#"}],"tier":4095}`),
			arg:  "test/test1.ts"},
        {
			name: "Stream SCTE35 Test test/test2.ts",
			want: []byte(`{"encryptedPacket":{},"protocolVersion":0,"ptsAdjustment":183265,"sapType":3,"spliceCommand":{"type":5,"program":{"spliceTime":{}},"spliceEventId":6111,"spliceEventCancelIndicator":false,"spliceImmediateFlag":true,"outOfNetworkIndicator":false},"tier":4095}`),
			arg:  "test/test2.ts"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			strm := MakeStream(c.arg)
			cue := &strm.Cues[0]
			got, _ := json.Marshal(cue)
			out, _ := AreEqualJSON(c.want, got)
			if !out {
				t.Errorf("Wanted:\n %s\nGot: %s", c.want, got)
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
			name: "Stream Packet Data Test test/test1.ts",
			want: PacketData{PacketNumber: 2283, PID: 515, Program: 51, PCR: 38199.918911, PTS: 38199.872111},
			arg:  "test/test1.ts"},
		{
			name: "Stream Packet Data Test test/test2.ts",
			want: PacketData{PacketNumber: 573, PID: 258, Program: 1, PCR: 595.9, PTS: 95443.717677},
			arg:  "test/test2.ts"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			strm := MakeStream(c.arg)
			got := strm.Cues[0].PacketData
			if got != c.want {
				t.Errorf("Wanted:\n %v\nGot: %v", c.want, got)
			}

		})
	}
}
