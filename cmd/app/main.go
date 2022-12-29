package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

var botToken string
var rules string
var topics map[int]string
var topicsCount = 136

// Create a struct that mimics the webhook response body
// https://core.telegram.org/bots/api#update
type webhookReqBody struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
	Callback struct {
		Message struct {
			Text string `json:"text"`
			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`
		} `json:"message"`
		Data string `json:"data"`
	} `json:"callback_query"`
}

type inlineKeyboardMarkup struct {
	AllButtons [][]inlineKeyboardButton `json:"inline_keyboard"`
}

type inlineKeyboardButton struct {
	Text         string `json:"text,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

// Create a struct to conform to the JSON body
// of the send message request
// https://core.telegram.org/bots/api#sendmessage
type sendMessageReqBody struct {
	ChatID      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *inlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// This handler is called everytime telegram sends us a webhook event
func handler(res http.ResponseWriter, req *http.Request) {
	// First, decode the JSON response body
	body := &webhookReqBody{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		fmt.Println("could not decode request body", err)
		return
	}

	var text = body.Message.Text
	if strings.Contains(strings.ToLower(text), "/") {
		handleCommands(body)
	}

	var callback = body.Callback.Data
	if strings.Contains(strings.ToLower(callback), "/next") {
		handleNextCommand(body.Callback.Message.Chat.ID)
	}
}

func handleNextCommand(chatId int64) {
	sendNewTopic(chatId)
}

func handleCommands(body *webhookReqBody) {
	switch body.Message.Text {
	case "/start":
		handleStartCommand(body.Message.Chat.ID)
	case "/rules":
		handleRulesCommand(body.Message.Chat.ID)
	case "/next":
		handleNextCommand(body.Message.Chat.ID)
	}
}

func handleRulesCommand(chatId int64) {
	if err := send(chatId, rules, false); err != nil {
		fmt.Println("error in sending reply on rules command:", err)
		return
	}

	// log a confirmation message if the message is sent successfully
	fmt.Println("rules sent chatid:", chatId)
}

func handleStartCommand(chatId int64) {
	if err := sendNewTopic(chatId); err != nil {
		fmt.Println("error in sending keyboard reply on start command:", err)
		return
	}

	// log a confirmation message if the message is sent successfully
	fmt.Println("keyboard sent")
}

func setupKeyboard() *inlineKeyboardMarkup {

	buttonNext := inlineKeyboardButton{
		Text:         "Следующая тема",
		CallbackData: "/next",
	}

	AllButtons := inlineKeyboardMarkup{[][]inlineKeyboardButton{{buttonNext}}}
	return &AllButtons
}

func send(chatID int64, text string, keyboard bool) error {
	// Create the request body struct
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   text,
	}

	if keyboard == true {
		reqBody.ReplyMarkup = setupKeyboard()
	}
	// Create the JSON body from the struct
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Send a post request with your token
	res, err := http.Post(fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", botToken), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status" + res.Status)
	}

	return nil
}

func sendNewTopic(chatID int64) error {
	var text = topics[rand.Intn(topicsCount)]
	if err := send(chatID, text, true); err != nil {
		fmt.Println("error in sending new topic reply command: ", err)
		return err
	}

	// log a confirmation message if the message is sent successfully
	fmt.Println("new topic sent")
	return nil
}

func main() {
	readTopics()
	readRules()
	botToken = os.Getenv("BOTTOKEN")
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.ListenAndServe(":"+port, http.HandlerFunc(handler))
}

func readTopics() {
	file, err := os.Open("topics.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	topics = make(map[int]string, topicsCount)

	scanner := bufio.NewScanner(file)

	var index = 0
	for scanner.Scan() {
		topics[index] = scanner.Text()
		index++
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}

func readRules() {
	file, err := os.Open("rules.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rules += scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
}
