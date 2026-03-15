package wiki

type CityData struct {
	ID             string `json:"Номер"`
	Name           string `json:"Город"`
	Region         string `json:"Субъект РФ"`
	Population2021 string `json:"Население перепись 2021"`
	Population2010 string `json:"Население перепись 2010"`
}
