package messages

const (
	CitiesStream           = "CITITES"
	CitiesSubjectPrefix    = "CITIES."
	CitiesSubjectRequested = CitiesSubjectPrefix + "requested"
	CitiesSubjectParsed    = CitiesSubjectPrefix + "parsed"

	TripsStream           = "TRIPS"
	TripsSubjectPrefix    = "TRIPS."
	TripsSubjectRequested = TripsSubjectPrefix + "requested"
	TripsSubjectParsed    = TripsSubjectPrefix + "parsed"

	SchedulesTripsSubjectPrefix = "schedules.trips."
	SchedulesTripsSubjectParsed = SchedulesTripsSubjectPrefix + "parsed"
)
