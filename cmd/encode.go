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
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/Comcast/scte35-go/pkg/scte35"
	"github.com/spf13/cobra"
)

// encodeCommand returns the command for `scte35 encode`
func encodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encode < filename or encode {\"protocolVersion\"... ",
		Short: "Encode a splice_info_section to binary being provided from stdin or as a parameter",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("invalid number of parameter provided")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var input string
			var sis *scte35.SpliceInfoSection

			if len(args) == 1 {
				input = args[0]
			} else {
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					input = input + scanner.Text()
				}
				err = scanner.Err()
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
					return
				}
			}

			if strings.HasPrefix(strings.TrimSpace(input), "<") {
				err = xml.Unmarshal([]byte(input), &sis)
			} else if strings.HasPrefix(strings.TrimSpace(input), "{") {
				err = json.Unmarshal([]byte(input), &sis)
			} else {
				err = fmt.Errorf("unrecognized or empty input")
			}

			if err == nil {
				// print encoded signal
				_, _ = fmt.Fprintf(os.Stdout, "Base64: %s\n", sis.Base64())
				_, _ = fmt.Fprintf(os.Stdout, "Hex   : %s\n", sis.Hex())
			} else {
				// print error
				_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}

		},
	}
	return cmd
}
