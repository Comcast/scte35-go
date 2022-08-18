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

	"github.com/Comcast/scte35-go/pkg/scte35"
	"github.com/spf13/cobra"
)

// encodeCommand returns the command for `scte35 encode`
func encodeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encode < filename",
		Short: "Encode a splice_info_section to binary being provided from stdin",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("additional command line arguments are not needed")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var input string
			var sis *scte35.SpliceInfoSection

			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				input = input + scanner.Text()
			}
			err = scanner.Err()
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				return
			}

			if len(input) == 0 {
				_, _ = fmt.Fprintf(os.Stderr, "Error: empty input\n")
				return
			}

			if input[0] == '<' {
				err = xml.Unmarshal([]byte(input), &sis)
			} else if input[0] == '{' {
				err = json.Unmarshal([]byte(input), &sis)
			} else {
				err = fmt.Errorf("unrecognized input")
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
