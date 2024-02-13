package isthereanydeal

import "time"

type ItadComplete struct {
	Price      *ItadPrices
	SearchData *ItadSearchResults
	GameInfo   *ItadGameInfo
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
	ID           string   `json:"id"`
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Type         string   `json:"type"`
	Mature       bool     `json:"mature"`
	EarlyAccess  bool     `json:"earlyAccess"`
	Achievements bool     `json:"achievements"`
	TradingCards bool     `json:"tradingCards"`
	Appid        int      `json:"appid"`
	Tags         []string `json:"tags"`
	ReleaseDate  string   `json:"releaseDate"`
	Developers   []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"developers"`
	Publishers []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"publishers"`
	Reviews []struct {
		Score  int    `json:"score"`
		Source string `json:"source"`
		Count  int    `json:"count"`
		URL    string `json:"url"`
	} `json:"reviews"`
	Stats struct {
		Rank       int `json:"rank"`
		Waitlisted int `json:"waitlisted"`
		Collected  int `json:"collected"`
	} `json:"stats"`
	Players struct {
		Recent int `json:"recent"`
		Day    int `json:"day"`
		Week   int `json:"week"`
		Peak   int `json:"peak"`
	} `json:"players"`
}
