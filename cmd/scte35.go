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

// Package cmd contains code related to the scte35 command line interface.
package cmd

import "github.com/spf13/cobra"

// SCTE35 returns the root command.
func SCTE35() *cobra.Command {
	c := &cobra.Command{
		Use:   "scte35",
		Short: "SCTE-35 CLI",
	}

	c.AddCommand(decodeCommand())
	c.AddCommand(encodeCommand())
	c.AddCommand(encodeFileCommand())
	return c
}

