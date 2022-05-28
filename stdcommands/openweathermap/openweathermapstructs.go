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

type owmGeoCodeStruct struct {
	GeoCodingMap geoCodingMap
}

type geoCodingMap []struct {
	Name       string   `json:"name"`
	Lat        float64  `json:"lat"`
	Lon        float64  `json:"lon"`
	Country    string   `json:"country"`
	State      string   `json:"state"`
	LocalNames locNames `json:"local_names"`
}

type locNames struct {
	Af string `json:"af"`
	Am string `json:"am"`
	An string `json:"an"`
	Ar string `json:"ar"`
	Ba string `json:"ba"`
	Be string `json:"be"`
	Bg string `json:"bg"`
	Bn string `json:"bn"`
	Bo string `json:"bo"`
	Br string `json:"br"`
	Bs string `json:"bs"`
	Ca string `json:"ca"`
	Co string `json:"co"`
	Cs string `json:"cs"`
	Cu string `json:"cu"`
	Cv string `json:"cv"`
	De string `json:"de"`
	El string `json:"el"`
	En string `json:"en"`
	Eo string `json:"eo"`
	Es string `json:"es"`
	Et string `json:"et"`
	Eu string `json:"eu"`
	Fa string `json:"fa"`
	Fi string `json:"fi"`
	Fr string `json:"fr"`
	Fy string `json:"fy"`
	Ga string `json:"ga"`
	Gl string `json:"gl"`
	Gn string `json:"gn"`
	Gu string `json:"gu"`
	Gv string `json:"gv"`
	Ha string `json:"ha"`
	He string `json:"he"`
	Hi string `json:"hi"`
	Hr string `json:"hr"`
	Ht string `json:"ht"`
	Hu string `json:"hu"`
	Hy string `json:"hy"`
	Is string `json:"is"`
	It string `json:"it"`
	Ja string `json:"ja"`
	Ka string `json:"ka"`
	Kk string `json:"kk"`
	Km string `json:"km"`
	Kn string `json:"kn"`
	Ko string `json:"ko"`
	Ku string `json:"ku"`
	Kv string `json:"kv"`
	Ky string `json:"ky"`
	La string `json:"la"`
	Lb string `json:"lb"`
	Li string `json:"li"`
	Ln string `json:"ln"`
	Lt string `json:"lt"`
	Lv string `json:"lv"`
	Mi string `json:"mi"`
	Mk string `json:"mk"`
	Ml string `json:"ml"`
	Mn string `json:"mn"`
	Mr string `json:"mr"`
	My string `json:"my"`
	Ne string `json:"ne"`
	Nl string `json:"nl"`
	No string `json:"no"`
	Oc string `json:"oc"`
	Or string `json:"or"`
	Os string `json:"os"`
	Pa string `json:"pa"`
	Pl string `json:"pl"`
	Ps string `json:"ps"`
	Pt string `json:"pt"`
	Ru string `json:"ru"`
	Sc string `json:"sc"`
	Sh string `json:"sh"`
	Sk string `json:"sk"`
	Sl string `json:"sl"`
	So string `json:"so"`
	Sq string `json:"sq"`
	Sr string `json:"sr"`
	Sv string `json:"sv"`
	Ta string `json:"ta"`
	Te string `json:"te"`
	Tg string `json:"tg"`
	Th string `json:"th"`
	Tk string `json:"tk"`
	Tl string `json:"tl"`
	Tt string `json:"tt"`
	Ug string `json:"ug"`
	Uk string `json:"uk"`
	Ur string `json:"ur"`
	Uz string `json:"uz"`
	Vi string `json:"vi"`
	Wa string `json:"wa"`
	Wo string `json:"wo"`
	Yi string `json:"yi"`
	Yo string `json:"yo"`
	Za string `json:"za"`
	Zh string `json:"zh"`
	Zu string `json:"zu"`
}
