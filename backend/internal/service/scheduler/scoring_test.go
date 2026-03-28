package scheduler

import (
	"testing"
	"time"
)

// --- tierFromPopulation ---

func TestTierFromPopulation(t *testing.T) {
	cases := []struct {
		pop  int
		want Tier
	}{
		{2_000_000, TierA},
		{1_000_001, TierA},
		{1_000_000, TierB}, // boundary: not > 1M
		{750_000, TierB},
		{500_001, TierB},
		{500_000, TierC}, // boundary: not > 500K
		{300_000, TierC},
		{100_001, TierC},
		{100_000, TierD}, // boundary: not > 100K
		{50_000, TierD},
		{0, TierD},
	}
	for _, c := range cases {
		got := tierFromPopulation(c.pop)
		if got != c.want {
			t.Errorf("tierFromPopulation(%d) = %v, want %v", c.pop, got, c.want)
		}
	}
}

// --- normPair ---

func TestNormPair(t *testing.T) {
	if normPair(TierB, TierA) != (tierPair{TierA, TierB}) {
		t.Error("normPair should put smaller tier first")
	}
	if normPair(TierA, TierB) != (tierPair{TierA, TierB}) {
		t.Error("normPair should be idempotent when already ordered")
	}
	if normPair(TierC, TierC) != (tierPair{TierC, TierC}) {
		t.Error("normPair should handle equal tiers")
	}
}

// --- parseInterval ---

func TestParseInterval_AllBrackets(t *testing.T) {
	cases := []struct {
		t1, t2       Tier
		daysUntil    int
		wantInterval int
	}{
		// A↔A
		{TierA, TierA, 1, 1}, {TierA, TierA, 14, 1},
		{TierA, TierA, 15, 2}, {TierA, TierA, 45, 2},
		{TierA, TierA, 46, 4}, {TierA, TierA, 90, 4},
		// symmetry: B↔A == A↔B
		{TierB, TierA, 1, 1}, {TierA, TierB, 1, 1},
		// A↔D
		{TierA, TierD, 1, 3}, {TierA, TierD, 20, 7}, {TierA, TierD, 60, 21},
		// D↔D
		{TierD, TierD, 1, 7}, {TierD, TierD, 20, 21}, {TierD, TierD, 60, 60},
	}
	for _, c := range cases {
		got := parseInterval(c.t1, c.t2, c.daysUntil)
		if got != c.wantInterval {
			t.Errorf("parseInterval(%v,%v, days=%d) = %d, want %d",
				c.t1, c.t2, c.daysUntil, got, c.wantInterval)
		}
	}
}

func TestParseInterval_BracketBoundaries(t *testing.T) {
	// day 14 is still "near", day 15 is "mid"
	if parseInterval(TierA, TierA, 14) != parseInterval(TierA, TierA, 1) {
		t.Error("day 14 should use near bracket")
	}
	if parseInterval(TierA, TierA, 15) == parseInterval(TierA, TierA, 14) {
		t.Error("day 15 should switch to mid bracket")
	}
	// day 45 is still "mid", day 46 is "far"
	if parseInterval(TierA, TierA, 45) != parseInterval(TierA, TierA, 15) {
		t.Error("day 45 should use mid bracket")
	}
	if parseInterval(TierA, TierA, 46) == parseInterval(TierA, TierA, 45) {
		t.Error("day 46 should switch to far bracket")
	}
}

// --- shouldParse ---

func TestShouldParse_OutOfRange(t *testing.T) {
	today := truncateToDay(time.Now())
	conn := Connection{OriginPopulation: 2_000_000, DestPopulation: 2_000_000}

	if shouldParse(conn, today, today) {
		t.Error("daysUntil=0 should return false")
	}
	if shouldParse(conn, today, today.AddDate(0, 0, 91)) {
		t.Error("daysUntil=91 should return false")
	}
	if shouldParse(conn, today, today.AddDate(0, 0, -1)) {
		t.Error("past date should return false")
	}
}

func TestShouldParse_NeverParsed(t *testing.T) {
	today := truncateToDay(time.Now())
	conn := Connection{
		OriginPopulation: 2_000_000,
		DestPopulation:   2_000_000,
		LastParsedAt:     time.Time{}, // zero = never parsed
		LastUsedAt:       today,
	}
	// should always parse if never parsed before
	if !shouldParse(conn, today, today.AddDate(0, 0, 1)) {
		t.Error("never-parsed connection should always be enqueued")
	}
}

func TestShouldParse_IntervalNotElapsed(t *testing.T) {
	today := truncateToDay(time.Now())
	// A↔A near interval = 1 day; parsed today → should NOT parse again today
	conn := Connection{
		OriginPopulation: 2_000_000,
		DestPopulation:   2_000_000,
		LastParsedAt:     today,
		LastUsedAt:       today,
	}
	if shouldParse(conn, today, today.AddDate(0, 0, 1)) {
		t.Error("parsed today, interval=1: should not re-parse on same day")
	}
}

func TestShouldParse_IntervalElapsed(t *testing.T) {
	today := truncateToDay(time.Now())
	// A↔A mid interval = 2 days; parsed 2 days ago → should parse
	conn := Connection{
		OriginPopulation: 2_000_000,
		DestPopulation:   2_000_000,
		LastParsedAt:     today.AddDate(0, 0, -2),
		LastUsedAt:       today,
	}
	if !shouldParse(conn, today, today.AddDate(0, 0, 20)) {
		t.Error("parsed 2 days ago, mid interval=2: should parse now")
	}
}

func TestShouldParse_SleepMode(t *testing.T) {
	today := truncateToDay(time.Now())
	// Connection unused for 31 days → sleep mode → interval = 21
	conn := Connection{
		OriginPopulation: 2_000_000, // TierA
		DestPopulation:   2_000_000, // TierA
		LastParsedAt:     today.AddDate(0, 0, -10),
		LastUsedAt:       today.AddDate(0, 0, -31),
	}

	// Normal A↔A near interval = 1. Parsed 10 days ago → would normally parse.
	// But in sleep mode, interval = 21. 10 < 21 → should NOT parse.
	if shouldParse(conn, today, today.AddDate(0, 0, 5)) {
		t.Error("sleep mode: 10 days since last parse < 21 day interval, should not parse")
	}

	// Parsed 21 days ago → sleep interval elapsed → should parse.
	conn.LastParsedAt = today.AddDate(0, 0, -21)
	if !shouldParse(conn, today, today.AddDate(0, 0, 5)) {
		t.Error("sleep mode: 21 days since last parse = interval, should parse")
	}
}

func TestShouldParse_NotSleeping(t *testing.T) {
	today := truncateToDay(time.Now())
	// Last used 29 days ago → NOT sleeping (threshold is > 30)
	conn := Connection{
		OriginPopulation: 2_000_000,
		DestPopulation:   2_000_000,
		LastParsedAt:     today.AddDate(0, 0, -1),
		LastUsedAt:       today.AddDate(0, 0, -29),
	}
	// A↔A near interval = 1; parsed 1 day ago → should parse
	if !shouldParse(conn, today, today.AddDate(0, 0, 5)) {
		t.Error("29 days unused is not sleeping; normal interval should apply")
	}
}

// --- computePriority ---

func TestComputePriority_Ordering(t *testing.T) {
	today := time.Now()
	// A↔A 1 day out should have lower priority number than D↔D 90 days out
	prioHigh := computePriority(Connection{OriginPopulation: 2_000_000, DestPopulation: 2_000_000}, today.AddDate(0, 0, 1))
	prioLow := computePriority(Connection{OriginPopulation: 50_000, DestPopulation: 50_000}, today.AddDate(0, 0, 90))

	if prioHigh >= prioLow {
		t.Errorf("A↔A near (%d) should have lower priority number than D↔D far (%d)", prioHigh, prioLow)
	}
}

func TestComputePriority_TierScoreContribution(t *testing.T) {
	today := time.Now()
	dep := today.AddDate(0, 0, 10) // same date for both

	prioAA := computePriority(Connection{OriginPopulation: 2_000_000, DestPopulation: 2_000_000}, dep)
	prioBB := computePriority(Connection{OriginPopulation: 750_000, DestPopulation: 750_000}, dep)
	prioDD := computePriority(Connection{OriginPopulation: 50_000, DestPopulation: 50_000}, dep)

	if !(prioAA < prioBB && prioBB < prioDD) {
		t.Errorf("priority order should be AA < BB < DD, got %d %d %d", prioAA, prioBB, prioDD)
	}
}
