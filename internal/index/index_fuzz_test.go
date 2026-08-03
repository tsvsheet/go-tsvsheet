package index_test

import (
	"bytes"
	"testing"

	"github.com/tsvsheet/go-tsvsheet/internal/index"
)

// FuzzScanAgreesWithTheOracle holds the index to the naive whole-file truth
// over arbitrary bytes: identical census, and every row locatable at a
// checkpoint whose offset is exactly that row's line start.
func FuzzScanAgreesWithTheOracle(f *testing.F) {
	for _, src := range corpus {
		f.Add([]byte(src), 3)
	}
	f.Fuzz(func(t *testing.T, data []byte, stride int) {
		if stride < 1 || stride > 1024 {
			return
		}
		ix, err := index.Scan(bytes.NewReader(data), index.SourceSize(len(data)), index.Options{
			Stride:       index.StrideLines(stride),
			Split:        splitTerminators,
			IsComment:    isComment,
			MaxLineBytes: 1 << 20,
		})
		if err != nil {
			t.Fatalf("scan refused fuzz input under a generous line bound: %v", err)
		}
		truth := oracleForFuzz(t, string(data))
		if ix.Census() != truth.census {
			t.Fatalf("census diverged: %+v vs oracle %+v", ix.Census(), truth.census)
		}
		for row := 0; row < truth.census.Rows; row++ {
			cp, ok := ix.Locate(index.GridRow(row))
			if !ok || int64(cp.Offset) != truth.rowOffsets[cp.Row] || int(cp.Row) > row {
				t.Fatalf("locate(%d) broke its contract: %+v ok=%v", row, cp, ok)
			}
		}
	})
}

// oracleForFuzz adapts the oracle to the fuzz body, where a failing oracle is
// a fatal harness defect rather than a finding.
func oracleForFuzz(t *testing.T, src string) oracleDoc { return oracle(t, src) }
