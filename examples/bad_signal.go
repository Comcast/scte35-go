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
	sis, err := scte35.DecodeBase64("FkC1lwP3uTQD0VvxHwVBEH89G6B7VjzaZ9eNuyUF9q8pYAIXsRM9ZpDCczBeDbytQhXkssQstGJVGcvjZ3tiIMULiA4BpRHlzLGFa0q6aVMtzk8ZRUeLcxtKibgVOKBBnkCbOQyhSflFiDkrAAIp+Fk+VRsByTSkPN3RvyK+lWcjHElhwa9hNFcAy4dm3DdeRXnrD3I2mISNc7DkgS0ReotPyp94FV77xMHT4D7SYL48XU20UM4bgg==")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error: %s\n", err)
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s\n", sis)
}
