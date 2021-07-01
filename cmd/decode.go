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
	"strings"

	"github.com/Comcast/scte35-go/pkg/scte35"
	"github.com/spf13/cobra"
)

// coreCommand returns the command for `scte35 decode`
func decodeCommand() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "decode",
		Short: "Decode a splice_info_section from binary",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires a binary signal")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			bin := args[0]
			var sis *scte35.SpliceInfoSection
			var err error

			// decode payload
			if strings.HasPrefix(bin, "0x") {
				sis, err = scte35.DecodeHex(bin)
			} else {
				sis, err = scte35.DecodeBase64(bin)
			}

			// print details (sis is never nil)
			switch format {
			case "json":
				b, _ := json.MarshalIndent(sis, "", "  ")
				_, _ = fmt.Fprintf(os.Stdout, "%s\n", b)
			case "xml":
				b, _ := xml.MarshalIndent(sis, "", "  ")
				_, _ = fmt.Fprintf(os.Stdout, "%s\n", b)
			default:
				_, _ = fmt.Fprintf(os.Stdout, "%s\n", sis)
			}

			// and any errors
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}
		},
	}
	cmd.PersistentFlags().StringVar(&format, "out", "text", "specify alternative output format (json, xml, text)")
	return cmd
}
