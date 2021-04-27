// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace

import (
  "io"
	_ "unsafe"
)


// reads trace from stderr (io.reader) and parse
func ParseTrace(r io.Reader, binary string) (*ParseResult, error) {
	parseResult, err := Parse(r,binary)
	if err != nil {
		return nil, err
	}

	err = Symbolize(parseResult.Events, binary)

	return &parseResult, err
}
