package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ResponseT struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID           int    `json:"id"`
				IsBot        bool   `json:"is_bool"`
				FirstName    string `json:"first_name"`
				LastName     string `json:"last_name"`
				Username     string `json:"username"`
				LanguageCode string `json:"language_code"`
			} `json:"from"`
			Chat struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				Username  string `json:"username"`
				Type      string `json:"type"`
			} `json:"chat"`
			Date int    `json:"date"`
			Text string `json:"text"`
		} `json:"message"`
	} `json:"result"`
}

type UserT struct {
	ID        int
	Username  string
	FirstName string
	LastName  string
	RegDate   int
	LastVisit int
	Messages  []MessagesT
}

// метод1 добавления сообщений
// func (u UserT) addMessage(user UserT, text string, messageTime int) UserT {
// 	message := MessagesT{}
// 	message.Message = text
// 	message.DateMessage = messageTime
// 	user.Messages = append(user.Messages, message)

// 	return user
// }

// метод2 добавления сообщений, где объект модифицируется
func (u *UserT) addMessage(text string, messageTime int) {
	message := MessagesT{}
	message.Message = text
	message.DateMessage = messageTime
	u.Messages = append(u.Messages, message)
}

type MessagesT struct {
	Message     string
	DateMessage int
}

var token string = "6025286750:AAHWYyfw1g4-QCP6iopsR5xkMprILA3vdkI"
var host string = "https://api.telegram.org/bot"

func main() {
	// создаём "базу данных" юзеров
	usersDB := make(map[int]UserT)

	// считываем бд из файлика при запуске (срез байт)
	data, _ := os.ReadFile("db.json")
	// разджейсоним данные и запишем в бд-оперативку
	json.Unmarshal(data, &usersDB)

	// идентификатор последнего сообщения для offset
	lastMessage := 0

	// делаем в цикле каждую секунду следующие действия
	for range time.Tick(time.Second * 2) {

		// отправляем запрос к телеграм апи на получение новых сообщений к боту
		var url string = host + token + "/getUpdates?offset=" + strconv.Itoa(lastMessage)
		// get-запрос
		response, err := http.Get(url)
		// проверка на ошибку, выведем её и дальше код в этой итерации не выполняем
		if err != nil {
			fmt.Println(err)
			continue
		}

		// прочитаем тело запроса из http-протокола в виде среза байт
		data, _ := io.ReadAll(response.Body)

		// распарсим json
		var responseObj ResponseT
		// декодируем json в ассоциативный массив
		json.Unmarshal(data, &responseObj) // разнесем срез байтов согласно структуре

		// посчитаем сколько всего сообщений len()
		number := len(responseObj.Result)

		// если сообщений нет, то дальше код в этой итерации не выполняем
		if number < 1 {
			continue
		}

		// в цикле будем просматривать каждое сообщение
		for i := 0; i < number; i++ {
			// текст сообщения
			text := responseObj.Result[i].Message.Text
			if text == "" {
				fmt.Println("пусто")
			} else {
				fmt.Println(text)
			}

			// чат айди пользователя (кому отвечать)
			chatId := responseObj.Result[i].Message.From.ID
			// время сообщения
			messageTime := responseObj.Result[i].Message.Date
			// данные пользователя
			firstName := responseObj.Result[i].Message.From.FirstName
			lastName := responseObj.Result[i].Message.From.LastName
			username := responseObj.Result[i].Message.From.Username

			// будем регистрировать новых пользователей и обновлять время последнего визита для уже зарегистр
			// провряем зарегистрирован ли уже пользователь
			user, exist := usersDB[chatId]

			if !exist {
				user = UserT{}
				user.ID = chatId
				user.FirstName = firstName
				user.LastName = lastName
				user.Username = username
				user.RegDate = messageTime
				user.LastVisit = messageTime
			} else {
				user.LastVisit = messageTime
			}

			// добавление сообщений

			// без использования метода
			// message := MessagesT{}
			// message.Message = text
			// message.DateMessage = messageTime
			// user.Messages = append(user.Messages, message)

			// с использованием метода1
			// user = user.addMessage(user, text, messageTime)

			// с использованием метода2
			user.addMessage(text, messageTime)

			usersDB[chatId] = user

			// запись юзеров в файл .json
			// 1.создание файла (если такой уже сущ, то просто перезапишется без ошибки)
			file, _ := os.Create("db.json")
			// 2. закодируем в json , но в виде среза байтов
			jsonString, _ := json.Marshal(usersDB)
			// 3. запишем в файл, только не срез байт, а строку string()
			// либо так
			// file.WriteString(string(jsonString))

			// либо так
			file.Write(jsonString)

			// отправка сообщения в ответ в многопоточном режиме
			go sendMessage(chatId, text)
		}

		// запомним update_id последнего сообщения
		lastMessage = responseObj.Result[number-1].UpdateID + 1

		fmt.Println(usersDB)
	}

}

func sendMessage(chatId int, text string) {
	http.Get(host + token + "/sendMessage?chat_id=" + strconv.Itoa(chatId) + "&text=" + text)
}
