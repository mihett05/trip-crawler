package scheduler

import "time"

type Config struct {
	TopNCities    int           `env:"SCHEDULER_TOP_N"              envDefault:"100"`
	DaysAhead     int           `env:"SCHEDULER_DAYS_AHEAD"         envDefault:"90"`
	BucketSizeMin int           `env:"SCHEDULER_BUCKET_SIZE_MIN"    envDefault:"5"`
	BucketSizeMax int           `env:"SCHEDULER_BUCKET_SIZE_MAX"    envDefault:"10"`
	BucketPauseMin time.Duration `env:"SCHEDULER_BUCKET_PAUSE_MIN"  envDefault:"15s"`
	BucketPauseMax time.Duration `env:"SCHEDULER_BUCKET_PAUSE_MAX"  envDefault:"30s"`
}
