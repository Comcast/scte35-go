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
		EncryptedPacket: scte35.EncryptedPacket{
			EncryptionAlgorithm: scte35.EncryptionAlgorithmNone,
			CWIndex:             255,
		},
		Tier:    4095,
		SAPType: 3,
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
	_, _ = fmt.Fprintf(os.Stdout, "Original:\n")
	_, _ = fmt.Fprintf(os.Stdout, "base-64: %s\n", sis.Base64())
	_, _ = fmt.Fprintf(os.Stdout, "hex    : %s\n", sis.Hex())
}
