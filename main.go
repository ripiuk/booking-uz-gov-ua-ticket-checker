package main

import (
	"io/ioutil"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v2"

	uzClient "BookingUzGovUaTicketCheckerBot/booking_uz"
)

//type trains struct {
//	fromStation string
//	toStation string
//	date string
//}

//var trainsList []map[string]trains // key here should be Chat ID

type credentials struct {
	Token string `yaml:"token"`
}

func (cred *credentials) getCredentials() {
	yamlFile, err := ioutil.ReadFile("credentials.yaml")
	if err != nil {
		log.Printf("FAILED Reading yaml file with credentials #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, cred)
	if err != nil {
		log.Fatalf("Unmarshal Error: %v", err)
	}
}

func monitor() {

}

func main() {
	// Get credentials from yaml file
	var cred credentials
	cred.getCredentials()

	// Init telegram bot
	bot, err := tgbotapi.NewBotAPI(cred.Token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	go monitor()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		var reply string

		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		switch update.Message.Command() {
		case "start":
			reply = "Вітаю"
		case "add":
			//uzClient.GetStations("Вінниця")
			uzClient.GetTrains("", "", "")
			reply = "Створено"
		case "stop":
			reply = "Розсилку зупинено"
		case "list":
			reply = "Список"
		case "help":
			reply = "Тут help message"
		case "find_station":
			uzClient.GetStations("Вінниця")
			reply = "Можливі варіанти:"
		default:
			reply = "Невідома команда"
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)

		bot.Send(msg)
	}
}
