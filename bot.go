package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	FBMessageURL    = "https://graph.facebook.com/v14.0/me/messages"
	PageToken       = "EAAPhKhZAz330BAEXyWZApb0M9ICUJII8ofDI8HbYrmFwpMYZCjMHMIRZA3lndDNwVOxtvtekkz3yl8VYg7pA7apdiZCWzBelnQoLA5ZAyAMlOsunvHbRrZBQ2eohuRrxbbQDlZCqZA6Cj4ZBy44JkkIyJw92eEU1cE6zPNJFSxkZBEUt6TzxEcBvE6F"
	MessageResponse = "RESPONSE"
	MarkSeen        = "mark_seen"
	TypingOff       = "typing_off"
	TypingOn        = "typing_on"
)

func sendFBRequest(url string, m interface{}) error {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&m)
	if err != nil {
		log.Printf("sendFBRequest: ", err.Error())
		return err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Printf("NewRequest: ", err.Error())
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = "access_token=" + PageToken
	client := &http.Client{Timeout: time.Second * 30}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("request: ", err.Error())
		return err
	}

	defer resp.Body.Close()
	return nil

}

func sendText(recipient *User, message string) error {
	m := ResponseMessage{
		MessageType: MessageResponse,
		Recipient:   recipient,
		Message: &ResMessage{
			Text: message,
		},
	}
	return sendFBRequest(FBMessageURL, &m)
}

func sendTextQuickReply(recipient *User, message string, replies []QuickReply) error {
	m := ResponseMessage{
		MessageType: MessageResponse,
		Recipient:   recipient,
		Message: &ResMessage{
			Text:       message,
			QuickReply: replies,
		},
	}

	return sendFBRequest(FBMessageURL, &m)
}

func sendAction(recipient *User, action string) error {
	m := ResponseMessage{
		MessageType: MessageResponse,
		Recipient:   recipient,
		Action:      action,
	}
	return sendFBRequest(FBMessageURL, &m)
}

type (
	ExchangeRate struct {
		DateTime string   `xml:"DateTime"`
		ExRate   []ExRate `xml:"Exrate"`
		Source   string   `xml:"Source"`
	}

	ExRate struct {
		CurrencyCode string `xml:"CurrencyCode,attr"`
		CurrencyName string `xml:"CurrencyName,attr"`
		Buy          string `xml:"Buy,attr"`
		Transfer     string `xml:"Transfer,attr"`
		Sell         string `xml:"Sell,attr"`
	}
)

var (
	exRateList     *ExchangeRate
	exRateGroupMap = make(map[string]int)
)

func processMessage(event *Messaging) {
	sendAction(event.Recipient, MarkSeen)
	sendAction(event.Recipient, TypingOn)

	if event.Message.QuickReply != nil {
		if exRateList == nil {
			exRateList = getListCurrency(event.Recipient)
		}
		processQuickReply(event)
		return
	}

	text := strings.ToLower(strings.TrimSpace(event.Message.Text))
	if text == "rate" {
		exRateGroupMap[event.Sender.Id] = 1
		sendExchangeRateList(event.Sender)
	} else {
		sendText(event.Sender, strings.ToUpper(event.Message.Text))
	}
	sendAction(event.Recipient, TypingOff)

}

func processQuickReply(event *Messaging) {
	recipient := event.Sender
	exRateGroup := exRateGroupMap[event.Sender.Id]
	if exRateGroup == 0 {
		exRateGroup = 1
	}
	switch event.Message.QuickReply.Payload {
	case "Next":
		var i int
		if exRateGroup*10 >= len(exRateList.ExRate) {
			exRateGroup = 1
		} else {
			exRateGroup++
		}
		exRateGroupMap[event.Sender.Id] = exRateGroup
		quickRep := []QuickReply{}
		for i = 10 * (exRateGroup - 1); i < 10*exRateGroup && i < len(exRateList.ExRate); i++ {
			exRate := exRateList.ExRate[i]
			quickRep = append(quickRep, QuickReply{ContentType: "text", Title: exRate.CurrencyName, Payload: exRate.CurrencyCode})
		}
		quickRep = append(quickRep, QuickReply{ContentType: "text", Title: "Xem tiếp", Payload: "Next"})
		sendTextQuickReply(recipient, "GotBot cung cấp chức năng xem tỉ giá giữa các đồng ngoại tệ và Việt Nam đồng. \nMời bạn chọn ngoại tệ", quickRep)
	default:
		var exRate ExRate
		for i := 10 * (exRateGroup - 1); i < 10*exRateGroup && i < len(exRateList.ExRate); i++ {
			if exRateList.ExRate[i].CurrencyCode == event.Message.QuickReply.Payload {
				exRate = exRateList.ExRate[i]
				break
			}
		}
		// dont match any item
		if len(exRate.CurrencyCode) == 0 {
			sendText(recipient, "Không có thông tin về ngoại tệ này")
			return
		}

		sendText(recipient, fmt.Sprintf("%s-VND\n Giá mua: %sđ\nGiá bán: %sđ\nGiá chuyển khoản: %sđ\n", exRate.CurrencyName, exRate.Buy, exRate.Sell, exRate.Transfer))

	}

}

func sendExchangeRateList(recipient *User) {
	var (
		exRateGroup = exRateGroupMap[recipient.Id]
	)
	// Get list currency
	exRateList = getListCurrency(recipient)
	quickRep := []QuickReply{}
	for i := 10 * (exRateGroup - 1); i < 10*exRateGroup && i < len(exRateList.ExRate); i++ {
		exRate := exRateList.ExRate[i]
		quickRep = append(quickRep, QuickReply{ContentType: "text", Title: exRate.CurrencyName, Payload: exRate.CurrencyCode})
	}
	quickRep = append(quickRep, QuickReply{ContentType: "text", Title: "Xem tiếp", Payload: "Next"})
	sendTextQuickReply(recipient, "GoBot cung cấp chức năng xem tỉ giá giữa các ngoại tệ và đồng Việt Nam.\nMời bạn chọn ngoại tệ:", quickRep)
}

func getListCurrency(recipient *User) *ExchangeRate {
	exRateList, ok := getExchangeRateVCB()
	if !ok {
		sendText(recipient, "Có lỗi trong quá trình xử lí. Bạn vui lòng gửi lại 'rate' cho tôi nhé. Cảm ơn!")
		return nil
	}
	return exRateList
}

func getExchangeRateVCB() (*ExchangeRate, bool) {
	var exRate *ExchangeRate

	req, err := http.NewRequest("GET", "http://www.vietcombank.com.vn/ExchangeRates/ExrateXML.aspx", nil)
	if err != nil {
		log.Printf("getExrateVCB", err.Error())
		return exRate, false
	}

	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("getExchangeVCB client.Do", err.Error())
		return exRate, false
	}

	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&exRate)
	if err != nil {
		log.Printf("getExchangeVCB decode", err.Error())
		return exRate, false
	}
	return exRate, true

}
