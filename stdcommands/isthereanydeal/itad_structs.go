package isthereanydeal

type ItadComplete struct {
	GameInfo ItadGameInfo
	Price    ItadPrice
	PriceLow ItadShopPriceLowest
	Search   ItadSearch
}

type ItadSearch struct {
	Data struct {
		Results []ItadSearchResult `json:"results"`
		Urls    struct {
			Search string `json:"search"`
		} `json:"urls"`
	} `json:"data"`
}

type ItadSearchResult struct {
	ID                    int    `json:"id"`
	Plain                 string `json:"plain"`
	Title                 string `json:"title"`
	JaroWinklerSimilarity float64
}

type ItadPlains struct {
	Data map[string]ItadPlainsStore `json:"data"`
}

type ItadPlainsStore map[string]string

type ItadPrice struct {
	Meta struct {
		Currency string `json:"currency"`
	} `json:".meta"`
	Data map[string]ItadPlainPriceList `json:"data,omitempty"`
}

type ItadPlainPriceList struct {
	List []struct {
		PriceNew float64 `json:"price_new"`
		PriceOld float64 `json:"price_old"`
		PriceCut int     `json:"price_cut"`
		URL      string  `json:"url"`
		Shop     struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"shop"`
		Drm []interface{} `json:"drm"`
	} `json:"list,omitempty"`
	Urls struct {
		Game string `json:"game"`
	} `json:"urls,omitempty"`
}

type ItadShopPriceLowest struct {
	Meta struct {
		Currency string `json:"currency"`
	} `json:".meta"`
	Data    map[string][]ItadPriceLowest  `json:"data,omitempty"`
	Compact map[string]map[string]float64 `json:"-"`
}

type ItadPriceLowest struct {
	Shop  string  `json:"shop"`
	Price float64 `json:"price"`
	//Compact map[string]float64 `json:"-"`
}

type ItadGameInfo struct {
	Data map[string]GameInfo `json:"data"`
}

type GameInfo struct {
	Title        string      `json:"title"`
	Image        string      `json:"image"`
	Greenlight   interface{} `json:"greenlight"`
	IsPackage    bool        `json:"is_package"`
	IsDlc        bool        `json:"is_dlc"`
	Achievements bool        `json:"achievements"`
	TradingCards bool        `json:"trading_cards"`
	EarlyAccess  bool        `json:"early_access"`
	Reviews      struct {
		Steam struct {
			PercPositive int    `json:"perc_positive"`
			Total        int    `json:"total"`
			Text         string `json:"text"`
			Timestamp    int    `json:"timestamp"`
		} `json:"steam"`
	} `json:"reviews"`
	Urls struct {
		Game    string `json:"game"`
		History string `json:"history"`
		Package string `json:"package"`
		Dlc     string `json:"dlc"`
	} `json:"urls"`
}
