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
	url := "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/kisumu/today?unitGroup=metric&include=events&key=" + os.Getenv("VISUAL_KEY") + "&contentType=json"
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
		{Name: "Hillary", Email: os.Getenv("EMAIL_HILL")},

	}

	for _, user := range users {
		message := geminiWrapper(user.Name)
		from := os.Getenv("EMAIL_FROM")
		password := os.Getenv("EMAIL_PASS")
		to := []string{user.Email}

		smtpServer := "smtp.gmail.com"
		port := "587"
		auth := smtp.PlainAuth("", from, password, smtpServer)
		err := smtp.SendMail(smtpServer+":"+port, auth, from, to, message)
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


	msg := "My name is " + name + ". Tell a joke first. Then, provide a brief Go language tip. Next, analyze the provided JSON weather data and create a meaningful weather forecast summary for the location and date specified. The forecast should be:\n\n" +
		"1. Personalized with the recipient's name and location.\n" +
		"2. Include a joke at the beginning.\n" +
		"3. Mention the likelihood of rain.\n" +
		"4. Specify the temperature range.\n" +
		"5. Provide suggestions on what to wear in the morning, afternoon, and evening.\n" +
		"6. Use natural language and make the forecast easy to understand.\n" +
		"7. Add emojis to make it visually appealing.\n" +
		"8. If possible suggest an activity that can be done in kisumu if one does get out of work early.\n" +
		"9. Conclude with a positive and engaging note. eg best regards CloudNine Forecasts by bravian\n" +
		getWeather()
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
