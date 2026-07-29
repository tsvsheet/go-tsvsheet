package engine_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tsvsheet/go-tsvsheet/internal/engine"
)

func TestDate_TodayNowInjectedClock(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 7, 14, 15, 30, 45, 0, time.UTC)
	s, err := engine.Parse([]byte("=today()\t=now()\t=if(today(), 1, 0)\n"))
	require.NoError(t, err)
	g := s.ComputeAt(at)
	assert.Equal(t, "2026-07-14", g[0][0])          // date renders ISO
	assert.Equal(t, "2026-07-14 15:30:45", g[0][1]) // datetime renders with time
	assert.Equal(t, "1", g[0][2])                   // a date is truthy
}

func TestDate_ClockArity(t *testing.T) {
	t.Parallel()

	assert.Equal(t, string(engine.ErrValue), formula1(t, "today(5)")) // no arguments allowed
	assert.Equal(t, string(engine.ErrValue), formula1(t, "now(5)"))
}

func TestDate_Components(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"year(date(2026, 7, 14))":    "2026",
		"month(date(2026, 7, 14))":   "7",
		"day(date(2026, 7, 14))":     "14",
		"weekday(date(2026, 7, 14))": "3",  // Tuesday (Sunday = 1)
		"hour(45000.5)":              "12", // half a day
		"minute(45000.25)":           "0",
		"second(45000)":              "0",
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, formula1(t, expr))
		})
	}
}

func TestDate_Constructors(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"date(2026, 7, 14)":                         "2026-07-14",
		"date(2026, 13, 1)":                         "2027-01-01", // month normalizes
		"edate(date(2026, 7, 14), 2)":               "2026-09-14",
		"edate(date(2026, 1, 31), 1)":               "2026-03-03", // Feb overflow, like Go
		"eomonth(date(2026, 2, 10), 0)":             "2026-02-28",
		"days(date(2026, 7, 14), date(2026, 7, 4))": "10",
		`datevalue("2020-06-15")`:                   "2020-06-15",
	}
	for expr, want := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, want, formula1(t, expr))
		})
	}
}

func TestDate_Errors(t *testing.T) {
	t.Parallel()

	v := "#VALUE!"
	cases := []string{
		`=year(A1)`, `=date(A1, 1, 1)`, `=edate(A1, 1)`, `=edate(45000, A1)`,
		`=eomonth(A1, 0)`, `=days(A1, 45000)`, `=days(45000, A1)`,
	}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, v, cellAt(t, compute(t, "hi\t"+expr+"\n"), 0, 1))
		})
	}
	assert.Equal(t, v, formula1(t, `datevalue("not a date")`))
}

// isnowAt computes a single ISNOW formula against a fixed instant.
func isnowAt(t *testing.T, formula string, at time.Time) string {
	t.Helper()
	s, err := engine.Parse([]byte("=" + formula + "\n"))
	require.NoError(t, err)
	return s.ComputeAt(at)[0][0]
}

func TestIsnow_MatchesTheClock(t *testing.T) {
	t.Parallel()

	noon := time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC) // a Monday, 12:00
	assert.Equal(t, "TRUE", isnowAt(t, `isnow("noon")`, noon))
	assert.Equal(t, "FALSE", isnowAt(t, `isnow("midnight")`, noon))
}

func TestIsnow_Errors(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, "#VALUE!", isnowAt(t, `isnow()`, at))      // arity
	assert.Equal(t, "#VALUE!", isnowAt(t, `isnow("$$$")`, at)) // malformed pattern
	assert.Equal(t, "#DIV/0!", isnowAt(t, `isnow(1/0)`, at))   // argument error propagates
}

func TestIsnow_IsKnownToCheck(t *testing.T) {
	t.Parallel()

	// ISNOW is a known function — Check must not flag it.
	s, err := engine.Parse([]byte(`=isnow("M-F noon")` + "\n"))
	require.NoError(t, err)
	assert.Empty(t, engine.Check(s))
}
