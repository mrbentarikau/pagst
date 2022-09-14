package howlongtobeat

import "time"

type HowlongToBeat struct {
	Color           string              `json:"color"`
	Title           string              `json:"title"`
	Category        string              `json:"category"`
	Count           int                 `json:"count"`
	PageCurrent     int                 `json:"pageCurrent"`
	PageTotal       int                 `json:"pageTotal"`
	PageSize        int                 `json:"pageSize"`
	Data            []HowlongToBeatData `json:"data"`
	DisplayModifier interface{}         `json:"displayModifier"`
}

type HowlongToBeatData struct {
	Count           int    `json:"count"`
	GameID          int    `json:"game_id"`
	GameName        string `json:"game_name"`
	GameNameDate    int    `json:"game_name_date"`
	GameAlias       string `json:"game_alias"`
	GameType        string `json:"game_type"`
	GameImage       string `json:"game_image"`
	CompLvlCombine  int    `json:"comp_lvl_combine"`
	CompLvlSp       int    `json:"comp_lvl_sp"`
	CompLvlCo       int    `json:"comp_lvl_co"`
	CompLvlMp       int    `json:"comp_lvl_mp"`
	CompLvlSpd      int    `json:"comp_lvl_spd"`
	CompMain        int    `json:"comp_main"`
	CompPlus        int    `json:"comp_plus"`
	Comp100         int    `json:"comp_100"`
	CompAll         int    `json:"comp_all"`
	CompMainCount   int    `json:"comp_main_count"`
	CompPlusCount   int    `json:"comp_plus_count"`
	Comp100Count    int    `json:"comp_100_count"`
	CompAllCount    int    `json:"comp_all_count"`
	InvestedCo      int    `json:"invested_co"`
	InvestedMp      int    `json:"invested_mp"`
	InvestedCoCount int    `json:"invested_co_count"`
	InvestedMpCount int    `json:"invested_mp_count"`
	CountComp       int    `json:"count_comp"`
	CountSpeedrun   int    `json:"count_speedrun"`
	CountBacklog    int    `json:"count_backlog"`
	CountReview     int    `json:"count_review"`
	ReviewScore     int    `json:"review_score"`
	CountPlaying    int    `json:"count_playing"`
	CountRetired    int    `json:"count_retired"`
	ProfileDev      string `json:"profile_dev"`
	ProfilePopular  int    `json:"profile_popular"`
	ProfileSteam    int    `json:"profile_steam"`
	ProfilePlatform string `json:"profile_platform"`
	ReleaseWorld    int    `json:"release_world"`

	CompMainDur           time.Duration
	CompPlusDur           time.Duration
	Comp100Dur            time.Duration
	CompMainHumanize      string
	CompPlusHumanize      string
	Comp100Humanize       string
	GameURL               string
	ImageURL              string
	JaroWinklerSimilarity float64
}

type HowlongToBeatQuery struct {
	SearchType    string   `json:"searchType"`
	SearchTerms   []string `json:"searchTerms"`
	SearchPage    int      `json:"searchPage"`
	Size          int      `json:"size"`
	SearchOptions struct {
		Games struct {
			UserID        int    `json:"userId"`
			Platform      string `json:"platform"`
			SortCategory  string `json:"sortCategory"`
			RangeCategory string `json:"rangeCategory"`
			RangeTime     struct {
				Min int `json:"min"`
				Max int `json:"max"`
			} `json:"rangeTime"`
			Gameplay struct {
				Perspective string `json:"perspective"`
				Flow        string `json:"flow"`
				Genre       string `json:"genre"`
			} `json:"gameplay"`
			Modifier string `json:"modifier"`
		} `json:"games"`
		Users struct {
			SortCategory string `json:"sortCategory"`
		} `json:"users"`
		Filter     string `json:"filter"`
		Sort       int    `json:"sort"`
		Randomizer int    `json:"randomizer"`
	} `json:"searchOptions"`
}
