# scte35-go: ANSI/SCTE 35 Decoder/Encoder 

`scte35-go` is a Go library to supports creating, decorating, and analyzing 
binary Digital Program Insertion Cueing Messages.

This library is fully compliant and compatible with all versions of the 
[ANSI/SCTE 35](https://www.scte.org/standards-development/library/standards-catalog/scte-35-2019/) 
specification up to and including [ANSI/SCTE 35 2020](./docs/SCTE-35_2020.pdf).

This project uses [Semantic Versioning](https://semver.org) and is published as
a [Go Module](https://blog.golang.org/using-go-modules).

![Build Status](https://github.com/Comcast/scte35-go/actions/workflows/build.yml/badge.svg)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Comcast/scte35-go)](https://pkg.go.dev/github.com/Comcast/scte35-go)

## Getting Started

Get the module:

```shell
$ go get github.com/Comcast/scte35-go
go get: added github.com/Comcast/scte35-go v1.0.0
```

## Code Examples

Additional examples can be found in [examples](./examples).

#### Decode Signal

Binary signals can be quickly and easily decoded from base-64 or hexadecimal
strings.

The results can be output as a:
* String - emulating the table structure used in the [SCTE 35 specification](./docs/ANSI_SCTE-35-2020.pdf).
* XML - compliant with the [SCTE 35 XML Schema](./docs/scte35.xsd)
* JSON - for integrating with JSON based tools such as [jq](https://stedolan.github.io/jq/)

[decode.go](../examples/decode.go)

```go
package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/Comcast/scte35-go/pkg/scte35"
)

func main() {
	sis, _ := scte35.DecodeBase64("/DA8AAAAAAAAAP///wb+06ACpQAmAiRDVUVJAACcHX//AACky4AMEERJU0NZTVdGMDQ1MjAwMEgxAQEMm4c0")

	// details
	_, _ = fmt.Fprintf(os.Stdout, "\nDetails: \n%s\n", sis)

	// xml
	b, _ := xml.MarshalIndent(sis, "", "\t")
	_, _ = fmt.Fprintf(os.Stdout, "\nXML: \n%s\n", b)

	// json
	b, _ = json.MarshalIndent(sis, "", "\t")
	_, _ = fmt.Fprintf(os.Stdout, "\nJSON: \n%s\n", b)
}
```

```shell
$ go run ./examples/decode.go

Details:
splice_info_section() {
    table_id: 0xfc
    section_syntax_indicator: false
    private_indicator: false
    sap_type: Not Specified
    section_length: 60 bytes
}
protocol_version: 0
encrypted_packet: false
encryption_algorithm: No encryption
pts_adjustment: 0 ticks (0s)
cw_index: 0
tier: 4095
splice_command_length: 5 bytes
splice_command_type: 0x06
time_signal() {
    time_specified_flag: true
    pts_time: 3550479013 ticks (10h57m29.766811111s)
}
descriptor_loop_length: 38 bytes
avail_descriptor() {
    splice_descriptor_tag: 0x02
    descriptor_length: 36 bytes
    identifier: CUEI
    segmentation_event_id: 39965
    segmentation_event_cancel_indicator: false
    program_segmentation_flag: true
    segmentation_duration_flag: true
    delivery_not_restricted_flag: true
    segmentation_duration: 10800000 ticks (2m0s)
    segmentation_upid_length: 16 bytes
    segmentation_upid {
        segmentation_upid_type: MPU() (0x0c)
        format_identifer: DISC
        segmentation_upid: 0x44495343594d57463034353230303048
    }
    segmentation_type_id: Provider Advertisement End (0x31)
    segment_num: 1
    segments_expected: 1
}

XML:
<SpliceInfoSection xmlns="http://www.scte.org/schemas/35" sapType="3" ptsAdjustment="0" protocolVersion="0" tier="4095">
  <TimeSignal xmlns="http://www.scte.org/schemas/35">
    <SpliceTime xmlns="http://www.scte.org/schemas/35" ptsTime="3550479013"></SpliceTime>
  </TimeSignal>
  <SegmentationDescriptor xmlns="http://www.scte.org/schemas/35" segmentationEventId="39965" segmentationEventCancelIndicator="false" segmentationDuration="10800000" segmentationTypeId="49" segmentNum="1" segmentsExpected="1">
    <SegmentationUpid xmlns="http://www.scte.org/schemas/35" segmentationUpidType="12" formatIdentifier="1145656131" format="base-64">WU1XRjA0NTIwMDBI</SegmentationUpid>
  </SegmentationDescriptor>
</SpliceInfoSection>

JSON:
{
  "protocolVersion": 0,
  "ptsAdjustment": 0,
  "sapType": 3,
  "spliceCommand": {
    "type": 6,
    "spliceTime": {
      "ptsTime": 3550479013
    }
  },
  "spliceDescriptors": [
    {
      "type": 2,
      "deliveryRestrictions": null,
      "segmentationUpids": [
        {
          "segmentationUpidType": 12,
          "formatIdentifier": 1145656131,
          "format": "base-64",
          "value": "WU1XRjA0NTIwMDBI"
        }
      ],
      "components": null,
      "segmentationEventId": 39965,
      "segmentationEventCancelIndicator": false,
      "segmentationDuration": 10800000,
      "segmentationTypeId": 49,
      "segmentNum": 1,
      "segmentsExpected": 1,
      "subSegmentNum": null,
      "subSegmentsExpected": null
    }
  ],
  "tier": 4095
}
```

#### Encode Signal

Encoding signals is equally simple. You can start from scratch and build a
`scte35.SpliceInfoSection` or decode an existing signal and modify it to suit 
your needs.

[encode.go](./examples/encode.go)

```go
package main

import (
	"fmt"
	"os"

	"github.com/Comcast/scte35-go/pkg/scte35"
)

func main() {
	// start with a signal
	sis := scte35.SpliceInfoSection{
		SpliceCommand: scte35.NewTimeSignal(0x072bd0050),
		SpliceDescriptors: []scte35.SpliceDescriptor{
			&scte35.SegmentationDescriptor{
				DeliveryRestrictions: &scte35.DeliveryRestrictions{
					NoRegionalBlackoutFlag: true,
					ArchiveAllowedFlag:     true,
					DeviceRestrictions:     scte35.DeviceRestrictionsNone,
				},
				SegmentationEventID: uint32(0x4800008e),
				SegmentationTypeID:  scte35.SegmentationTypeProviderPOStart,
				SegmentationUPIDs: []scte35.SegmentationUPID{
					scte35.NewSegmentationUPID(scte35.SegmentationUPIDTypeTI, []byte("78511452")),
				},
				SegmentNum: 2,
			},
		},
		EncryptedPacket: scte35.EncryptedPacket{CWIndex: 255},
		Tier:            4095,
		SAPType:         3,
	}

	// encode it
	_, _ = fmt.Fprintf(os.Stdout, "Original:\n")
	_, _ = fmt.Fprintf(os.Stdout, "base-64: %s\n", sis.Base64())
	_, _ = fmt.Fprintf(os.Stdout, "hex    : %s\n", sis.Hex())

	// add a segmentation descriptor
	sis.SpliceDescriptors = append(
		sis.SpliceDescriptors,
		&scte35.DTMFDescriptor{
			DTMFChars: "ABC*",
		},
	)

	// encode it again
	_, _ = fmt.Fprintf(os.Stdout, "\nModified:\n")
	_, _ = fmt.Fprintf(os.Stdout, "base-64: %s\n", sis.Base64())
	_, _ = fmt.Fprintf(os.Stdout, "hex    : %s\n", sis.Hex())
}

```

```shell
$ go run ./examples/encode.go
Original:
base-64: /DAvAAAAAAAA///wBQb+cr0AUAAZAhdDVUVJSAAAjn+PCAg3ODUxMTQ1MjQCADhqB9E=
hex    : fc302f000000000000fffff00506fe72bd005000190217435545494800008e7f8f08083738353131343532340200386a07d1

Modified:
base-64: /DA7AAAAAAAA///wBQb+cr0AUAAlAhdDVUVJSAAAjn+PCAg3ODUxMTQ1MjQCAAEKQ1VFSQCfQUJDKqtwQlQ=
hex    : fc303b000000000000fffff00506fe72bd005000250217435545494800008e7f8f08083738353131343532340200010a43554549009f4142432aab704254
```

#### Decoding Non-Compliant Signals

The SCTE 35 decoder will always return a non-nil `SpliceInfoSection`, even when
an error occurs. This is done to help better identify the specific cause of the
decoding failure.

[bad-signal](./examples/bad_signal.go)

```go
package main

import (
	"fmt"
	"os"

	"code.comcast.com/jbaile223/scte35/pkg/scte35"
)

func main() {
	sis, err := scte35.DecodeBase64("FkC1lwP3uTQD0VvxHwVBEH89G6B7VjzaZ9eNuyUF9q8pYAIXsRM9ZpDCczBeDbytQhXkssQstGJVGcvjZ3tiIMULiA4BpRHlzLGFa0q6aVMtzk8ZRUeLcxtKibgVOKBBnkCbOQyhSflFiDkrAAIp+Fk+VRsByTSkPN3RvyK+lWcjHElhwa9hNFcAy4dm3DdeRXnrD3I2mISNc7DkgS0ReotPyp94FV77xMHT4D7SYL48XU20UM4bgg==")
	if err != nil {
		fmt.Fprintf(os.Stdout, "Error: %s\n", err)
	}
	// splice info section contains best-effort decoding
	fmt.Fprintf(os.Stdout, "Signal: %s", sis)
}
```

As we can see from the output below, the signal has a corrupted `component_count`,
causing the decoder to return a `scte.ErrBufferOverflow`:

```shell
$ go run ./examples/bad_signal.go
Error: splice_insert: buffer overflow
Signal: splice_info_section() {
    table_id: 0xfc
    section_syntax_indicator: false
    private_indicator: false
    sap_type: Type 1
    section_length: 343 bytes
}
protocol_version: 151
encrypted_packet: false
encryption_algorithm: No encryption
pts_adjustment: 8451077123 ticks (26h5m0.856922222s)
cw_index: 209
tier: 1471
splice_command_length: 326 bytes
splice_command_type: 0x05
splice_insert() {
    splice_event_id: 1091600189
    splice_event_cancel_indicator: false
    out_of_network_indicator: true
    program_splice_flag: false
    duration_flag: true
    splice_immediate_flag: false
    component_count: 123
    component[0] {
        component_tag: 86
        time_specified_flag: false
    }

    ... additional components removed

    component[122] {
        component_tag: 0
        time_specified_flag: false
    }
    auto_return: false
    duration: 0 ticks (0s)
    unique_program_id: 0
    avail_num: 0
    avails_expected: 0    
}
```

#### CRC_32 Validation

The SCTE 35 decoder performs automatic `CRC_32` validation. The returned error
can be explicitly ignored if desired.

```go
sis, err := scte35.DecodeBase64("/DA4AAAAAAAAAP/wFAUABDEAf+//mWEhzP4Azf5gAQAAAAATAhFDVUVJAAAAAX+/AQIwNAEAAKeYO3Q=")
if err != nil && !errors.Is(err, scte35.ErrCRC32Invalid) {
  return err
} 
```

#### Logging

Additional diagnostics can be enabled by redirecting the output of
`scte35.Logger`

```go
scte35.Logger.SetOutput(os.Stderr)
```

## Command Line Interface

This package also provides a simple command line interface that supports
encoding and decoding signals from the command line.

```shell
$ ./scte35-go --help
SCTE-35 CLI

Usage:
  scte35-go [command]

Available Commands:
  decode      Decode a splice_info_section from binary
  encode      Encode a splice_info_section to binary
  help        Help about any command

Flags:
  -h, --help   help for scte35-go
```

## License

`scte35-go` is licensed under [Apache License 2.0](/LICENSE.md). 

## Code of Conduct

We take our [code of conduct](CODE_OF_CONDUCT.md) very seriously. Please abide 
by it.

## Contributing

Please read our [contributing guide](CONTRIBUTING.md) for details on how to 
contribute to our project.

## Releases

* [Change Log](CHANGELOG.md)
