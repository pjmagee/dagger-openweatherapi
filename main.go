package main

import (
	"context"
	"dagger/dagger-openweatherapi/internal/dagger"
	"encoding/json"
	"fmt"
	owm "github.com/briandowns/openweathermap"
	"strconv"
)

type OpenWeatherApi struct {
	// The apiKey to use for the OpenWeatherMap API
	ApiKey *dagger.Secret
	// The unit to use (C, F, or K)
	Unit string
	// The 2-letter ISO code for the language to use (en, es, etc.)
	Lang string
	// The weather data returned by the API call
	Result Weather
}

type Weather struct {
	// The temperature in the requested unit
	Temp string `json:"temp"`
	// The unit of the temperature (C, F, or K)
	Unit string `json:"unit"`
	// The description of the weather (e.g. "clear sky", "light rain", etc.)
	Description string `json:"description"`
	// The "feels like" temperature in the requested unit
	FeelsLike string `json:"feels_like"`
	// The summary of the weather (e.g. "London, clear sky, 20°C (🌤️)")
	Summary string `json:"summary"`
	// The icon to use for the weather (e.g. "🌤️")
	Icon string `json:"icon"`
}

// creates a new instance of the OpenWeatherApi
func New(
// The apiKey to use for the OpenWeatherMap API
	apiKey *dagger.Secret,
// The unit to use (C, F, or K)
// +default="C"
	unit string,
// The 2-letter ISO code for the language to use (en, es, etc.)
// +default="en"
	lang string) *OpenWeatherApi {
	return &OpenWeatherApi{
		ApiKey: apiKey,
		Unit:   unit,
		Lang:   lang,
	}
}

// returns the weather data as a JSON string
func (m *OpenWeatherApi) AsJson() (dagger.JSON, error) {

	data, err := json.Marshal(m.Result)
	if err != nil {
		return "", err
	}

	return dagger.JSON(data), nil
}

// retrieves the current weather for the given latitude and longitude
func (m *OpenWeatherApi) UseCoordinates(
// The latitude
	lat string,
// The longitude
	lon string) (*OpenWeatherApi, error) {

	w, err := m.newCurrent()
	if err != nil {
		return nil, err
	}

	latf, err := strconv.ParseFloat(lat, 32)
	if err != nil {
		return nil, err
	}

	lonf, err := strconv.ParseFloat(lon, 32)
	if err != nil {
		return nil, err
	}

	err = w.CurrentByCoordinates(&owm.Coordinates{
		Latitude:  latf,
		Longitude: lonf,
	})

	if err != nil {
		return nil, err
	}

	weather, _ := m.getWeather(w)
	m.Result = weather
	return m, nil
}

// retrieves the current weather for the given location name
func (m *OpenWeatherApi) UseLocation(
// The name of the location (e,g "London,UK", "New York,US", "Tokyo,JP", "Sydney,AU")
	name string) (*OpenWeatherApi, error) {

	w, err := m.newCurrent()
	if err != nil {
		return nil, err
	}

	err = w.CurrentByName(name)
	if err != nil {
		return nil, err
	}

	weather, _ := m.getWeather(w)
	m.Result = weather
	return m, nil
}

func (m *OpenWeatherApi) newCurrent() (*owm.CurrentWeatherData, error) {
	apiKey, err := m.ApiKey.Plaintext(context.Background())
	if err != nil {
		return nil, err
	}
	return owm.NewCurrent(m.Unit, m.Lang, apiKey)
}

func (m *OpenWeatherApi) getWeather(current *owm.CurrentWeatherData) (Weather, error) {

	unit := displayUnit(current.Unit)
	icon := m.getIcon(current)
	desc := current.Weather[0].Description

	weather := Weather{
		Temp:        fmt.Sprintf("%g", current.Main.Temp),
		Unit:        unit,
		Description: desc,
		FeelsLike:   fmt.Sprintf("%g", current.Main.FeelsLike),
		Icon:        icon,
	}

	fullTemp := fmt.Sprintf("%s%s", weather.Temp, unit)
	feelsLike := fmt.Sprintf("%s%s", weather.FeelsLike, unit)

	weather.Summary = fmt.Sprintf("%s, %s, %s (%s) %s",
		current.Name,
		desc,
		fullTemp,
		feelsLike,
		icon,
	)

	return weather, nil
}

func (m *OpenWeatherApi) getIcon(current *owm.CurrentWeatherData) string {

	icon := current.Weather[0].Icon

	switch icon {
	case "01d":
		return "☀️"
	case "01n":
		return "🌙"
	case "02d":
		return "🌤️"
	case "02n":
		return "🌙️"
	case "03d":
		return "🌥️"
	case "03n":
		return "☁️"
	case "04d", "04n":
		return "☁️"
	case "09d", "09n", "10d", "10n":
		return "🌧️"
	case "11d", "11n":
		return "🌩️"
	case "13d":
		return "❄️"
	case "13n":
		return "❄️"
	case "50d", "50n":
		return "🌫️"
	default:
		return "🤷"
	}
}

func displayUnit(unit string) string {
	switch unit {
	case "metric", "C":
		return "°C"
	case "imperial", "F":
		return "°F"
	case "kelvin", "K":
		return "K"
	default:
		return unit
	}
}
