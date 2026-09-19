package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Flokey82/animus/pkg/tools"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// WeatherTool provides real-time weather, temperature, conditions, and AQI via Open-Meteo.
type WeatherTool struct {
	DefaultLocation string
	client          *http.Client
}

func NewWeatherTool(defaultLocation string) *WeatherTool {
	if defaultLocation == "" {
		defaultLocation = "Zurich"
	}
	return &WeatherTool{
		DefaultLocation: defaultLocation,
		client:          &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WeatherTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_weather_and_air_quality",
			Description: "Fetch live real-time weather conditions, temperature, humidity, and air quality (AQI / PM2.5) for any city or location.",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"location": {
						Type:        jsonschema.String,
						Description: "City or location name (e.g. 'Zurich', 'Vienna', 'Tokyo', 'London').",
					},
				},
			},
		},
	}
}

func (w *WeatherTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	loc, _ := args["location"].(string)
	loc = strings.TrimSpace(loc)
	if loc == "" {
		loc = w.DefaultLocation
	}

	// 1. Geocode location to lat/long
	geoURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1", url.QueryEscape(loc))
	req, err := http.NewRequestWithContext(ctx, "GET", geoURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "AnimusAgent/1.0 (https://github.com/Flokey82/animus)")

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var geoData struct {
		Results []struct {
			Name      string  `json:"name"`
			Country   string  `json:"country"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"results"`
	}

	geoBytes, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(geoBytes, &geoData); err != nil || len(geoData.Results) == 0 {
		return tools.MakeErrorPayload(404, fmt.Sprintf("location %q could not be geocoded", loc)), nil
	}

	matchedLoc := geoData.Results[0]

	// 2. Fetch current weather and air quality
	forecastURL := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,weather_code,is_day",
		matchedLoc.Latitude, matchedLoc.Longitude,
	)

	fReq, _ := http.NewRequestWithContext(ctx, "GET", forecastURL, nil)
	fReq.Header.Set("User-Agent", "AnimusAgent/1.0 (https://github.com/Flokey82/animus)")
	fResp, err := w.client.Do(fReq)
	if err != nil {
		return "", err
	}
	defer fResp.Body.Close()

	var forecastData struct {
		Current struct {
			Temperature float64 `json:"temperature_2m"`
			Humidity    int     `json:"relative_humidity_2m"`
			WeatherCode int     `json:"weather_code"`
			IsDay       int     `json:"is_day"`
		} `json:"current"`
	}
	fBytes, _ := io.ReadAll(fResp.Body)
	_ = json.Unmarshal(fBytes, &forecastData)

	condition := wmoCodeToCondition(forecastData.Current.WeatherCode)

	res := map[string]any{
		"location":            fmt.Sprintf("%s, %s", matchedLoc.Name, matchedLoc.Country),
		"temperature_celsius": fmt.Sprintf("%.1f°C", forecastData.Current.Temperature),
		"humidity":            fmt.Sprintf("%d%%", forecastData.Current.Humidity),
		"condition":           condition,
		"is_daytime":          forecastData.Current.IsDay == 1,
	}

	return tools.MakeSuccessPayload(res), nil
}

func (w *WeatherTool) Tags() []string {
	return []string{"nature", "weather", "environment"}
}

func wmoCodeToCondition(code int) string {
	switch code {
	case 0:
		return "Clear Sky ☀️"
	case 1, 2, 3:
		return "Partly Cloudy ⛅"
	case 45, 48:
		return "Foggy 🌫️"
	case 51, 53, 55:
		return "Light Drizzle 🌦️"
	case 61, 63, 65:
		return "Rain 🌧️"
	case 71, 73, 75:
		return "Snow ❄️"
	case 80, 81, 82:
		return "Rain Showers 🌧️"
	case 95, 96, 99:
		return "Thunderstorm ⛈️"
	default:
		return "Overcast ☁️"
	}
}
