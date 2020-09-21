package covidstats

type coronaWorldWideStruct struct {
	Updated                int64
	Country                string
	CountryInfo            countryInfoStruct
	Cases                  int64
	TodayCases             int64
	Deaths                 int64
	TodayDeaths            int64
	Recovered              int64
	TodayRecovered         int64
	Active                 int64
	Critical               int64
	CasesPerOneMillion     int64
	DeathsPerOneMillion    int64
	Tests                  int64
	TestsPerOneMillion     int64
	Population             int64
	Continent              string
	OneCasePerPeople       int64
	OneDeathPerPeople      int64
	OneTestPerPeople       int64
	ActivePerOneMillion    float64
	RecoveredPerOneMillion float64
	CriticalPerOneMillion  float64
}

type countryInfoStruct struct {
	_id  int64 //does not parse this not important
	Iso2 string
	Iso3 string
	Lat  int64
	Long int64
	Flag string
}

type coronaStatesStruct struct {
	State               string
	Updated             int64
	Cases               int64
	TodayCases          int64
	Deaths              int64
	TodayDeaths         int64
	Recovered           int64
	Active              int64
	CasesPerOneMillion  int64
	DeathsPerOneMillion int64
	Tests               int64
	TestsPerOneMillion  int64
	Population          int64
}
