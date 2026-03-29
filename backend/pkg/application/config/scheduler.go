package config

import "time"

type SchedulerConfig struct {
	TopNCities     int           `env:"TOP_N"              envDefault:"100"`
	DaysAhead      int           `env:"DAYS_AHEAD"         envDefault:"90"`
	BucketSizeMin  int           `env:"BUCKET_SIZE_MIN"    envDefault:"5"`
	BucketSizeMax  int           `env:"BUCKET_SIZE_MAX"    envDefault:"10"`
	BucketPauseMin time.Duration `env:"BUCKET_PAUSE_MIN"  envDefault:"15s"`
	BucketPauseMax time.Duration `env:"BUCKET_PAUSE_MAX"  envDefault:"30s"`
}
