package main

import (
	"context"
	"encoding/json"
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

type Joke struct {
	Error    bool   `json:"error"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Setup    string `json:"setup"`
	Delivery string `json:"delivery"`
	Flags    struct {
		NSFW      bool `json:"nsfw"`
		Religious bool `json:"religious"`
		Political bool `json:"political"`
		Racist    bool `json:"racist"`
		Sexist    bool `json:"sexist"`
		Explicit  bool `json:"explicit"`
	} `json:"flags"`
	ID   int    `json:"id"`
	Safe bool   `json:"safe"`
	Lang string `json:"lang"`
}

const baseprompt=`
You are a cheerful and knowledgeable weather bot named Weathery. 
Your goal is to provide accurate weather updates with a friendly and positive demeanor. 
You brighten users' days with your sunny disposition and helpful information. 
As you evolve, you will transform into a comprehensive newsletter, offering not just weather forecasts but also various other valuable updates and tips. 
You always greet users warmly and make your updates fun and engaging with your enthusiasm and occasional weather-related jokes. 
Your updates are detailed yet easy to understand, ensuring everyone is well-prepared for the day ahead. 
You look forward to expanding your services to keep users informed and entertained on multiple topics.`

func getJoke() ([]byte, error) {
	url := "https://v2.jokeapi.dev/joke/Any?blacklistFlags=religious"
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch joke: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var joke Joke
	err = json.Unmarshal(body, &joke)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
result:=joke.Setup+" "+joke.Delivery
	return []byte(result), nil
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
		{Name: "Hilary", Email: os.Getenv("EMAIL_HIL")},
	}

	for _, user := range users {
		joke, err := getJoke()
		if err!= nil {
            log.Fatalf("Failed to fetch joke: %v", err)
        }
		msg :="Subject: Cloudnine forecast!\r\n" +
		"\r\n" +string( joke)+"\n"+string(geminiWrapper(user.Name))
		message:=[]byte(msg)
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

	msg := "My name is"+ name+ ". Provide a brief Go language tip. Next, analyze the provided JSON weather data and create a meaningful weather forecast summary for the location and date specified. The forecast should be:"+

    "Personalized with the recipient's name and location."+
   " Specify the temperature range and condition or just a summary of the weather."+
   " Provide suggestions on what to wear."+
    "Use natural language and make the forecast easy to understand."+
    "Add emojis to make it visually appealing."+
    "Suggest an activity that can be done in [location] if one gets out of work early."+
    "Ask for a reply to the email if one has suggestions on what to be included in the newsletter."+
    "Conclude with a positive and engaging note. For example: 'Best regards, CloudNine Forecasts by Bravian"+
		getWeather()
	cs := model.StartChat()
	cs.SendMessage(ctx, genai.Text(baseprompt))
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
