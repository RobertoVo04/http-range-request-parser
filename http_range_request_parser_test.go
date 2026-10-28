package http_range_request_parser

import (
	"errors"
	"testing"
)

func TestParseRange_SingleRange(t *testing.T) {
	r, err := ParseRange("bytes=0-499")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Unit != "bytes" {
		t.Fatalf("unit = %q, want bytes", r.Unit)
	}
	if len(r.Ranges) != 1 {
		t.Fatalf("ranges len = %d, want 1", len(r.Ranges))
	}
	want := RangeSpec{Start: 0, End: 499}
	if r.Ranges[0] != want {
		t.Fatalf("range = %+v, want %+v", r.Ranges[0], want)
	}
}

func TestParseRange_MultipleRanges(t *testing.T) {
	r, err := ParseRange("bytes=0-499,1000-1999,3000-3999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Ranges) != 3 {
		t.Fatalf("ranges len = %d, want 3", len(r.Ranges))
	}
	expected := []RangeSpec{
		{Start: 0, End: 499},
		{Start: 1000, End: 1999},
		{Start: 3000, End: 3999},
	}
	for i, want := range expected {
		if r.Ranges[i] != want {
			t.Fatalf("range[%d] = %+v, want %+v", i, r.Ranges[i], want)
		}
	}
}

func TestParseRange_OpenEnded(t *testing.T) {
	r, err := ParseRange("bytes=500-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Ranges) != 1 {
		t.Fatalf("ranges len = %d, want 1", len(r.Ranges))
	}
	want := RangeSpec{Start: 500, End: -1}
	if r.Ranges[0] != want {
		t.Fatalf("range = %+v, want %+v", r.Ranges[0], want)
	}
}

func TestParseRange_SuffixRange(t *testing.T) {
	r, err := ParseRange("bytes=-500")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Ranges) != 1 {
		t.Fatalf("ranges len = %d, want 1", len(r.Ranges))
	}
	want := RangeSpec{Start: 0, End: 0, SuffixLength: 500, IsSuffixRange: true}
	if r.Ranges[0] != want {
		t.Fatalf("range = %+v, want %+v", r.Ranges[0], want)
	}
}

func TestParseRange_MixedRanges(t *testing.T) {
	r, err := ParseRange("bytes=0-99,500-,-200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Ranges) != 3 {
		t.Fatalf("ranges len = %d, want 3", len(r.Ranges))
	}
	expected := []RangeSpec{
		{Start: 0, End: 99},
		{Start: 500, End: -1},
		{Start: 0, End: 0, SuffixLength: 200, IsSuffixRange: true},
	}
	for i, want := range expected {
		if r.Ranges[i] != want {
			t.Fatalf("range[%d] = %+v, want %+v", i, r.Ranges[i], want)
		}
	}
}

func TestParseRange_Empty(t *testing.T) {
	_, err := ParseRange("")
	if err == nil {
		t.Fatalf("expected error for empty header")
	}
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("err = %v, want ErrInvalidRange", err)
	}
}

func TestParseRange_MissingUnit(t *testing.T) {
	_, err := ParseRange("0-499")
	if err == nil {
		t.Fatalf("expected error for missing unit")
	}
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("err = %v, want ErrInvalidRange", err)
	}
}

func TestParseRange_UnsupportedUnit(t *testing.T) {
	_, err := ParseRange("lines=0-10")
	if err == nil {
		t.Fatalf("expected error for unsupported unit")
	}
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("err = %v, want ErrInvalidRange", err)
	}
}

func TestParseRange_EmptyRangeSet(t *testing.T) {
	_, err := ParseRange("bytes=")
	if err == nil {
		t.Fatalf("expected error for empty range set")
	}
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("err = %v, want ErrInvalidRange", err)
	}
}

func TestParseRange_NoDash(t *testing.T) {
	_, err := ParseRange("bytes=0to499")
	if err == nil {
		t.Fatalf("expected error for missing dash")
	}
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("err = %v, want ErrInvalidRange", err)
	}
}

func TestParseRange_EndBeforeStart(t *testing.T) {
	_, err := ParseRange("bytes=1000-500")
	if err == nil {
		t.Fatalf("expected error for end < start")
	}
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("err = %v, want ErrInvalidRange", err)
	}
}

func TestParseRange_InvalidStart(t *testing.T) {
	_, err := ParseRange("bytes=abc-500")
	if err == nil {
		t.Fatalf("expected error for non-numeric start")
	}
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("err = %v, want ErrInvalidRange", err)
	}
}

func TestParseRange_InvalidEnd(t *testing.T) {
	_, err := ParseRange("bytes=0-xyz")
	if err == nil {
		t.Fatalf("expected error for non-numeric end")
	}
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("err = %v, want ErrInvalidRange", err)
	}
}

func TestParseRange_EmptySuffixRange(t *testing.T) {
	_, err := ParseRange("bytes=-")
	if err == nil {
		t.Fatalf("expected error for empty suffix range")
	}
	if !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("err = %v, want ErrInvalidRange", err)
	}
}

func TestParseRange_LeadingWhitespaceInUnit(t *testing.T) {
	r, err := ParseRange("  bytes=0-99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Unit != "bytes" {
		t.Fatalf("unit = %q, want bytes", r.Unit)
	}
}

func TestParseRange_WhitespaceAroundRanges(t *testing.T) {
	r, err := ParseRange("bytes=0-99 , 200-299")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Ranges) != 2 {
		t.Fatalf("ranges len = %d, want 2", len(r.Ranges))
	}
}
