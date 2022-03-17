package openweathermap

type openWeatherMap struct {
	Coord      map[string]float64 // keys are lon and lat
	Weather    []weatherStruct    // 0 key to access struct
	Base       string             // internal parameter
	Main       mainStruct
	Visibility int64
	Wind       windStruct
	Clouds     map[string]int64 // has key "all"
	Rain       interface{}
	Snow       interface{}
	Dt         int64 // Time of data calculation
	Sys        sysStruct
	Timezone   int64  // Shift in seconds from UTC*/
	ID         int64  // City ID
	Name       string // City name
	Cod        int64  // internal parameter
}

type coordStruct struct {
	Lon float64
	Lat float64
}

type weatherStruct struct {
	Description string // Weather condition within the group.
	Icon        string
	ID          int64
	Main        string // Group of weather parameters (Rain, Snow, Extreme etc.)
}

type windStruct struct {
	Speed float64
	Deg   int64
	Gust  float64
}

type mainStruct struct {
	Temp      float64
	FeelsLike float64 `json:"feels_like"`
	TempMin   float64 `json:"temp_min"`
	TempMax   float64 `json:"temp_max"`
	Pressure  int64
	Humidity  int64
}

type sysStruct struct {
	Type    int64
	ID      int64
	Country string
	Sunrise int64
	Sunset  int64
	Message float64
}
