package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type User struct {
	Name  string
	Email string
}

// the getweather function is used to get the weather data from the visual crossing api which is free
// we first create a client to make the request to the api then we get the response from the api
// we use io.ReadAll to read the body of the response which should be a json and return it as a string
func getWeather() string {
	url := "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/kisumu/tomorrow?unitGroup=metric&include=events&key=" + os.Getenv("VISUAL_KEY") + "&contentType=json"
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
	
	users := []User{
		{Name: "Bravian", Email: os.Getenv("EMAIL_BRAVIAN")},
		{Name: "John", Email: os.Getenv("EMAIL_JOHN")},
		{Name: "Sheila", Email: os.Getenv("EMAIL_SHEILA")},
	}

	for _, user := range users {
		message := geminiWrapper(user.Name)
		from := os.Getenv("EMAIL_FROM")
		password := os.Getenv("EMAIL_PASS")
		to := []string{user.Email}

		smtpServer := "smtp.gmail.com"
		port := "587"
		auth := smtp.PlainAuth("", from, password, smtpServer)
		err = smtp.SendMail(smtpServer+":"+port, auth, from, to, message)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println("Email Sent")
	}
}

func geminiWrapper(name string) []byte {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		err := client.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()
	model := client.GenerativeModel("gemini-1.5-flash")

	msg := "My name is " + name + ". Analyze the provided JSON weather data and create a meaningful weather forecast summary for the location and date specified. be funny Focus on the likelihood of rain, temperature range, and provide suggestions on what to wear and when to return home based on the weather conditions. Use natural language and make the forecast easy to understand." + getWeather()
	cs := model.StartChat()
	resp, err := cs.SendMessage(ctx, genai.Text(msg))
	if err != nil {
		log.Fatalln(err)
	}

	message := []byte{}
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
				message = append(message, []byte(fmt.Sprintf("%s", part))...)
			}
		}
	}
	return message
}
