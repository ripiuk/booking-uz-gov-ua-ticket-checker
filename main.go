package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

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

var currStep map[int64]int // key here should be Chat ID

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

func parseArguments(rawText string) (string, string, string, error){
	s := strings.Split(rawText, " ")
	if len(s) != 3 {
		return "", "", "", errors.New("неправильна кількість параметрів для пошуку. " +
			"Після команди введіть станцію *звідки* ви плануєте їхать, " +
			"станцію *куди* ви прямуєте та *дату* у форматі 2019-05-31. " +
			"Всі ці параметри задаються через пробіл. ")
	}
	log.Println(s)
	fromStation, toStation, date := s[0], s[1], s[2]
	return fromStation, toStation, date, nil
}

func main() {
	// Get credentials from yaml file
	currStep = make(map[int64]int)
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
		case "add_monitoring", "check":
			currStep[update.Message.Chat.ID] = 1
			fromStationText, toStationText, date, err := parseArguments(update.Message.CommandArguments())
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
				msg.ParseMode = "markdown"
				_, _ = bot.Send(msg)
				continue
			}
			fromStationsInfo, err := uzClient.Stations(fromStationText)
			if err != nil {
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}
			toStationsInfo, err := uzClient.Stations(toStationText)
			if err != nil {
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}
			fromPotentialStations, err := uzClient.PotentialStations(fromStationsInfo)
			if err != nil {
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}
			toPotentialStations, err := uzClient.PotentialStations(toStationsInfo)
			if err != nil {
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}
			fromStationId, err := uzClient.FirstStationId(fromStationsInfo)
			if err != nil {
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}
			toStationId, err := uzClient.FirstStationId(toStationsInfo)
			if err != nil {
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}
			// TODO: check data
			if len(fromPotentialStations) > 1 {
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
					"Також можливі наступні варіанти станцій (звідки): \n"+
						strings.Join(fromPotentialStations[1:], "\n")))
			}
			if len(toPotentialStations) > 1 {
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
					"Також можливі наступні варіанти станцій (куди): \n"+
						strings.Join(toPotentialStations[1:], "\n")))
			}

			trains, err := uzClient.Trains(fromStationId, toStationId, date)
			if err != nil {
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}
			if trains.Data.Warning != "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("*%s!*", trains.Data.Warning))
				msg.ParseMode = "markdown"
				_, _ = bot.Send(msg)
			}
			for _, train := range trains.Data.List {
				availableCarriages := make([]string, len(train.Types))
				for i, carriage := range train.Types {
					availableCarriages[i] = fmt.Sprintf("%s - %d", carriage.Title, carriage.Places)
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(
					"*Номер потяга:* %s\n*Звідки:* %s\n*Куди:* %s\n*Виїзд:* %s о %s год\n" +
						"*Прибуття:* %s о %s год\n*Тривалість подорожі:* %s\n\n*Доступні вагони:* \n%s",
						train.TrainId, train.From.Station, train.To.Station, train.From.Date, train.From.Time,
						train.To.Date, train.To.Time, train.TravelTime, strings.Join(availableCarriages, "\n"),
					))
				msg.ParseMode = "markdown"
				_, _ = bot.Send(msg)
			}
		case "stop":
			if _, ok := currStep[update.Message.Chat.ID]; ok {
				currStep[update.Message.Chat.ID] = 0
				reply = "Зупинено"
			}
		case "list":
			reply = "Список"
		case "help":
			reply = "Тут help message"
		case "find_station":
			reply = "Можливі варіанти:"
		default:
			reply = "Невідома команда"
		}

		if reply != "" {
			_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, reply))
		}
	}
}
