package myanimelist

import "time"

type AnimeMangaList struct {
	Data    []AnimeMangaListNode `json:"data"`
	Ranking struct {
		Rank int `json:"rank"`
	} `json:"ranking"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type AnimeMangaListNode struct {
	Node AnimeMangaDetails `json:"node"`
	// for user queries
	ListStatus struct {
		Status             string    `json:"status"`
		Score              int       `json:"score"`
		NumWatchedEpisodes int       `json:"num_watched_episodes"`
		IsRewatching       bool      `json:"is_rewatching"`
		IsRereading        bool      `json:"is_rereading"`
		NumVolumesRead     int       `json:"num_volumes_read"`
		NumChaptersRead    int       `json:"num_chapters_read"`
		UpdatedAt          time.Time `json:"updated_at"`
	} `json:"list_status,omitempty"`
}

type AnimeMangaDetails struct {
	ID                    int    `json:"id"`
	Title                 string `json:"title"`
	JaroWinklerSimilarity float64
	MainPicture           struct {
		Medium string `json:"medium"`
		Large  string `json:"large"`
	} `json:"main_picture"`
	AlternativeTitles struct {
		Synonyms []string `json:"synonyms"`
		En       string   `json:"en"`
		Ja       string   `json:"ja"`
	} `json:"alternative_titles"`
	StartDate       string                `json:"start_date"`
	EndDate         string                `json:"end_date"`
	Synopsis        string                `json:"synopsis"`
	Mean            float64               `json:"mean"`
	Rank            int                   `json:"rank"`
	Popularity      int                   `json:"popularity"`
	NumListUsers    int                   `json:"num_list_users"`
	NumScoringUsers int64                 `json:"num_scoring_users"`
	Nsfw            string                `json:"nsfw"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	MediaType       string                `json:"media_type"`
	Status          string                `json:"status"`
	Genres          []AnimatedMangaIDName `json:"genres"`
	MyListStatus    struct {
		Status             string    `json:"status"`
		Score              int       `json:"score"`
		NumEpisodesWatched int       `json:"num_episodes_watched"`
		IsRewatching       bool      `json:"is_rewatching"`
		UpdatedAt          time.Time `json:"updated_at"`
	} `json:"my_list_status"`
	NumEpisodes int `json:"num_episodes"`
	StartSeason struct {
		Year   int    `json:"year"`
		Season string `json:"season"`
	} `json:"start_season"`
	Broadcast struct {
		DayOfTheWeek string `json:"day_of_the_week"`
		StartTime    string `json:"start_time"`
	} `json:"broadcast"`
	Source                 string `json:"source"`
	AverageEpisodeDuration int    `json:"average_episode_duration"`
	Rating                 string `json:"rating"`
	NumVolumes             int    `json:"num_volumes"`
	NumChapters            int    `json:"num_chapters"`
	Authors                []struct {
		Node struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"node"`
		Role string `json:"role"`
	} `json:"authors"`
	Pictures []struct {
		Medium string `json:"medium"`
		Large  string `json:"large"`
	} `json:"pictures"`
	Background      string    `json:"background"`
	RelatedAnime    []Related `json:"related_anime"`
	RelatedManga    []Related `json:"related_manga"`
	Recommendations []struct {
		Node struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			MainPicture struct {
				Medium string `json:"medium"`
				Large  string `json:"large"`
			} `json:"main_picture"`
		} `json:"node"`
		NumRecommendations int `json:"num_recommendations"`
	} `json:"recommendations"`
	Studios    []AnimatedMangaIDName `json:"studios"`
	Statistics struct {
		Status struct {
			Watching    string `json:"watching"`
			Completed   string `json:"completed"`
			OnHold      string `json:"on_hold"`
			Dropped     string `json:"dropped"`
			PlanToWatch string `json:"plan_to_watch"`
		} `json:"status"`
		NumListUsers int `json:"num_list_users"`
	} `json:"statistics"`
	Serialization []struct {
		Node struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"node"`
	} `json:"serialization"`
}

type Related struct {
	Node struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		MainPicture struct {
			Medium string `json:"medium"`
			Large  string `json:"large"`
		} `json:"main_picture"`
	} `json:"node"`
	RelationType          string `json:"relation_type"`
	RelationTypeFormatted string `json:"relation_type_formatted"`
}

type AnimatedMangaIDName struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (a AnimeMangaDetails) Related() []Related {
	if len(a.RelatedAnime) == 0 {
		return a.RelatedManga
	}
	return a.RelatedAnime
}
