package isthereanydeal

import "time"

type ItadComplete struct {
	Price      *ItadPrices
	SearchData *ItadSearchResults
	//GameInfo   *ItadGameInfo
	//PriceLow   *ItadShopPriceLowest
}

type ItadSearchResults []struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Mature bool   `json:"mature"`
}

type ItadPrices []struct {
	ID    string           `json:"id"`
	Deals []ItadPricesDeal `json:"deals"`
}

type ItadPricesDeal struct {
	Shop struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"shop"`
	Price struct {
		Amount    float64 `json:"amount"`
		AmountInt int     `json:"amountInt"`
		Currency  string  `json:"currency"`
	} `json:"price"`
	Regular struct {
		Amount    float64 `json:"amount"`
		AmountInt int     `json:"amountInt"`
		Currency  string  `json:"currency"`
	} `json:"regular"`
	Cut      int `json:"cut"`
	Voucher  any `json:"voucher"`
	StoreLow struct {
		Amount    float64 `json:"amount"`
		AmountInt int     `json:"amountInt"`
		Currency  string  `json:"currency"`
	} `json:"storeLow"`
	HistoryLow struct {
		Amount    float64 `json:"amount"`
		AmountInt int     `json:"amountInt"`
		Currency  string  `json:"currency"`
	} `json:"historyLow"`
	Flag string `json:"flag"`
	Drm  []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"drm"`
	Platforms []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"platforms"`
	Timestamp time.Time `json:"timestamp"`
	Expiry    time.Time `json:"expiry"`
	URL       string    `json:"url"`
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
