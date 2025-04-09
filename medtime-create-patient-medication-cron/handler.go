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

	// for medicine_taking_id, v := range patientMedicationsHour {
	// 	var (
	// 		timeTakeStr string
	// 		timeTake    time.Time
	// 	)

	// 	patientMedicationLastHour, ok := v.(map[string]interface{})
	// 	if ok {
	// 		timeTakeStr, _ = patientMedicationLastHour["time_take"].(string)
	// 		timeTake, _ = time.Parse("2006-01-02T15:04:05.000Z", timeTakeStr)
	// 	} else {
	// 		timeTake = time.Now()
	// 	}
	// 	// fmt.Println(timeTake)
	// 	if nextData2.After(timeTake) {
	// 		Send("medicine_taking_id: " + medicine_taking_id + "TimeTake: " + timeTake.String() + "nextDate" + nextData2.String())
	// 		medicineTaking, ok := medicineTakings[medicine_taking_id]
	// 		if ok {
	// 			var (
	// 				medicine          Medicine
	// 				clientId          string
	// 				medicineTakingInt = medicineTaking.(map[string]interface{})
	// 			)

	// 			if medicineTakingInt["cleints_id"] != nil {
	// 				clientId = medicineTakingInt["cleints_id"].(string)
	// 			}

	// 			body := medicineTakingInt["json_body"].(string)
	// 			err = json.Unmarshal([]byte(body), &medicine)
	// 			if err != nil {
	// 				return Handler("error 2", err.Error())

	// 			}

	// 			days := []int{0, 1, 2, 3, 4, 5, 6}

	// 			frequencyInt, _ := medicineTakingInt["frequency"].([]interface{})
	// 			if len(frequencyInt) < 1 {
	// 				return Handler("error 3", err.Error())

	// 			}

	// 			sort.Ints(days)

	// 			timeString := medicine.HoursOfDay
	// 			sortedTimes, _ := sortHours(timeString)
	// 			if len(sortedTimes) < 1 {
	// 				continue
	// 			}

	// 			var dosage float64
	// 			dosageStr, ok := medicineTakingInt["dosage"].(string)
	// 			if ok {
	// 				dosageInt, err := strconv.Atoi(dosageStr)
	// 				if err != nil {
	// 					return Handler("error 3", err.Error())

	// 				}
	// 				dosage = float64(dosageInt)

	// 			} else {

	// 				dosage, ok = medicineTakingInt["dosage"].(float64)
	// 				if !ok {
	// 					Send("failed to get dosage; medicineTakingGuid:" + medicineTakingInt["guid"].(string))
	// 					continue

	// 				}

	// 			}

	// 			var (
	// 				afterFoodInt  = medicineTakingInt["description"].([]interface{})
	// 				afterFoodList = []string{}
	// 			)

	// 			for _, v := range afterFoodInt {
	// 				val, _ := v.(string)
	// 				afterFoodList = append(afterFoodList, val)
	// 			}

	// 			var (
	// 				boolAfterFood = afterFoodList[0]
	// 				preparatId, _ = medicineTakingInt["preparati_id"].(string)
	// 				requests      = []map[string]interface{}{}
	// 				notifRequests = MultipleUpdateRequest{}
	// 				currentTime   =  timeTake
	// 			)
	// 			// fmt.Println(, "currentTime", currentTime)
	// 			for nextData1.After(currentTime) {
	// 				timeee := getNextDate(currentTime, days, sortedTimes)
	// 				fmt.Println("timeee", timeee)
	// 				currentTime = timeee
	// 				var (
	// 					serverTime           = currentTime
	// 					stringTime           = serverTime.Format("2006-01-02T15:04:05.000Z")
	// 					preparatName, _      = medicineTakingInt["medicine_name"].(string)
	// 					createtObjectRequest = map[string]interface{}{
	// 						"medicine_taking_id": medicineTakingInt["guid"].(string),
	// 						"time_take":          stringTime,
	// 						"before_after_food":  boolAfterFood,
	// 						"cleints_id":         clientId,
	// 						"preparati_id":       preparatId,
	// 						"is_from_patient":    true,
	// 						"count":              dosage,
	// 						"preparat_name":      preparatName,
	// 					}
	// 				)

	// 				requests = append(requests, createtObjectRequest)

	// 				notifRequests.Data.Objects = append(notifRequests.Data.Objects, map[string]interface{}{
	// 					"client_id":    clientId,
	// 					"title":        "Время принятия препарата!",
	// 					"body":         "Вам назначен препарат: ",
	// 					"title_uz":     "Preparatni qabul qilish vaqti bo'ldi!",
	// 					"body_uz":      "Sizga preparat tayinlangan: ",
	// 					"is_read":      false,
	// 					"preparati_id": preparatId,
	// 					"time_take":    stringTime,
	// 				})
	// 			}

	// 			leng := len(requests)

	// 			req := Request{
	// 				Data: map[string]interface{}{
	// 					"objects": requests,
	// 				},
	// 			}
	// 			// fmt.Println("req --- >", req)
	// 			Send("successfully created patient medications" + medicineTakingInt["guid"].(string) + "clients_id" + clientId + "preparati_id" + preparatId + "time_take" + timeTakeStr + "elements count " + strconv.Itoa(leng))

	// 			jsonData, err := json.Marshal(req)
	// 			if err != nil {
	// 				Send("error marshal mult update req " + err.Error())
	// 			}

	// 			Send(string(jsonData))

	// 			multipleUpdateTime := time.Now()

	// 			err = MultipleUpdateObject(urlConst, "patient_medication", appId, req)
	// 			if err != nil {
	// 				return Handler("error 4", err.Error())

	// 			}
	// 			timeTook := time.Since(multipleUpdateTime).Seconds()
	// 			s := fmt.Sprintf("%v", timeTook)
	// 			Send(s + " seconds took to multiple update" + strconv.Itoa(leng) + "objects")

	// 			_, err = DoRequest(multipleUpdateUrl+"notifications", "PUT", notifRequests, appId)
	// 			if err != nil {
	// 				return Handler("error 5", err.Error())

	// 			}

	// 			// req = Request{
	// 			// 	Data: map[string]interface{}{
	// 			// 		"objects": notifRequests,
	// 			// 	},
	// 			// }

	// 			// err = MultipleUpdateObject(urlConst, "notifications", appId, req)
	// 			// if err != nil {
	// 			// 	response.Data = map[string]interface{}{"error in" + notifSlug + "multiple update, message": err.Error()}
	// 			// 	response.Status = "error"
	// 			// 	responseByte, _ := json.Marshal(response)
	// 			// 	return string(responseByte)
	// 			// }

	// 		} else {
	// 			Send("medicine taking not found")
	// 			// medicine taking id not found no need to create any patient medications since we dont have any continious (always) medicine taking
	// 			// fmt.Println("k", medicine_taking_id, "v", v.(map[string]interface{})["time_take"].(string), v.(map[string]interface{})["preparat_name"].(string), v.(map[string]interface{})["preparati_id"].(string))
	// 		}
	// 	}
	// }

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

// func convertToWeekday(day string) time.Weekday {
// 	switch day {
// 	case "sunday":
// 		return time.Sunday
// 	case "monday":
// 		return time.Monday
// 	case "tuesday":
// 		return time.Tuesday
// 	case "wednesday":
// 		return time.Wednesday
// 	case "thursday":
// 		return time.Thursday
// 	case "friday":
// 		return time.Friday
// 	case "saturday":
// 		return time.Saturday
// 	default:
// 		return time.Sunday // Return a default value (Sunday) in case of an invalid day.
// 	}
// }

// func getDatesInRange(startDate, endDate time.Time, dayTimes []DayTime) []time.Time {
// 	var datesInRange []time.Time

// 	for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
// 		for _, dt := range dayTimes {
// 			day := convertToWeekday(dt.Day)
// 			if current.Weekday() == day {
// 				t, _ := time.Parse("15:04:05", dt.Time)
// 				dateTime := current.Add(time.Hour*time.Duration(t.Hour()) + time.Minute*time.Duration(t.Minute()))
// 				datesInRange = append(datesInRange, dateTime)
// 			}
// 		}
// 	}

// 	return datesInRange
// }

// func calculatePillDates(startDate, endDate time.Time, customData CustomDataObj) []time.Time {
// 	var pillDates []time.Time

// 	// Parse time from customData.Time
// 	customTime, _ := time.Parse("15:04:05", customData.Time)

// 	// Define the duration based on the cycle_name and cycle_count
// 	var duration time.Duration
// 	switch customData.CycleName {
// 	case "day":
// 		duration = time.Duration(customData.CycleCount) * 24 * time.Hour
// 	case "month":
// 		duration = time.Duration(customData.CycleCount) * 30 * 24 * time.Hour
// 	default:
// 		return pillDates
// 	}

// 	// Initialize the current date as the start date
// 	currentDate := startDate

// 	// Loop to calculate pill dates between startDate and endDate
// 	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
// 		// Combine currentDate and customTime to get the pillDate
// 		pillDate := currentDate.Add(time.Duration(customTime.Hour()) * time.Hour)
// 		pillDate = pillDate.Add(time.Duration(customTime.Minute()) * time.Minute)
// 		pillDate = pillDate.Add(time.Hour * -5)
// 		// Add the pillDate to the result
// 		pillDates = append(pillDates, pillDate)

// 		// Move currentDate to the next cycle based on the duration
// 		currentDate = currentDate.Add(duration)
// 	}

// 	return pillDates
// }

// week logics

// func getDatesInRangeWeek(startDate, endDate string, days []string, cycleCount int, hourStr string) []DateObject {
// 	layout := "2006-01-02"
// 	start, err := time.Parse(layout, startDate)
// 	if err != nil {
// 		panic(err)
// 	}
// 	end, err := time.Parse(layout, endDate)
// 	if err != nil {
// 		panic(err)
// 	}

// 	currentTime := time.Now()
// 	if hourStr == "" {
// 		hourStr = currentTime.Format("15:04:05")
// 	} else {
// 		parsedTime, err := time.Parse("15:04:05", hourStr)
// 		if err != nil {
// 			// fmt.Println("Error parsing time:", err)
// 			panic(err)
// 		}
// 		subtractedTime := parsedTime.Add(-5 * time.Hour)
// 		outputLayout := "15:04:05"

// 		hourStr = subtractedTime.Format(outputLayout)
// 	}

// 	dayDates := make([]DateObject, 0)

// 	for !start.After(end) {
// 		for _, day := range days {
// 			weekday := getWeekday(day)
// 			if start.Weekday() == weekday {
// 				dateStr := start.Format(layout)
// 				foundDay := false

// 				// Check if the day already exists in the result slice
// 				for i, obj := range dayDates {
// 					if obj.Day == day {
// 						dayDates[i].Dates = append(dayDates[i].Dates, dateStr)
// 						foundDay = true
// 						break
// 					}
// 				}

// 				// If the day doesn't exist in the result slice, add a new DateObject
// 				if !foundDay {
// 					dayDates = append(dayDates, DateObject{
// 						Day:   day,
// 						Hour:  hourStr, // You can set your desired hour here
// 						Dates: []string{dateStr},
// 					})
// 				}
// 			}
// 		}
// 		start = start.AddDate(0, 0, 1)
// 	}

// 	// Add the hour to the dates and format them as "2023-07-11T14:00:00.000Z"
// 	for i, obj := range dayDates {
// 		for j, dateStr := range obj.Dates {
// 			dayDates[i].Dates[j] = dateStr + "T" + obj.Hour + ".000Z"
// 		}
// 	}

// 	return dayDates
// }

// func getWeekday(day string) time.Weekday {
// 	switch day {
// 	case "monday":
// 		return time.Monday
// 	case "tuesday":
// 		return time.Tuesday
// 	case "wednesday":
// 		return time.Wednesday
// 	case "thursday":
// 		return time.Thursday
// 	case "friday":
// 		return time.Friday
// 	case "saturday":
// 		return time.Saturday
// 	case "sunday":
// 		return time.Sunday
// 	default:
// 		panic("Invalid weekday")
// 	}
// }

type DateObject struct {
	Day   string
	Hour  string
	Dates []string
}

// Datas This is response struct from create
type Datas struct {
	Data struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	} `json:"data"`
}

// ClientApiResponse This is get single api response
type ClientApiResponse struct {
	Data ClientApiData `json:"data"`
}

type ClientApiData struct {
	Data ClientApiResp `json:"data"`
}

type ClientApiResp struct {
	Response map[string]interface{} `json:"response"`
}

type Response struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// NewRequestBody's Data (map) field will be in this structure
//.   fields
// objects_ids []string
// table_slug string
// object_data map[string]interface
// method string
// app_id string

// but all field will be an interface, you must do type assertion

type HttpRequest struct {
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Headers http.Header `json:"headers"`
	Params  url.Values  `json:"params"`
	Body    []byte      `json:"body"`
}

type AuthData struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type NewRequestBody struct {
	RequestData HttpRequest            `json:"request_data"`
	Auth        AuthData               `json:"auth"`
	Data        map[string]interface{} `json:"data"`
}
type Request struct {
	Data map[string]interface{} `json:"data"`
}

type RequestMany2Many struct {
	IdFrom    string   `json:"id_from"`
	IdTo      []string `json:"id_to"`
	TableFrom string   `json:"table_from"`
	TableTo   string   `json:"table_to"`
}

// GetListClientApiResponse This is get list api response
type GetListClientApiResponse struct {
	Data GetListClientApiData `json:"data"`
}

type GetListClientApiData struct {
	Data GetListClientApiResp `json:"data"`
}

type GetListClientApiResp struct {
	Response []map[string]interface{} `json:"response"`
}

type CustomDataObj struct {
	CycleName  string   `json:"cycle_name"`
	CycleCount int      `json:"cycle_count"`
	Time       string   `json:"time"`
	Dates      []string `json:"dates"`
}

type DayTime struct {
	Day  string `json:"day"`
	Time string `json:"time"`
}
type DateTime struct {
	Date string `json:"date"`
	Time string `json:"time"`
}

type Medicine struct {
	Type            string        `json:"type"`
	DayData         []string      `json:"dayData"`
	CustomData      CustomDataObj `json:"customData"`
	WeekData        []DayTime     `json:"weekData"`
	MonthData       []DateTime    `json:"monthData"`
	BeforeAfterFood string        `json:"before_after_food"`
	StartDate       string        `json:"start_date"`
	EndDate         string        `json:"end_date"`
	CurrentAmount   int           `json:"current_amount"`
	DaysOfWeek      []int         `json:"days_of_week"`
	HoursOfDay      []string      `json:"hours_of_day"`
	WithoutBreak    bool          `json:"without_break"`
}

type MultipleUpdateRequest struct {
	Data struct {
		Objects []map[string]interface{} `json:"objects"`
	} `json:"data"`
}
