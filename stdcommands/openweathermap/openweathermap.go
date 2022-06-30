//inspired by ansiweather cli application > https://github.com/fcambus/ansiweather/

package openweathermap

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common/config"
	"github.com/mrbentarikau/pagst/lib/dcmd"
	"github.com/mrbentarikau/pagst/lib/discordgo"
)

var (
	confOpenWeaterMapAPIKey = config.RegisterOption("yagpdb.openweathermap.apikey", "OpenWeatherMap API key", "")
	units                   = "metric"
	openWeatherMapAPIHost   = "https://api.openweathermap.org/data/2.5/"
	geoCodingAPIHost        = "http://api.openweathermap.org/geo/1.0/direct"
	queryParam              = "?q="
)

var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryTool,
	Name:                "OpenWeatherMap",
	Aliases:             []string{"owm", "oweather", "ow"},
	Description:         "Shows the weather using OpenWeatherMap API. \nLocation is set by city name and optional state code, country code \n eg. <prefix>owm Paris,AR,US.\n -zip needs zipcode, countrycode\n-c gives compact output",
	RunInDM:             true,
	RequiredArgs:        0,
	SlashCommandEnabled: true,
	DefaultEnabled:      true,
	Arguments: []*dcmd.ArgDef{
		{Name: "Location", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "c", Help: "Compact output"},
		{Name: "zip", Type: dcmd.String},
	},
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var compactView bool

		if data.Args[0].Value == nil && data.Switches["zip"].Value == nil {
			return "Provide at least a location name or use -zip flag...", nil
		}

		if data.Switches["c"].Value != nil && data.Switches["c"].Value.(bool) {
			compactView = true
		}

		where := data.Args[0].Str()

		if data.Switches["c"].Value != nil && data.Switches["c"].Value.(bool) {
			compactView = true
		}

		if data.Switches["zip"].Value != nil {
			queryParam = "?zip="
			where = data.Switch("zip").Str()
		}

		geoCode, err := geoCodingAPI(where)
		if err != nil {
			return "", err
		}

		weather, err := weatherFromAPI(where, geoCode)
		if err != nil {
			return "", err
		}

		windDirectionInt := int((float64(weather.Wind.Deg)+11.25)/22.5) % 16 //azimuth degrees to 16 wind rose literals
		windDirection := windDirSlice[windDirectionInt]

		if compactView {
			return createCompact(*weather, windDirection), nil
		}

		// this is necessary currently, because owm geocoding returns very specific location name, e.g. Alt-Kölln vs Berlin
		// but coordinate based query gives more accurate results, e.g. first Rome result is not in US but Italy
		wName := weather.Name
		if len(geoCode.GeoCodingMap) > 0 {
			gcName := ""
			for _, v := range geoCode.GeoCodingMap {
				if v.LocalNames.En != "" {
					gcName = v.LocalNames.En
					break
				}
			}
			if !strings.EqualFold(gcName, wName) && gcName != "" {
				wName = gcName + ", " + wName
			}
		}

		embed := &discordgo.MessageEmbed{
			Title:       "Weather report: " + wName,
			Description: "**" + wName + ", " + weather.Sys.Country + "**" + fmt.Sprintf("%s%.2f %s%.2f", "\nlat: ", weather.Coord["lat"], "lon: ", weather.Coord["lon"]),
			Color:       int(rand.Int63n(16777215)),
			URL:         fmt.Sprintf("https://openweathermap.com/city/%d", weather.ID),
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: fmt.Sprintf("https://openweathermap.com/img/wn/%s@2x.png", weather.Weather[0].Icon)},
			Fields: []*discordgo.MessageEmbedField{
				{Name: fmt.Sprintf("%.0f%s%d%s", weather.Main.Temp, "°C (", int(float64(weather.Main.Temp)*1.8+32), "°F)"), Value: strings.Title(weather.Weather[0].Description), Inline: true},
				{Name: "Wind:", Value: fmt.Sprintf("\\%s %.1f %s %s", windDir[windDirection], weather.Wind.Speed, "m/s", windDirection), Inline: true},
				{Name: "Feels like:", Value: fmt.Sprintf("%.0f%s%.0f%s", weather.Main.FeelsLike, "°C (", weather.Main.FeelsLike*1.8+32, "°F)"), Inline: false},
				{Name: "Pressure", Value: fmt.Sprintf("%d%s", weather.Main.Pressure, "hPa"), Inline: true},
				{Name: "Humidity", Value: fmt.Sprintf("%d%s", weather.Main.Humidity, "%"), Inline: true},
				{Name: "Dew point", Value: fmt.Sprintf("%.0f%s%.0f%s", dewPoint(weather.Main.Humidity, weather.Main.Temp), "°C (", weather.Main.Temp*1.8+32, "°F)"), Inline: false},
				{Name: "Sunrise", Value: fmt.Sprintf("<t:%d:R>\n*%s*", weather.Sys.Sunrise, time.Unix(weather.Sys.Sunrise, 0).UTC().Format(time.RFC822)), Inline: true},
				{Name: "Sunset", Value: fmt.Sprintf("<t:%d:R>\n*%s*", weather.Sys.Sunset, time.Unix(weather.Sys.Sunset, 0).UTC().Format(time.RFC822)), Inline: true},
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		return embed, nil
	},
}

func weatherFromAPI(where string, geoCode *owmGeoCodeStruct) (*openWeatherMap, error) {
	weather := openWeatherMap{}
	queryURL := ""

	if len(geoCode.GeoCodingMap) > 0 {
		lat := geoCode.GeoCodingMap[0].Lat
		lon := geoCode.GeoCodingMap[0].Lon
		queryURL = fmt.Sprint(openWeatherMapAPIHost, "weather?lat=", lat, "&lon=", lon, "&units=", units, "&appid=", confOpenWeaterMapAPIKey.GetString())
	} else {
		queryURL = fmt.Sprint(openWeatherMapAPIHost, "weather", queryParam, where, "&units=", units, "&appid=", confOpenWeaterMapAPIKey.GetString())
	}

	body, _ := requestAPI(queryURL)

	queryErr := json.Unmarshal(body, &weather)
	if queryErr != nil {
		return nil, queryErr
	}

	return &weather, nil
}

func geoCodingAPI(where string) (*owmGeoCodeStruct, error) {
	out := owmGeoCodeStruct{}
	coordinates := geoCodingMap{}
	queryURL := fmt.Sprint(geoCodingAPIHost, queryParam, where, "&limit=", 10, "&appid=", confOpenWeaterMapAPIKey.GetString())

	body, _ := requestAPI(queryURL)

	queryErr := json.Unmarshal(body, &coordinates)
	if queryErr != nil {
		return nil, queryErr
	}

	out.GeoCodingMap = coordinates
	return &out, nil

}

func requestAPI(queryURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "PAGSTDB/20.42.6702")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		return nil, commands.NewPublicError("The API key is incorrect.")
	} else if resp.StatusCode == 429 {
		return nil, commands.NewPublicError("The free tariff has made over 60 API calls per minute.")
	} else if resp.StatusCode != 200 {
		return nil, commands.NewPublicError("Cannot fetch data, status code: ", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Values are calculated using the August-Roche-Magnus approximation.
func dewPoint(humidity int64, temperature float64) float64 {
	beta := 17.625
	lambda := 243.04
	blt := (beta * temperature) / (lambda + temperature)
	dewMath := math.Log(float64(humidity)/100) + blt
	dividend := lambda * dewMath
	divisor := beta - dewMath

	return dividend / divisor
}

func createCompact(weather openWeatherMap, windDirection string) string {
	queryData := []string{
		"Weather report: " + weather.Name + ", " + weather.Sys.Country + fmt.Sprintf("%s%.2f %s%.2f", "\nlat:", weather.Coord["lat"], "lon:", weather.Coord["lon"]),
		strings.Title(weather.Weather[0].Description),
		fmt.Sprintf("%.0f%s%.0f%s", weather.Main.Temp, "°C (", weather.Main.Temp*1.8+32, "°F)"),
		fmt.Sprintf("%s %.0f%s%.0f%s", "Feels like:", weather.Main.FeelsLike, "°C (", weather.Main.FeelsLike*1.8+32, "°F)"),
		fmt.Sprintf("%s %s %.1f %s %s", "Wind:", windDir[windDirection], weather.Wind.Speed, "m/s", windDirection),
		fmt.Sprintf("%s %d%s %d %s", "Humidity:", weather.Main.Humidity, "% // Pressure:", weather.Main.Pressure, "hPa"),
		fmt.Sprintf("%s %s %s %s", "Sunrise:", time.Unix(weather.Sys.Sunrise, 0).Format(time.UnixDate), "\nSunset: ", time.Unix(weather.Sys.Sunset, 0).Format(time.UnixDate)),
	}

	//Weather condition data referenced > https://openweathermap.org/weather-conditions
	var icon []string
	switch weather.Weather[0].ID {
	case 200, 210, 221, 230, 231:
		icon = iconThunderyShowers
	case 201, 202, 211, 212, 232:
		icon = iconThunderyHeavyRain
	case 300, 301, 310, 311, 321:
		icon = iconLightShowers
	case 302, 312, 313, 314, 531:
		icon = iconHeavyShowers
	case 500, 501, 520, 521:
		icon = iconLightRain
	case 502, 503, 504, 522:
		icon = iconHeavyRain
	case 511, 611, 615, 616:
		icon = iconLightSleet
	case 600:
		icon = iconLightSnow
	case 601, 602:
		icon = iconHeavySnow
	case 612, 613:
		icon = iconLightSleetShowers
	case 620:
		icon = iconLightSnowShowers
	case 621, 622:
		icon = iconHeavySnowShowers
	case 701, 711, 721, 731, 741, 751, 761, 762, 771, 781:
		icon = iconFog
	case 800:
		icon = iconSunny
	case 801, 802:
		icon = iconPartlyCloudy
	case 803, 804:
		icon = iconCloudy
	default:
		icon = iconUnknown
	}

	// More general way to get the weather icon
	/*switch weather.Weather[0].Main {
	case "Clear":
		icon = iconSunny
	case "Clouds":
		icon = iconCloudy
	case "Drizzle":
		icon = iconLightShowers
	case "Rain":
		icon = iconLightRain
	case "Fog", "Mist", "Haze", "Dust", "Sand", "Smoke", "Ash", "Squall", "Tornado":
		icon = iconFog
	case "Snow":
		icon = iconLightSnow
	case "Thunderstorm":
		icon = iconThunderyHeavyRain
	}*/

	out := "```\n"
	out += queryData[0] + "\n\n"
	for i := 0; i < 5; i++ {
		if i >= len(icon) {
			break
		}
		out += icon[i] + "\t" + queryData[i+1] + "\n"
	}

	out += "\n" + queryData[6] + "\n```" + fmt.Sprintf("*Source*: <https://openweathermap.com/city/%d>", weather.ID)
	return out
}
