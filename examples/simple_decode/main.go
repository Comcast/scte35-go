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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/Comcast/scte35-go/pkg/scte35"
)

func main() {
	sis, _ := scte35.DecodeBase64("/DA8AAAAAAAAAP///wb+06ACpQAmAiRDVUVJAACcHX//AACky4AMEERJU0NZTVdGMDQ1MjAwMEgxAQEMm4c0")

	// details
	_, _ = fmt.Fprintf(os.Stdout, "\nTable: \n%s\n", sis.Table("", "\t"))

	// xml
	b, _ := xml.MarshalIndent(sis, "", "\t")
	_, _ = fmt.Fprintf(os.Stdout, "\nXML: \n%s\n", b)

	// json
	b, _ = json.MarshalIndent(sis, "", "\t")
	_, _ = fmt.Fprintf(os.Stdout, "\nJSON: \n%s\n", b)
}
