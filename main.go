package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"gopkg.in/yaml.v2"

	uzClient "booking-uz-gov-ua-ticket-checker/booking_uz"
)

const (
	TypeStationFrom   = "TYPE_STATION_FROM"
	ChooseStationFrom = "CHOOSE_STATION_FROM"
	TypeStationTo     = "TYPE_STATION_TO"
	ChooseStationTo   = "CHOOSE_STATION_TO"
	ChooseDate        = "CHOOSE_DATE"
	GetResults        = "GET_RESULTS"
	FindStation       = "FIND_STATION"
)

type credentials struct {
	Token string `yaml:"token"`
}

type train struct {
	fromStation string
	toStation string
	date string
}

var (
	currTrain map[int64]train
	currStep map[int64]string
)

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

func monitoring() {
	// TODO: implement me
}

func OneTimeReplyKeyboard(rows ...[]tgbotapi.KeyboardButton) tgbotapi.ReplyKeyboardMarkup {
	var keyboard [][]tgbotapi.KeyboardButton

	keyboard = append(keyboard, rows...)

	return tgbotapi.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard:       keyboard,
		OneTimeKeyboard: true,
	}
}

func validateDate(rawDate string) (string, error) {
	s := strings.Split(rawDate, ".")
	if len(s) != 2 {
		return "", errors.New("неправильний формат дати")
	}
	day, month := s[0], s[1]
	if d, err := strconv.Atoi(day); err != nil {
		return "", errors.New("день повинен бути цілочисельним числом")
	} else if d < 1 || d > 31 {
		return "", errors.New("день повинен бути у діапазоні 01-31")
	}
	if m, err := strconv.Atoi(month); err != nil {
		return "", errors.New("місяць повинен бути цілочисельним числом")
	} else if m < 1 || m > 12 {
		return "", errors.New("місяць повинен бути у діапазоні 01-12")
	}
	if len(day) < 2 {
		day = "0" + day
	}
	if len(month) < 2 {
		month = "0" + month
	}
	year := time.Now().Year()
	return fmt.Sprintf("%d-%s-%s", year, month, day), nil
}

func getStation(chatId int64, message string, bot *tgbotapi.BotAPI, isFrom bool) (string, string, error) {
	var typeStep, chooseStep string
	if isFrom {
		typeStep, chooseStep = TypeStationFrom, ChooseStationFrom
	} else {
		typeStep, chooseStep = TypeStationTo, ChooseStationTo
	}

	stationsInfo, err := uzClient.Stations(message)
	if err != nil {
		return "", "", err
	}
	potentialStations, err := uzClient.PotentialStations(stationsInfo)
	if err != nil {
		return "", "", err
	}
	if len(potentialStations) > 1 && currStep[chatId] == typeStep {
		msg := tgbotapi.NewMessage(chatId,
			"Оберіть станцію із запропонованих варіантів.\n\nАбо виконайте пошук /search" +
			"\nБот знаходиться на стадії розробки, можливий спам та нестабільність роботи.")

		var keyboard [][]tgbotapi.KeyboardButton
		for _, station := range potentialStations {
			keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(station)))
		}
		msg.ReplyMarkup = OneTimeReplyKeyboard(keyboard...)
		_, _ = bot.Send(msg)
		currStep[chatId] = chooseStep
	} else {
		stationId, err := uzClient.FirstStationId(stationsInfo)
		if err != nil {
			return "", "", err
		}
		return potentialStations[0], stationId, nil
	}
	return "", "", nil
}

func main() {
	currStep = make(map[int64]string)
	currTrain = make(map[int64]train)

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

	go monitoring()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.IsCommand() == false {
			if _, ok := currStep[update.Message.Chat.ID]; !ok {
				continue
			}
			if _, ok := currTrain[update.Message.Chat.ID]; !ok {
				currTrain[update.Message.Chat.ID] = train{}
			}

			if currStep[update.Message.Chat.ID] == TypeStationFrom ||
				currStep[update.Message.Chat.ID] == ChooseStationFrom {
				// Станція відправлення
				stationName, stationId, err := getStation(
					update.Message.Chat.ID,
					update.Message.Text,
					bot,
					true)
				if err != nil {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error() +
						"\nСпробуйте ще раз\nЩоб припинити пошук введіть /stop"))
					continue
				}
				if stationName == "" {
					continue
				}

				var train = currTrain[update.Message.Chat.ID]
				train.fromStation = stationId
				currTrain[update.Message.Chat.ID] = train

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(
					"Станція відправлення: *%s*\n\nТепер вкажіть станцію *прибуття*" +
						"\n*Наприклад:* Львів", stationName))
				msg.ParseMode = "markdown"
				_, _ = bot.Send(msg)

				currStep[update.Message.Chat.ID] = TypeStationTo
				continue
			}

			if currStep[update.Message.Chat.ID] == TypeStationTo ||
				currStep[update.Message.Chat.ID] == ChooseStationTo {
				// Станція прибуття
				stationName, stationId, err := getStation(
					update.Message.Chat.ID,
					update.Message.Text,
					bot,
					false)
				if err != nil {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error() +
						"\nСпробуйте ще раз\nЩоб припинити пошук введіть /stop"))
					continue
				}
				if stationName == "" {
					continue
				}

				var train = currTrain[update.Message.Chat.ID]
				train.toStation = stationId
				currTrain[update.Message.Chat.ID] = train

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(
					"Станція прибуття: *%s*\n\nТепер вкажіть дату поїздки у форматі день.місяць" +
						"\n*Наприклад:* 30.06", stationName))
				msg.ParseMode = "markdown"
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				_, _ = bot.Send(msg)

				currStep[update.Message.Chat.ID] = ChooseDate
				continue
			}

			if currStep[update.Message.Chat.ID] == ChooseDate {
				// Введення дати, виведення результату
				date, err := validateDate(update.Message.Text)
				if err != nil {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error() +
						"\nСпробуйте ще раз\nЩоб припинити пошук введіть /stop"))
					continue
				}
				var train = currTrain[update.Message.Chat.ID]
				train.date = date
				currTrain[update.Message.Chat.ID] = train

				trains, err := uzClient.Trains(
					currTrain[update.Message.Chat.ID].fromStation,
					currTrain[update.Message.Chat.ID].toStation,
					currTrain[update.Message.Chat.ID].date)
				if err != nil {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error() +
						"\nСпробуйте ще раз\nЩоб припинити пошук введіть /stop"))
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
						"*Номер потяга:* %s\n*Звідки:* %s\n*Куди:* %s\n*Виїзд:* %s о %s год\n"+
							"*Прибуття:* %s о %s год\n*Тривалість подорожі:* %s\n\n*Доступні вагони:* \n%s",
						train.TrainId, train.From.Station, train.To.Station, train.From.Date, train.From.Time,
						train.To.Date, train.To.Time, train.TravelTime, strings.Join(availableCarriages, "\n"),
					))
					msg.ParseMode = "markdown"
					_, _ = bot.Send(msg)
				}
				currStep[update.Message.Chat.ID] = GetResults
				continue
			}

			if currStep[update.Message.Chat.ID] == FindStation {
				// Пошук станцій
				stationsInfo, err := uzClient.Stations(update.Message.Text)
				if err != nil {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error() +
						"\nСпробуйте ще раз\nЩоб припинити пошук введіть /stop"))
					continue
				}
				potentialStations, err := uzClient.PotentialStations(stationsInfo)
				if err != nil {
					_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()+
						"\nСпробуйте ще раз\nЩоб припинити пошук введіть /stop"))
					continue
				}
				var msg string
				if len(potentialStations) > 0 {
					msg = "Результати пошуку: \n" + strings.Join(potentialStations, "\n") +
						"\n\nВведіть нові параметри пошуку, або /stop для завершення пошуку"
				} else {
					msg = "За заданими параметрами пошуку нічого не знайдено\n" +
						"Спробуйте змінити Ваш запит, або введіть /stop для завершення пошуку"
				}
				_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
				continue
			}
		} else if update.Message.Command() == "start" {
			_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				"Привіт.\nБот знаходиться на стадії розробки, можливий спам та нестабільність роботи." +
				"\nЩоб перевірити наявність квитків введіть команду /check." +
				"\nЩоб виконати пошук станцій станцій введіть команду /search"))
		} else if update.Message.Command() == "stop" {
			if _, ok := currStep[update.Message.Chat.ID]; ok {
				delete(currStep, update.Message.Chat.ID)
				delete(currTrain, update.Message.Chat.ID)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пошук зупинено")
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				_, _ = bot.Send(msg)
			}
		} else if update.Message.Command() == "check" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(
				"Вкажіть станцію *відправлення*\n*Наприклад:* Вінниця"))
			msg.ParseMode = "markdown"
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = bot.Send(msg)
			currStep[update.Message.Chat.ID] = TypeStationFrom
		} else if update.Message.Command() == "search" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(
				"Вкажіть назву *станції* для пошуку\n*Наприклад:* Вінниця"))
			msg.ParseMode = "markdown"
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			_, _ = bot.Send(msg)
			currStep[update.Message.Chat.ID] = FindStation
		} else if update.Message.Command() == "list" {
			// TODO: implement me
		} else if update.Message.Command() == "help" {
			_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				"Бот знаходиться на стадії розробки, можливий спам та нестабільність роботи." +
				"\nЩоб перевірити наявність квитків введіть команду /check." +
				"\nЩоб виконати пошук станцій станцій введіть команду /search"))
		} else {
			_, _ = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,"Невідома команда"))
		}
	}
}
