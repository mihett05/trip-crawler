package scheduler

type Config struct {
	TopNCities          int `env:"SCHEDULER_TOP_N"            envDefault:"100"`
	DaysAhead           int `env:"SCHEDULER_DAYS_AHEAD"       envDefault:"90"`
	PublishRatePerSec   int `env:"SCHEDULER_PUBLISH_RATE"     envDefault:"500"`
	StationLoadHour     int `env:"SCHEDULER_STATION_HOUR"     envDefault:"2"`
	StationLoadMinute   int `env:"SCHEDULER_STATION_MINUTE"   envDefault:"0"`
	TicketEnqueueHour   int `env:"SCHEDULER_TICKET_HOUR"      envDefault:"3"`
	TicketEnqueueMinute int `env:"SCHEDULER_TICKET_MINUTE"    envDefault:"0"`
}
