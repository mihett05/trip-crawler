package scheduler

import (
	"time"
)

// Tier classifies a city by population size.
// Lower numeric value = larger city.
type Tier int

const (
	TierA Tier = iota // > 1 000 000
	TierB             // 500 000 – 1 000 000
	TierC             // 100 000 – 500 000
	TierD             // < 100 000
)

func tierFromPopulation(pop int) Tier {
	switch {
	case pop > 1_000_000:
		return TierA
	case pop > 500_000:
		return TierB
	case pop > 100_000:
		return TierC
	default:
		return TierD
	}
}

// intervalBracket holds parse intervals (in days) for three date horizons:
// near  = 1–14 days until departure
// mid   = 15–45 days until departure
// far   = 46–90 days until departure
type intervalBracket struct{ near, mid, far int }

type tierPair struct{ a, b Tier }

// normPair ensures the pair is always ordered (smaller Tier value first).
func normPair(t1, t2 Tier) tierPair {
	if t1 > t2 {
		t1, t2 = t2, t1
	}
	return tierPair{t1, t2}
}

// frequencyMatrix defines how often (in days) a route should be re-parsed
// based on city tiers and how far ahead the departure date is.
var frequencyMatrix = map[tierPair]intervalBracket{
	{TierA, TierA}: {1, 2, 4},
	{TierA, TierB}: {1, 3, 7},
	{TierA, TierC}: {2, 5, 14},
	{TierA, TierD}: {3, 7, 21},
	{TierB, TierB}: {1, 4, 10},
	{TierB, TierC}: {2, 7, 14},
	{TierB, TierD}: {3, 10, 21},
	{TierC, TierC}: {3, 10, 21},
	{TierC, TierD}: {5, 14, 30},
	{TierD, TierD}: {7, 21, 60},
}

// parseInterval returns the minimum number of days that must have elapsed
// since a connection was last parsed before it should be enqueued again.
func parseInterval(t1, t2 Tier, daysUntilDeparture int) int {
	b := frequencyMatrix[normPair(t1, t2)]
	switch {
	case daysUntilDeparture <= 14:
		return b.near
	case daysUntilDeparture <= 45:
		return b.mid
	default:
		return b.far
	}
}

const (
	sleepThresholdDays = 30
	sleepIntervalDays  = 21
	maxDaysAhead       = 90
)

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// shouldParse reports whether a parse task should be enqueued today
// for the given connection and departure date.
func shouldParse(conn Connection, today, departureDate time.Time) bool {
	daysUntil := int(departureDate.Sub(today).Hours() / 24)
	if daysUntil < 1 || daysUntil > maxDaysAhead {
		return false
	}

	oTier := tierFromPopulation(conn.OriginPopulation)
	dTier := tierFromPopulation(conn.DestPopulation)

	interval := parseInterval(oTier, dTier, daysUntil)

	// Sleep mode: connection unused by the route builder for too long.
	if today.Sub(conn.LastUsedAt).Hours()/24 > sleepThresholdDays {
		interval = sleepIntervalDays
	}

	daysSinceLastParse := today.Sub(conn.LastParsedAt).Hours() / 24
	return daysSinceLastParse >= float64(interval)
}

// computePriority returns a scheduling priority for a task.
// Lower value = higher priority. A↔A routes with near departure dates rank highest.
func computePriority(conn Connection, departureDate time.Time) int {
	tierScore := int(tierFromPopulation(conn.OriginPopulation)) +
		int(tierFromPopulation(conn.DestPopulation)) // 0 (A+A) … 6 (D+D)
	daysUntil := int(time.Until(departureDate).Hours() / 24)
	return tierScore*100 + daysUntil
}
