# http-range-request-parser

Parses HTTP `Range` header values into byte offset tuples, including multipart (multi-range) requests.

```go
package main

import (
    "fmt"
    "http-range-request-parser"
)

func main() {
    pr, err := http_range_request_parser.ParseRange("bytes=0-499,1000-")
    if err != nil {
        panic(err)
    }
    for _, r := range pr.Ranges {
        if r.IsSuffixRange {
            fmt.Printf("last %d bytes\n", r.SuffixLength)
        } else if r.End == -1 {
            fmt.Printf("from %d to end\n", r.Start)
        } else {
            fmt.Printf("%d-%d\n", r.Start, r.End)
        }
    }
}
```

## Why

HTTP range parsing is fiddly: the `Range` header has three legal shapes (`start-end`, `start-`, `-suffix`), they can be chained with commas, and the syntax is strict enough that hand-rolled parsers miss edge cases. This library exists to handle that parsing in one place and return structured data the caller can clamp against a resource length.

The trade-off: this library validates syntax only. It does not know your resource size, so it does not clamp, merge, or reject ranges that exceed the resource. That is the caller's job — the parser returns offsets, the caller applies them to bytes.

## Edge cases

- `bytes=-500` is a suffix range: it means "the last 500 bytes". It is represented with `IsSuffixRange: true` and `SuffixLength: 500`, not as a normal start/end pair.
- `bytes=500-` is open-ended: `End` is set to `-1` to signal "to the end of the resource".
- Ranges are returned in the order they appear; no merging or deduplication is performed.
- Only the `bytes` unit is accepted; anything else returns `ErrInvalidRange`.
