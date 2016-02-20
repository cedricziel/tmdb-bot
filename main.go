package main;

import (
	"os"
	"fmt"
	"github.com/tucnak/telebot"
	"time"
	"log"
	"github.com/ryanbradynd05/go-tmdb"
	"encoding/json"
	"strings"
	"unicode/utf8"
)

var (
	bot *telebot.Bot
	tmdbToken string
	secureBaseUrl string
)

func main() {
	fmt.Println("Loading Telegram bot ..")

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	tmdbToken = os.Getenv("TMDB_TOKEN")

	if telegramToken == "" {
		fmt.Println("You need to set the bot's Telegram token through the 'TELEGRAM_TOKEN' environment variable.")
		return
	}
	if tmdbToken == "" {
		fmt.Println("You need to set the bot's TMDB token through the 'TMDB_TOKEN' environment variable")
		return
	}

	if newBot, err := telebot.NewBot(telegramToken); err != nil {
		fmt.Printf("Token used: %s \n", telegramToken)
		fmt.Println("Error creating telegram bot.", err)
		return
	} else {
		bot = newBot
		fmt.Println("Listening.")
	}

	bot.Messages = make(chan telebot.Message, 1000)
	bot.Queries = make(chan telebot.Query, 1000)

	go query()

	bot.Start(1 * time.Second)
}

// query responds to inline bot queries
func query() {
	var options = make(map[string]string, 3)
	var tmdbApi = tmdb.Init(tmdbToken)

	options["page"] = "1"
	options["language"] = "en"
	options["append_to_response"] = "movie"

	configuration, err := tmdbApi.GetConfiguration()
	if err != nil {
		log.Println("Error getting coniguration", err)
	}

	secureBaseUrl = configuration.Images.SecureBaseURL

	for query := range bot.Queries {
		var results = []telebot.Result{}

		log.Println("--- new query ---")
		log.Println("from:", query.From)
		log.Println("text:", query.Text)

		searchResults, err := tmdbApi.SearchMulti(query.Text, options)
		if err != nil {
			log.Println("error searching tmdb: ", err)
		}

		for _, res := range searchResults.Results {

			var title string

			switch res.MediaType {
			case "tv":
				fmt.Println("Original Name: ", res.OriginalName)
				title = res.OriginalName
				break
			case "movie":
				fmt.Println("Original Name: ", res.OriginalTitle)
				title = res.OriginalTitle
				break
			default:
				title = "no title given"
			}

			imageUrl := secureBaseUrl + "w92" + res.PosterPath
			log.Println("Poster URL: ", secureBaseUrl + "w92" + res.PosterPath)
			log.Println("Vote average : ", res.VoteAverage)

			text := shortenString(title, 100) + "\n"
			text += shortenString(res.Overview, 300) + "\n"
			text += shortenString(imageUrl, 100)

			results = append(results, telebot.ArticleResult{
				Text:        escapeString(text, 500),
				Title:       escapeString(title, 100),
				Description: escapeString(res.Overview, 100),
				ThumbURL:    escapeString(imageUrl, 100),
				Mode:        "Markdown",
			})
		}

		if err := bot.Respond(query, results); err != nil {
			log.Println("ouch:", err)
		}
	}
}

func message() {
	for message := range bot.Messages {
		log.Println("--- new query ---")
		log.Println("text:", message.Text)
		log.Println("sender:", message.Sender)

		if message.Text == "/hi" {
			bot.SendMessage(message.Chat,
				"Hello, " + message.Sender.FirstName + "!", nil)
		}
	}
}

// escapeString escapes a string to be json encodable
func escapeString(str string, len int) string {

	marshaledByteArray, err := json.Marshal(str)
	if err != nil {
		marshaledByteArray = []byte("")
	}

	mString := string(marshaledByteArray)
	mString = strings.Trim(mString, "\"")

	return shortenString(mString, len)
}

// shorten shortens a string to the specified length
func shortenString(s string, i int) string {
	if len(s) < i {
		return s
	}

	if len(s) < 1 {
		return "No Text"
	}

	if utf8.ValidString(s[:i]) {
		return s[:i]
	}
	return s[:i + 1] // or i-1
}
