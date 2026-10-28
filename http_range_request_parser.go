package http_range_request_parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// RangeSpec describes a single byte range within a multipart Range header value.
// Start and End are byte offsets inclusive, matching RFC 7233 semantics.
// A suffix range ("-500") is represented as Start=0 and End with SuffixLength
// set to the number of trailing bytes requested.
type RangeSpec struct {
	Start         int64
	End           int64
	SuffixLength  int64
	IsSuffixRange bool
}

// ParsedRange is the result of parsing a Range header.
type ParsedRange struct {
	Unit   string
	Ranges []RangeSpec
}

// ErrInvalidRange is returned when the Range header value is malformed
// or does not satisfy RFC 7233 syntax.
var ErrInvalidRange = errors.New("invalid Range header")

// ParseRange parses an HTTP Range header value into byte offset tuples.
//
// The returned ParsedRange contains the unit ("bytes") and a slice of
// individual range specifications. Ranges are not normalized or merged —
// they are returned in the order they appear in the header.
//
// This function validates syntax only. It does not clamp ranges against
// a resource length; that is the caller's responsibility because the
// parser has no knowledge of the resource size.
func ParseRange(header string) (*ParsedRange, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return nil, fmt.Errorf("%w: empty header", ErrInvalidRange)
	}

	unitSplit := strings.SplitN(header, "=", 2)
	if len(unitSplit) != 2 {
		return nil, fmt.Errorf("%w: missing unit separator '='", ErrInvalidRange)
	}

	unit := strings.TrimSpace(unitSplit[0])
	if unit != "bytes" {
		return nil, fmt.Errorf("%w: unsupported unit %q", ErrInvalidRange, unit)
	}

	rangeSet := strings.TrimSpace(unitSplit[1])
	if rangeSet == "" {
		return nil, fmt.Errorf("%w: empty range set", ErrInvalidRange)
	}

	parts := strings.Split(rangeSet, ",")
	ranges := make([]RangeSpec, 0, len(parts))

	for _, part := range parts {
		rangeSpec, err := parseSingleRange(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, rangeSpec)
	}

	return &ParsedRange{Unit: unit, Ranges: ranges}, nil
}

// parseSingleRange parses one range specification like "0-499", "500-",
// or "-500".
func parseSingleRange(s string) (RangeSpec, error) {
	if s == "" {
		return RangeSpec{}, fmt.Errorf("%w: empty range spec", ErrInvalidRange)
	}

	dashIdx := strings.Index(s, "-")
	if dashIdx == -1 {
		return RangeSpec{}, fmt.Errorf("%w: missing '-' in %q", ErrInvalidRange, s)
	}

	startStr := s[:dashIdx]
	endStr := s[dashIdx+1:]

	if startStr == "" {
		// Suffix range: -N means last N bytes
		if endStr == "" {
			return RangeSpec{}, fmt.Errorf("%w: suffix range has no length", ErrInvalidRange)
		}
		length, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil || length <= 0 {
			return RangeSpec{}, fmt.Errorf("%w: invalid suffix length %q", ErrInvalidRange, endStr)
		}
		return RangeSpec{
			Start:         0,
			End:           0,
			SuffixLength:  length,
			IsSuffixRange: true,
		}, nil
	}

	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil || start < 0 {
		return RangeSpec{}, fmt.Errorf("%w: invalid start %q", ErrInvalidRange, startStr)
	}

	if endStr == "" {
		// Open-ended range: start- means from start to end of resource
		return RangeSpec{Start: start, End: -1}, nil
	}

	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil || end < 0 {
		return RangeSpec{}, fmt.Errorf("%w: invalid end %q", ErrInvalidRange, endStr)
	}

	if end < start {
		return RangeSpec{}, fmt.Errorf("%w: end %d precedes start %d", ErrInvalidRange, end, start)
	}

	return RangeSpec{Start: start, End: end}, nil
}
