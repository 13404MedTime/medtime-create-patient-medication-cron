package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"io"
	"net/http"
	"net/url"
	"sort"

	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	botToken        = "5625907982:AAGf-AKQCngObyXjpxQBWBiKhZhmmq-HP_k"
	chatID          = int64(-4141682093)
	baseUrl         = "https://api.admin.u-code.io"
	logFunctionName = "ucode-template"
	IsHTTP          = true // if this is true banchmark test works.
)

// func main() {
// 	Handle([]byte(""))
// } //

const (
	appId             = "P-JV2nVIRUtgyPO5xRNeYll2mT4F5QG4bS"
	urlConst          = "https://api.admin.u-code.io"
	multipleUpdateUrl = "https://api.admin.u-code.io/v1/object/multiple-update/"

	getListUrl = "https://api.admin.u-code.io/v1/object-slim/get-list/"
)

// const (
// 	getSingleURL = "https://api.admin.u-code.io/v1/object/"
// )

// Handle a serverless request
func Handle(req []byte) string {

	Handler("debug", "hello")

	var (
	// response    Response
	// requestData = map[string]interface{}{} dont touch -> "guid":"4cd0770c-22ae-42b4-8c21-66002ca3d899", "cleints_id":"adda1b93-3a8b-4ff8-922c-bb4f13d559aa"
	)

	var (
		getMadicineTakingUrl  = getListUrl + "medicine_taking" + `?data={"frequency":["always"],"with_relations":true}`
		getMadicineTakingResp = GetListClientApiResponse{}
	)

	body, err := DoRequest(getMadicineTakingUrl, "GET", nil, appId)
	if err != nil {
		return Handler("error", err.Error())
	}
	// Handler("", string(body))
	if err := json.Unmarshal(body, &getMadicineTakingResp); err != nil {
		return Handler("error", err.Error())
	}

	// get all medicine_taking
	// data, err, response := GetListObject(urlConst, "medicine_taking", appId, Request{requestData})
	// if err != nil {
	// 	responseByte, _ := json.Marshal(response)
	// 	return string(responseByte)
	// }

	var (
		medicineTakingResponse = getMadicineTakingResp.Data.Data.Response
		patientMedicationsHour = map[string]interface{}{}
		medicineTakingIds      = []string{}
		medicineTakings        = map[string]interface{}{}
		medicineTakingGuids    = ""
	)

	for _, v := range medicineTakingResponse {
		patientMedicationsHour[v["guid"].(string)] = ""
		medicineTakingIds = append(medicineTakingIds, v["guid"].(string))
		medicineTakingGuids = "\"" + strings.Join(medicineTakingIds, "\",\"") + "\""
		medicineTakings[v["guid"].(string)] = v
	}
	var (
		currentTime = time.Now()
		nextData1   = time.Now().AddDate(0, 0, 10)
		nextData2   = time.Now().AddDate(0, 0, 5)
	)
	nextData2 = time.Date(nextData2.Year(), nextData2.Month(), nextData2.Day(), 23, 59, 59, 0, nextData2.Location())
	currentTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	nextData1 = time.Date(nextData1.Year(), nextData1.Month(), nextData1.Day(), 23, 59, 59, 0, nextData1.Location())

	// get all patient_medications
	// requestData = map[string]interface{}{
	// 	"offset": 0,
	// 	"time_take": map[string]interface{}{
	// 		"$gte": currentTime.Format("2006-01-02T15:04:05.000Z"),
	// 		"$lte": nextData1.Format("2006-01-02T15:04:05.000Z"),
	// 	},
	// 	"medicine_taking_id": medicineTakingIds,
	// }

	// data, err, _ := GetListObject(urlConst, "patient_medication", appId, Request{requestData})
	// if err != nil {
	// 	return Handler("error 1", err.Error())
	// }

	var (
		getPatientMedicationUrl  = getListUrl + "patient_medication" + fmt.Sprintf(`?data={"time_take":{"$gte":"%s","$lte":"%s"},"medicine_taking_id":[%s]}`, currentTime.Format("2006-01-02T15:04:05.000Z"), nextData1.Format("2006-01-02T15:04:05.000Z"), medicineTakingGuids)
		getPatientMedicationResp = GetListClientApiResponse{}
	)

	body, err = DoRequest(getPatientMedicationUrl, "GET", nil, appId)
	if err != nil {
		return Handler("error", err.Error())
	}
	if err := json.Unmarshal(body, &getPatientMedicationResp); err != nil {
		return Handler("error", err.Error())
	}

	patientMedicationsResponse := getPatientMedicationResp.Data.Data.Response

	for _, v := range patientMedicationsResponse {
		medicine_taking_id := v["medicine_taking_id"].(string)
		time_take := v["time_take"].(string)
		if _, ok := patientMedicationsHour[medicine_taking_id]; ok {
			newTime, _ := time.Parse("2006-01-02T15:04:05.000Z", time_take)
			if patientMedicationsHour[medicine_taking_id] == "" {
				patientMedicationsHour[medicine_taking_id] = v
			} else {
				patientMedication := patientMedicationsHour[medicine_taking_id]
				latestTime, _ := time.Parse("2006-01-02T15:04:05.000Z", patientMedication.(map[string]interface{})["time_take"].(string))
				if newTime.After(latestTime) {
					v["time_take"] = time_take
					patientMedicationsHour[medicine_taking_id] = v
				}
			}
		}
	}
	return ""
}

func MultipleUpdateObject(url, tableSlug, appId string, request Request) error {
	_, err := DoRequest(url+"/v1/object/multiple-update/"+tableSlug, "PUT", request, appId)
	// fmt.Println("resp", string(resp), "err", err)
	if err != nil {
		return errors.New("error while updating multiple objects" + err.Error())
	}
	return nil
}

func sortHours(timeStrings []string) ([]time.Time, error) {
	// Parse the time strings into time.Time objects
	times := make([]time.Time, len(timeStrings))
	for i, str := range timeStrings {
		parsedTime, err := time.Parse("15:04:05", str)
		if err != nil {
			// fmt.Println("Error parsing time:", err)
			return nil, err
		}
		parsedTime = parsedTime.Add(time.Hour * -5)
		times[i] = parsedTime
	}

	// Sort the time.Time objects
	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})

	// // Format the sorted times as strings
	// sortedTimeStrings := make([]string, len(times))
	// for i, t := range times {
	// 	sortedTimeStrings[i] = t.Format("15:04:05")
	// }

	return times, nil
}

func getNextDate(current time.Time, days []int, times []time.Time) time.Time {
	nextDate := current

	// Get next hour
	var nextTime time.Time

	for _, t := range times {
		if t.Hour() == current.Hour() {
			if t.Minute() > current.Minute() {
				nextTime = t
				break
			}
		} else if t.Hour() > current.Hour() {
			nextTime = t
			break
		}
	}

	if nextTime == (time.Time{}) {
		nextTime = times[0]
		nextDate = nextDate.AddDate(0, 0, 1)
	}
	// current day of the week
	currentDay := int(nextDate.Weekday())

	// iterate days array and find next upcoming day
	addition := -1
	for _, day := range days {
		if day >= currentDay {
			addition = day - currentDay
			nextDate = nextDate.AddDate(0, 0, day-currentDay)
			break
		}
	}
	if addition == -1 {
		nextDate = nextDate.AddDate(0, 0, days[0]+7-currentDay)
	}

	// Combine the next date and time
	nextDateTime := time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), nextTime.Hour(), nextTime.Minute(), nextTime.Second(), 0, nextDate.Location())
	return nextDateTime
}

// func addToNotifTable(timeTake, preparatId, clientId, appId, tableSlug string) {
// 	notifRequest := Request{
// 		Data: map[string]interface{}{
// 			"client_id":    clientId,
// 			"title":        "Время принятия препарата!",
// 			"body":         "Вам назначен препарат: ",
// 			"title_uz":     "Preparatni qabul qilish vaqti bo'ldi!",
// 			"body_uz":      "Sizga preparat tayinlangan: ",
// 			"is_read":      false,
// 			"preparati_id": preparatId,
// 			"time_take":    timeTake,
// 		},
// 	}
// 	urlConst := "https://api.admin.u-code.io"
// 	CreateObject(urlConst, tableSlug, appId, notifRequest)
// }

func GetListObject(url, tableSlug, appId string, request Request) (GetListClientApiResponse, error, Response) {
	response := Response{}

	getListResponseInByte, err := DoRequest(url+"/v1/object/get-list/"+tableSlug+"?from-ofs=true", "POST", request, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while getting list of object"}
		response.Status = "error"
		return GetListClientApiResponse{}, errors.New("error"), response
	}
	var getListObject GetListClientApiResponse
	err = json.Unmarshal(getListResponseInByte, &getListObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling get list object"}
		response.Status = "error"
		return GetListClientApiResponse{}, errors.New("error"), response
	}
	return getListObject, nil, response
}

func GetSingleObject(url, tableSlug, appId, guid string) (ClientApiResponse, error, Response) {
	response := Response{}

	var getSingleObject ClientApiResponse
	getSingleResponseInByte, err := DoRequest(url+"/v1/object/{table_slug}/{guid}?from-ofs=true", "GET", nil, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while getting single object"}
		response.Status = "error"
		return ClientApiResponse{}, errors.New("error"), response
	}
	err = json.Unmarshal(getSingleResponseInByte, &getSingleObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling single object"}
		response.Status = "error"
		return ClientApiResponse{}, errors.New("error"), response
	}
	return getSingleObject, nil, response
}

func CreateObject(url, tableSlug, appId string, request Request) (Datas, error, Response) {
	response := Response{}

	var createdObject Datas
	createObjectResponseInByte, err := DoRequest(url+"/v1/object/"+tableSlug+"?from-ofs=true&project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", "POST", request, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while creating object"}
		response.Status = "error"
		return Datas{}, errors.New("error"), response
	}
	err = json.Unmarshal(createObjectResponseInByte, &createdObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling create object object"}
		response.Status = "error"
		return Datas{}, errors.New("error"), response
	}
	return createdObject, nil, response
}

func UpdateObject(url, tableSlug, appId string, request Request) (error, Response) {
	response := Response{}

	_, err := DoRequest(url+"/v1/object/{table_slug}?from-ofs=true", "PUT", request, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while updating object"}
		response.Status = "error"
		return errors.New("error"), response
	}
	return nil, response
}

func UpdateObjectMany2Many(url, appId string, request RequestMany2Many) (error, Response) {
	response := Response{}

	_, err := DoRequest(url+"/v1/many-to-many/", "PUT", request, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while updating object"}
		response.Status = "error"
		return errors.New("error"), response
	}
	return nil, response
}

func DeleteObject(url, tableSlug, appId, guid string) (error, Response) {
	response := Response{}

	_, err := DoRequest(url+"/v1/object/{table_slug}/{guid}?from-ofs=true", "DELETE", Request{}, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while updating object"}
		response.Status = "error"
		return errors.New("error"), response
	}
	return nil, response
}

func DoRequest(url string, method string, body interface{}, appId string) ([]byte, error) {
	data, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	request.Header.Add("authorization", "API-KEY")
	request.Header.Add("X-API-KEY", appId)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respByte, nil
}

func Handler(status, message string) string {
	var (
		response Response
		Message  = make(map[string]interface{})
	)

	Send(status + message)
	response.Status = status
	Message["message"] = message
	response.Data = Message
	respByte, _ := json.Marshal(response)
	return string(respByte)

}

func Send(text string) {
	bot, _ := tgbotapi.NewBotAPI(botToken)

	chatID := int64(chatID)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("message from madad payme route function: %s", text))

	bot.Send(msg)

}
