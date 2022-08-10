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

package cmd

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/Comcast/scte35-go/pkg/scte35"
	"github.com/spf13/cobra"
)

// encodeFileCommand returns the command for `scte35 encodefile`
func encodeFileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encodefile filename",
		Short: "Encode a splice_info_section read from input file to binary",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("filename must be specified\n")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			bin, err := os.ReadFile(args[0])
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "could not read file %v because %v", bin, err)
			}

			var sis *scte35.SpliceInfoSection

			if bin[0] == '<' {
				err = xml.Unmarshal(bin, &sis)
			} else {
				err = json.Unmarshal(bin, &sis)
			}

			// print encoded signal
			_, _ = fmt.Fprintf(os.Stdout, "Base64: %s\n", sis.Base64())
			_, _ = fmt.Fprintf(os.Stdout, "Hex   : %s\n", sis.Hex())

			// and any errors
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}
		},
	}
	return cmd
}
