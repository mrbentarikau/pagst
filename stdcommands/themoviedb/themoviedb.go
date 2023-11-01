package themoviedb

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/bot/paginatedmessages"
	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/config"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/stdcommands/util"

	tmdb "github.com/cyruzin/golang-tmdb"
)

var confTMDBAPIKey = config.RegisterOption("yagpdb.tmdb_api_key", "The Movie DB API key", "")
var imageBaseURL = "https://image.tmdb.org/t/p/original/"
var logger = common.GetFixedPrefixLogger("the_movie_db_cmd")
var timeLayout = "2006-01-02"
var tmdbAPI *tmdb.Client
var tmdbBaseURL = "https://www.themoviedb.org/"

func ShouldRegister() bool {
	return confTMDBAPIKey.GetString() != ""
}

var Command = &commands.YAGCommand{
	CmdCategory:               commands.CategoryFun,
	Name:                      "TheMovieDB",
	Aliases:                   []string{"film", "movie", "tmdb", "tv"},
	Description:               "Info about movies, TV-shows and persons from The Movie DB (tmdb) API.\nWithout switches makes a general query covering all subjects.",
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	Arguments: []*dcmd.ArgDef{
		{Name: "Query", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "p", Help: "Paginated output"},
		{Name: "movie", Help: "Query for movies"},
		{Name: "person", Help: "Query for persons"},
		{Name: "tv", Help: "Query for TV"},
		{Name: "nsfw", Help: "Include adult to query"},
		{Name: "year", Help: "Year filter for Movies & TV", Type: dcmd.String},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var embed *discordgo.MessageEmbed
		var err error
		var movie, paginatedView, person, tv bool
		var tmdbSearchResultSlice tmdbSearchResults
		var totalResults int64
		var year string

		include_adult := "false"

		tmdbAPI, err = tmdb.Init(confTMDBAPIKey.GetString())
		if err != nil {
			return fmt.Sprintf("API Error %s", err), err
		}

		if data.Switches["person"].Value != nil && data.Switches["person"].Value.(bool) {
			movie = false
			person = true
			tv = false
		}

		if data.Switches["tv"].Value != nil && data.Switches["tv"].Value.(bool) {
			movie = false
			person = false
			tv = true
			if data.Switches["year"].Str() != "" {
				year = data.Switch("year").Str()
			}
		}

		if data.Switches["movie"].Value != nil && data.Switches["movie"].Value.(bool) {
			movie = true
			person = false
			tv = false
			if data.Switches["year"].Str() != "" {
				year = data.Switch("year").Str()
			}
		}

		if data.Switches["nsfw"].Value != nil && data.Switches["nsfw"].Value.(bool) {
			include_adult = "true"
		}

		queryAPIString := data.Args[0].Str()

		if movie {
			queryResults, err := tmdbAPI.GetSearchMovies(url.QueryEscape(queryAPIString), map[string]string{"include_adult": include_adult, "lang": "en-US", "page": "1", "primary_release_year": year})
			if err != nil {
				return fmt.Sprintf("API Error %s", err), err
			}

			tmdbJSON, _ := json.Marshal(queryResults.SearchMoviesResults)
			err = json.Unmarshal(tmdbJSON, &tmdbSearchResultSlice)
			if err != nil {
				return "error unmarshaling tmdb query to different type", err
			}
			for i := range tmdbSearchResultSlice.Results {
				tmdbSearchResultSlice.Results[i].MediaType = "movie"
			}

			totalResults = queryResults.TotalResults

		} else if tv {
			queryResults, err := tmdbAPI.GetSearchTVShow(url.QueryEscape(queryAPIString), map[string]string{"include_adult": include_adult, "lang": "en-US", "page": "1", "first_air_date_year": year})
			if err != nil {
				return fmt.Sprintf("API Error %s", err), err
			}

			tmdbJSON, _ := json.Marshal(queryResults.SearchTVShowsResults)
			err = json.Unmarshal(tmdbJSON, &tmdbSearchResultSlice)
			if err != nil {
				return "error unmarshaling tmdb query to different type", err
			}
			for i := range tmdbSearchResultSlice.Results {
				tmdbSearchResultSlice.Results[i].MediaType = "tv"
			}

			totalResults = queryResults.TotalResults

		} else if person {
			queryResults, err := tmdbAPI.GetSearchPeople(url.QueryEscape(queryAPIString), map[string]string{"include_adult": include_adult, "lang": "en-US", "page": "1"})
			if err != nil {
				return fmt.Sprintf("API Error %s", err), err
			}

			tmdbJSON, _ := json.Marshal(queryResults.SearchPeopleResults)
			err = json.Unmarshal(tmdbJSON, &tmdbSearchResultSlice)
			if err != nil {
				return "error unmarshaling tmdb query to different type", err
			}
			for i := range tmdbSearchResultSlice.Results {
				tmdbSearchResultSlice.Results[i].MediaType = "person"
			}

			totalResults = queryResults.TotalResults

		} else {
			queryResults, err := tmdbAPI.GetSearchMulti(url.QueryEscape(queryAPIString), map[string]string{"include_adult": "true", "lang": "en-US", "page": "1"})
			if err != nil {
				return fmt.Sprintf("API Error %s", err), err
			}

			tmdbJSON, _ := json.Marshal(queryResults.SearchMultiResults)
			err = json.Unmarshal(tmdbJSON, &tmdbSearchResultSlice)
			if err != nil {
				return "error unmarshaling tmdb query to different type", err
			}

			totalResults = queryResults.TotalResults
		}

		if data.Switches["p"].Value != nil && data.Switches["p"].Value.(bool) {
			paginatedView = true
		}

		var footerExtra = "themoviedb.org"
		var pm *paginatedmessages.PaginatedMessage
		if totalResults > 0 {
			if paginatedView {
				pm, err = paginatedmessages.CreatePaginatedMessage(
					data.GuildData.GS.ID, data.ChannelID, 1, len(tmdbSearchResultSlice.Results), func(p *paginatedmessages.PaginatedMessage, page int) (interface{}, error) {
						i := page - 1
						paginatedEmbed := embedCreator(tmdbSearchResultSlice, i, paginatedView)
						p.FooterExtra = []string{footerExtra}

						return paginatedEmbed, nil
					}, footerExtra)
				if err != nil {
					return "Something went wrong making the paginated messages", nil
				}
			} else {
				embed = embedCreator(tmdbSearchResultSlice, 0, false)
				return embed, nil
			}
		} else {
			return "No match found for query", nil
		}

		return pm, nil

	},
}

func embedCreator(tmdbData tmdbSearchResults, i int, paginated bool) *discordgo.MessageEmbed {
	var date, description, imageURL, title, movieURL string
	var embedFields []*discordgo.MessageEmbedField

	switch t := tmdbData.Results[i]; t.MediaType {
	case "movie":
		if date = t.ReleaseDate; len(date) >= 4 {
			date = fmt.Sprintf(" (%s)", t.ReleaseDate[0:4])
		}

		title = t.Title
		if title == "" {
			title = t.OriginalTitle
		}
		title += date

		movieURL = fmt.Sprintf("%s%s/%d", tmdbBaseURL, t.MediaType, t.ID)
		imageURL = imageBaseURL + t.PosterPath

		if t.VoteAverage != 0 {
			voteAverage := &discordgo.MessageEmbedField{Name: "User Score", Value: fmt.Sprintf("**%.0f%%** (by %s users)", t.VoteAverage*10, util.HumanizeThousands(int64(t.VoteCount))), Inline: true}
			embedFields = append(embedFields, voteAverage)
		}

		movie, err := tmdbAPI.GetMovieDetails(int(t.ID), map[string]string{"append_to_response": "release_dates"})
		if err == nil {
			if movie.Tagline != "" {
				description += fmt.Sprintf("*%s*\n", movie.Tagline)
			}

			if len(movie.Genres) > 0 {
				tmdbJSON, err := json.Marshal(movie.Genres)
				if err != nil {
					logger.WithError(err).Error("movie genres failed to marshal")
				}
				genres := &discordgo.MessageEmbedField{Name: "Genres", Value: genreSliceJoin(tmdbJSON, 4), Inline: true}
				embedFields = append(embedFields, genres)
			}

			if movie.Runtime != 0 {
				duration := time.Minute * time.Duration(movie.Runtime)
				runtime := &discordgo.MessageEmbedField{Name: "Runtime", Value: common.HumanizeDurationShort(common.DurationPrecisionMinutes, duration), Inline: true}
				embedFields = append(embedFields, runtime)
			}

			if movie.Budget != 0 {
				budget := &discordgo.MessageEmbedField{Name: "Budget", Value: fmt.Sprintf("$%s", util.HumanizeThousands(int64(movie.Budget))), Inline: true}
				embedFields = append(embedFields, budget)
			}

			if movie.Revenue != 0 {
				revenue := &discordgo.MessageEmbedField{Name: "Revenue", Value: fmt.Sprintf("$%s", util.HumanizeThousands(int64(movie.Revenue))), Inline: true}
				embedFields = append(embedFields, revenue)
			}

			if len(movie.ProductionCompanies) > 0 {
				tmdbJSON, err := json.Marshal(movie.ProductionCompanies)
				if err != nil {
					logger.WithError(err).Error("movie production companies failed to marshal")
				}
				prodCompanies := &discordgo.MessageEmbedField{Name: "Production Companies", Value: genreSliceJoin(tmdbJSON, 2), Inline: true}
				embedFields = append(embedFields, prodCompanies)
			}

			if movie.Status != "" {
				var cCode, rating, releaseDate string
				for _, r := range movie.MovieReleaseDatesAppend.ReleaseDates.Results {
					for _, rDates := range r.ReleaseDates {
						if r.Iso3166_1 == "US" {
							if rDates.Certification != "" {
								rating = rDates.Certification + " (US)"
							}
						}
						if len(rDates.ReleaseDate) >= 10 && rDates.ReleaseDate[0:10] == movie.ReleaseDate {
							releaseDate = rDates.ReleaseDate[0:10]
							cCode = r.Iso3166_1
						}
					}
				}

				parsedTime, _ := time.Parse(timeLayout, releaseDate)
				status := &discordgo.MessageEmbedField{Name: "Status", Value: fmt.Sprintf("%s\n%s\n%s", movie.Status, rating, parsedTime.Format(fmt.Sprintf("02 Jan 2006 (%s)", cCode))), Inline: true}
				embedFields = append(embedFields, status)
			}

			var moviePages string
			if movie.Homepage != "" {
				moviePages += fmt.Sprintf("\n[Homepage](%s)", movie.Homepage)
			}

			if movie.IMDbID != "" {
				moviePages += fmt.Sprintf("\n[IMDB](https://www.imdb.com/title/%s/)", movie.IMDbID)
			}

			if moviePages != "" {
				externalLinks := &discordgo.MessageEmbedField{Name: "External Links", Value: moviePages, Inline: true}
				embedFields = append(embedFields, externalLinks)
			}
		}
		description += t.Overview

	case "tv":
		date = t.FirstAirDate
		if len(date) >= 4 {
			date = fmt.Sprintf(" (%s", t.FirstAirDate[0:4])
		}

		if t.VoteAverage != 0 {
			voteAverage := &discordgo.MessageEmbedField{Name: "User Score", Value: fmt.Sprintf("**%.0f%%** (by %s users)", t.VoteAverage*10, util.HumanizeThousands(int64(t.VoteCount))), Inline: true}
			embedFields = append(embedFields, voteAverage)
		}

		title = t.Name
		if title == "" {
			title = t.OriginalName
		}

		movieURL = fmt.Sprintf("%s%s/%d", tmdbBaseURL, t.MediaType, t.ID)
		imageURL = imageBaseURL + t.PosterPath

		tv, err := tmdbAPI.GetTVDetails(int(t.ID), map[string]string{"append_to_response": "content_ratings"})
		if err == nil {
			if tv.LastAirDate != "" && len(tv.LastAirDate) >= 4 {
				date += fmt.Sprintf(" - %s)", tv.LastAirDate[0:4])
			} else {
				date += ")"
			}

			var rating string
			for _, r := range tv.TVContentRatingsAppend.ContentRatings.Results {
				if r.Iso3166_1 == "US" {
					if r.Rating != "" {
						rating = r.Rating + " (US)"
					}
				}

			}

			if rating != "" {
				description += rating + "\n\n"
			}

			if tv.Tagline != "" {
				description += fmt.Sprintf("*%s*\n", tv.Tagline)
			}

			if len(tv.Genres) > 0 {
				tmdbJSON, err := json.Marshal(tv.Genres)
				if err != nil {
					logger.WithError(err).Error("tv genres failed to marshal")
				}
				genres := &discordgo.MessageEmbedField{Name: "Genres", Value: genreSliceJoin(tmdbJSON, 4), Inline: true}
				embedFields = append(embedFields, genres)
			}

			if len(tv.CreatedBy) > 0 {
				tmdbJSON, err := json.Marshal(tv.CreatedBy)
				if err != nil {
					logger.WithError(err).Error("tv created by failed to marshal")
				}
				createdBy := &discordgo.MessageEmbedField{Name: "Created by", Value: genreSliceJoin(tmdbJSON, 4), Inline: true}
				embedFields = append(embedFields, createdBy)
			}

			if len(tv.Networks) > 0 {
				tmdbJSON, err := json.Marshal(tv.Networks)
				if err != nil {
					logger.WithError(err).Error("tv networks failed to marshal")
				}
				prodCompanies := &discordgo.MessageEmbedField{Name: "Networks", Value: genreSliceJoin(tmdbJSON, 3), Inline: true}
				embedFields = append(embedFields, prodCompanies)
			}

			if len(tv.ProductionCompanies) > 0 {
				tmdbJSON, err := json.Marshal(tv.ProductionCompanies)
				if err != nil {
					logger.WithError(err).Error("tv production companies failed to marshal")
				}
				prodCompanies := &discordgo.MessageEmbedField{Name: "Production Companies", Value: genreSliceJoin(tmdbJSON, 2), Inline: true}
				embedFields = append(embedFields, prodCompanies)
			}

			if len(tv.EpisodeRunTime) > 0 {
				var averageRunTime int
				for _, runTime := range tv.EpisodeRunTime {
					averageRunTime += runTime
				}
				averageRunTime = averageRunTime / len(tv.EpisodeRunTime)

				duration := time.Minute * time.Duration(averageRunTime)
				runtime := &discordgo.MessageEmbedField{Name: "Avg. Runtime", Value: common.HumanizeDurationShort(common.DurationPrecisionMinutes, duration), Inline: true}
				embedFields = append(embedFields, runtime)
			}

			if tv.NumberOfEpisodes > 0 {
				numEpisodes := &discordgo.MessageEmbedField{Name: "Episodes", Value: fmt.Sprintf("%d", tv.NumberOfEpisodes), Inline: true}
				embedFields = append(embedFields, numEpisodes)
			}

			if tv.NumberOfSeasons > 0 {
				numSeasons := &discordgo.MessageEmbedField{Name: "Seasons", Value: fmt.Sprintf("%d", tv.NumberOfSeasons), Inline: true}
				embedFields = append(embedFields, numSeasons)
			}

			if tv.Status != "" {
				status := &discordgo.MessageEmbedField{Name: "Status", Value: tv.Status, Inline: true}
				embedFields = append(embedFields, status)
			}

			if lastEpisodeToAir := tv.LastEpisodeToAir; lastEpisodeToAir.AirDate != "" {
				parsedTime, _ := time.Parse(timeLayout, lastEpisodeToAir.AirDate)
				lastEpisodeDetails := fmt.Sprintf("S%02dE%02d - %s\n%s", lastEpisodeToAir.SeasonNumber, lastEpisodeToAir.EpisodeNumber, lastEpisodeToAir.Name, parsedTime.Format("02 Jan 2006"))
				lastEpisode := &discordgo.MessageEmbedField{Name: "Last Episode", Value: lastEpisodeDetails, Inline: true}
				embedFields = append(embedFields, lastEpisode)
			}

			if nextEpisodeToAir := tv.NextEpisodeToAir; nextEpisodeToAir.AirDate != "" {
				parsedTime, _ := time.Parse(timeLayout, nextEpisodeToAir.AirDate)
				nextEpisodeDetails := fmt.Sprintf("S%02dE%02d - %s\n%s", nextEpisodeToAir.SeasonNumber, nextEpisodeToAir.EpisodeNumber, nextEpisodeToAir.Name, parsedTime.Format("02 Jan 2006"))
				nextEpisode := &discordgo.MessageEmbedField{Name: "Next Episode", Value: nextEpisodeDetails, Inline: true}
				embedFields = append(embedFields, nextEpisode)
			}

			var tvPages string
			if tv.Homepage != "" {
				tvPages += fmt.Sprintf("\n[Homepage](%s)", tv.Homepage)
			}

			if tvPages != "" {
				externalLinks := &discordgo.MessageEmbedField{Name: "External Links", Value: tvPages, Inline: false}
				embedFields = append(embedFields, externalLinks)
			}
		} else {
			if date != "" {
				date += ")"
			}
		}
		title += date
		description += t.Overview

	case "person":
		movieURL = fmt.Sprintf("%s%s/%d", tmdbBaseURL, t.MediaType, t.ID)
		imageURL = imageBaseURL + t.ProfilePath
		title = "unknown, API issues"
		description = "unknown, API issues"

		person, err := tmdbAPI.GetPersonDetails(int(t.ID), map[string]string{})
		if err == nil {
			title = person.Name
			description = common.CutStringShort(person.Biography, 2048)

			if person.Birthday != "" {
				parsedTime, _ := time.Parse(timeLayout, person.Birthday)
				birthDateField := &discordgo.MessageEmbedField{Name: "Birthdate", Value: parsedTime.Format("02 Jan 2006"), Inline: true}
				embedFields = append(embedFields, birthDateField)
			}

			if person.Deathday != "" {
				parsedTime, _ := time.Parse(timeLayout, person.Deathday)
				dayOfDeathField := &discordgo.MessageEmbedField{Name: "Day of Death", Value: parsedTime.Format("02 Jan 2006"), Inline: true}
				embedFields = append(embedFields, dayOfDeathField)
			}

			if person.PlaceOfBirth != "" {
				placeOfBirth := &discordgo.MessageEmbedField{Name: "Place of Birth", Value: person.PlaceOfBirth, Inline: true}
				embedFields = append(embedFields, placeOfBirth)
			}

			var personPages string
			if person.Homepage != "" {
				personPages += fmt.Sprintf("\n[Homepage](%s)", person.Homepage)
			}

			if person.IMDbID != "" {
				personPages += fmt.Sprintf("\n[IMDB](https://www.imdb.com/name/%s/)", person.IMDbID)
			}

			if personPages != "" {
				externalLinks := &discordgo.MessageEmbedField{Name: "External Links", Value: personPages, Inline: true}
				embedFields = append(embedFields, externalLinks)
			}
		}
	default:
		description = "Nothing found, API wonky..."
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		URL:         movieURL,
		Description: description,
		Color:       int(rand.Int63n(0xffffff)),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: imageURL,
		},
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "The Movie DB",
			URL:     tmdbBaseURL,
			IconURL: "https://www.themoviedb.org/assets/2/apple-touch-icon-57ed4b3b0450fd5e9a0c20f34e814b82adaa1085c79bdde2f00ca8787b63d2c4.png",
		},
		Fields: embedFields,
	}

	if !paginated {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: "themoviedb.org"}
		embed.Timestamp = time.Now().Format(time.RFC3339)
	}

	return embed
}

func genreSliceJoin(tmdbSlice []byte, cap int) string {
	var builder strings.Builder
	var tmdbCCGSlice tmdbCreatorsCompaniesGenres
	err := json.Unmarshal(tmdbSlice, &tmdbCCGSlice)
	if err != nil {
		logger.WithError(err).Error("genreSliceJoin failed to unmarshal")
	}

	for i, j := range tmdbCCGSlice {
		if builder.Len() != 0 {
			builder.WriteString(", ")
		}

		builder.WriteString(j.Name)
		if j.OriginCountry != "" {
			builder.WriteString(fmt.Sprintf(" (%s)", j.OriginCountry))
		}
		if i >= cap {
			builder.WriteString("...")
			break
		}
	}
	return builder.String()
}
