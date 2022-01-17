//inspired by ansiweather cli application > https://github.com/fcambus/ansiweather/

package openweathermap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/commands"
	"github.com/mrbentarikau/pagst/common/config"
	"github.com/mrbentarikau/pagst/lib/dcmd"
)

var (
	confOpenWeaterMapAPIKey = config.RegisterOption("yagpdb.openweathermap.apikey", "OpenWeatherMap API key", "")
	units                   = "metric"
	openWeatherMapAPIHost   = "https://api.openweathermap.org/data/2.5/"
)

var Command = &commands.YAGCommand{
	CmdCategory:         commands.CategoryTool,
	Name:                "OpenWeatherMap",
	Aliases:             []string{"owm", "oweather", "ow"},
	Description:         "Shows the weather using OpenWeatherMap API. \nLocation is set by city name and optional state code, country code \n eg. <prefix>owm Paris,AR,US.\n -zip needs zipcode, countrycode",
	RunInDM:             true,
	RequiredArgs:        0,
	SlashCommandEnabled: true,
	DefaultEnabled:      true,
	Arguments: []*dcmd.ArgDef{
		{Name: "Location", Type: dcmd.String},
	},
	ArgSwitches: []*dcmd.ArgDef{
		{Name: "zip", Type: dcmd.String},
	},
	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		var queryData []string
		weather := openWeatherMap{}
		queryParam := "?q="

		if data.Args[0].Value == nil && data.Switches["zip"].Value == nil {
			return "Provide at least a location name or use -zip flag...", nil
		}

		where := data.Args[0].Str()

		if data.Switches["zip"].Value != nil {
			queryParam = "?zip="
			where = data.Switch("zip").Str()
		}

		queryURL := fmt.Sprintf(openWeatherMapAPIHost + "weather" + queryParam + where + "&units=" + units + "&appid=" + confOpenWeaterMapAPIKey.GetString())

		req, err := http.NewRequest("GET", queryURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "curlPAGST/7.65.1")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == 401 {
			return "Your API key is incorrect.", nil
		} else if resp.StatusCode == 429 {
			return "Your free tariff has made over 60 API calls per minute.", nil
		} else if resp.StatusCode != 200 {
			return "Cannot fetch weather data for the given location: " + where, nil
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		queryErr := json.Unmarshal(body, &weather)
		if queryErr != nil {
			return nil, queryErr
		}

		windDirectionInt := int((float64(weather.Wind.Deg)+11.25)/22.5) % 16 //azimuth degrees to 16 wind rose literals
		windDirection := windDirSlice[windDirectionInt]
		queryData = []string{
			"Weather report: " + weather.Name + ", " + weather.Sys.Country + fmt.Sprintf("%s%.2f %s%.2f", "\nlat:", weather.Coord["lat"], "lon:", weather.Coord["lon"]),
			strings.Title(weather.Weather[0].Description),
			fmt.Sprintf("%.0f%s%d%s", weather.Main.Temp, "°C (", int(float64(weather.Main.Temp)*1.8+32), "°F)"),
			fmt.Sprintf("%s %.0f%s%d%s", "Feels like:", weather.Main.Feels_Like, "°C (", int(float64(weather.Main.Feels_Like)*1.8+32), "°F)"),
			//fmt.Sprintf("%s %.0f%s%.0f %s%d%s%d %s", "Min..Max:", weather.Main.Temp_Min, "..", weather.Main.Temp_Max, "°C (", int(float64(weather.Main.Temp_Min)*1.8+32), "..", int(float64(weather.Main.Temp_Max)*1.8+32), "°F)"),
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

		out += "\n" + queryData[6] + "\n```"
		return out, nil
	},
}
