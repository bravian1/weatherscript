package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const API_KEY = "AIzaSyB0EdiuE2W_fVSPpfI_soLGymvDoNtA17Y"

func getWeather() string{
	url := "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/manyatta%20kisumu/today?unitGroup=metric&include=events&key=J296FMANPFPLAKB9N8Q8W2YC7&contentType=json"
	client := http.Client{
		Transport: nil,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.URL.Opaque = req.URL.Path
			return nil
		},
		
	}
	response, err := client.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	return string(body)
}
func main() {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(API_KEY))
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		err := client.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()
	model:=client.GenerativeModel("gemini-1.5-flash")
	msg:="Analyze the provided JSON weather data and create a meaningful weather forecast summary for the location and date specified. Focus on the likelihood of rain, temperature range, and provide suggestions on what to wear and when to return home based on the weather conditions. Use natural language and make the forecast easy to understand."+getWeather()
	cs:=model.StartChat()
	resp, err:=cs.SendMessage(ctx, genai.Text(msg))
	if err!= nil {
        log.Fatalln(err)
    }

	for _, cand:=range resp.Candidates{
		if cand.Content !=nil{
			for _, part :=range cand.Content.Parts{
				fmt.Println(part)
            }
		}
	}



}
