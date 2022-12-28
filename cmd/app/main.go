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
)

var botToken string
var topics map[int]string
var topicsCount = 136
var test map[int]int = map[int]int{}

// Create a struct that mimics the webhook response body
// https://core.telegram.org/bots/api#update
type webhookReqBody struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

// This handler is called everytime telegram sends us a webhook event
func Handler(res http.ResponseWriter, req *http.Request) {
	// First, decode the JSON response body
	body := &webhookReqBody{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		fmt.Println("could not decode request body", err)
		return
	}

	if err := sendNewQuestion(body.Message.Chat.ID); err != nil {
		fmt.Println("error in sending reply:", err)
		return
	}

	// log a confirmation message if the message is sent successfully
	fmt.Println("reply sent")
}

// Create a struct to conform to the JSON body
// of the send message request
// https://core.telegram.org/bots/api#sendmessage
type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func sendNewQuestion(chatID int64) error {
	// Create the request body struct
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   topics[rand.Intn(topicsCount)],
	}
	//fmt.Println("sending reply:", topics[rand.Intn(topicsCount)], " | random:", rand.Intn(topicsCount))

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

func main() {
	_, exist := test[0]
	if !exist {
		test[0] = 9
		fmt.Println("there is no value in memory")
	} else {
		fmt.Printf("got value in memory: %v", test[0])
	}
	readTopics()
	botToken = os.Getenv("BOTTOKEN")
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.ListenAndServe(":"+port, http.HandlerFunc(Handler))
}

func readTopics() {
	file, err := os.Open("topics.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	topics = make(map[int]string, 100)

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
