package rzd

type SuggestResponse struct {
	City  []Node `json:"city"`
	Train []Node `json:"train"`
}

type Node struct {
	NodeID        string `json:"nodeId"`
	CityID        string `json:"cityId"`
	ExpressCode   string `json:"expressCode"`
	Name          string `json:"name"`
	NodeType      string `json:"nodeType"`
	TransportType string `json:"transportType"`
	Region        string `json:"region"`
}

type TrainResponse struct {
	OriginCode      string  `json:"OriginCode"`
	DestinationCode string  `json:"DestinationCode"`
	Trains          []Train `json:"Trains"`
}

type Train struct {
	TrainNumber       string     `json:"TrainNumber"`
	TrainName         string     `json:"TrainName"`
	OriginName        string     `json:"OriginName"`
	DestinationName   string     `json:"DestinationName"`
	DepartureDateTime string     `json:"DepartureDateTime"`
	ArrivalDateTime   string     `json:"ArrivalDateTime"`
	TripDistance      int        `json:"TripDistance"`
	TripDuration      float64    `json:"TripDuration"`
	IsBranded         bool       `json:"IsBranded"`
	CarGroups         []CarGroup `json:"CarGroups"`
}

type CarGroup struct {
	CarTypeName        string   `json:"CarTypeName"`
	ServiceClasses     []string `json:"ServiceClasses"`
	MinPrice           *float64 `json:"MinPrice"`
	MaxPrice           *float64 `json:"MaxPrice"`
	TotalPlaceQuantity int      `json:"TotalPlaceQuantity"`
	LowerPlaceQuantity int      `json:"LowerPlaceQuantity"`
	UpperPlaceQuantity int      `json:"UpperPlaceQuantity"`
}
