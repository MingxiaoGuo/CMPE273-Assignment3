package main

import (
	"Assignment3/permutation"
	"Assignment3/structure"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

var id int
var hmap map[int]structure.DataStorage

var session *mgo.Session

func CreateLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	var locationInfo structure.Request
	json.NewDecoder(req.Body).Decode(&locationInfo)
	response := structure.Response{}
	response.Id = bson.NewObjectId()
	response.Name = locationInfo.Name
	response.Address = locationInfo.Address
	response.City = locationInfo.City
	response.State = locationInfo.State
	response.Zip = locationInfo.Zip
	requestURL := createURL(response.Address, response.City, response.State)
	getLocation(&response, requestURL)
	session.DB("cmpe273_project").C("hello").Insert(response)

	location, _ := json.Marshal(response)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(201)
	fmt.Fprintf(rw, "%s", location)
}

func CreateTrip(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	var tripRequest structure.TripRequest
	json.NewDecoder(req.Body).Decode(&tripRequest)

	tripData := structure.DataStorage{}
	tripResponse := structure.TripResponse{}
	tripResponse.Id = getID()
	tripResponse.Status = "planning"
	tripResponse.Starting_from_location_id = tripRequest.Starting_from_location_id
	tripResponse.Best_route_location_id = tripRequest.Location_ids

	getBestRoute(&tripResponse, &tripData, tripRequest.Starting_from_location_id, tripRequest.Location_ids)

	hmap[tripResponse.Id] = tripData

	trip, _ := json.Marshal(tripResponse)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(201)
	fmt.Fprintf(rw, "%s", trip)
}

func GetLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}
	userId := bson.ObjectIdHex(id)
	response := structure.Response{}
	if err := session.DB("cmpe273_project").C("hello").FindId(userId).One(&response); err != nil {
		w.WriteHeader(404)
		return
	}
	location, _ := json.Marshal(response)
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", location)
}

func GetTrip(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tId := p.ByName("trip_id")
	checkID, _ := strconv.Atoi(tId)
	var tripData structure.DataStorage
	findTarget := false

	for key, value := range hmap {
		if key == checkID {
			tripData = value
			findTarget = true
		}
	}

	if findTarget == false {
		w.WriteHeader(404)
		return
	}

	tripResponse := structure.TripResponse{}
	tripResponse.Id = tripData.Id
	tripResponse.Status = tripData.Status
	tripResponse.Starting_from_location_id = tripData.Starting_from_location_id
	tripResponse.Best_route_location_id = tripData.Best_route_location_id
	tripResponse.Total_uber_costs = tripData.Total_uber_costs
	tripResponse.Total_distance = tripData.Total_distance
	tripResponse.Total_uber_duration = tripData.Total_uber_duration

	trip, _ := json.Marshal(tripResponse)
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", trip)
}

func UpdateLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}
	var locationInfo structure.Request
	json.NewDecoder(r.Body).Decode(&locationInfo)
	response := structure.Response{}
	response.Id = bson.ObjectIdHex(id)
	response.Name = locationInfo.Name
	response.Address = locationInfo.Address
	response.City = locationInfo.City
	response.State = locationInfo.State
	response.Zip = locationInfo.Zip
	requestURL := createURL(response.Address, response.City, response.State)
	getLocation(&response, requestURL)

	if err := session.DB("cmpe273_project").C("hello").Update(bson.M{"_id": response.Id}, bson.M{"$set": bson.M{"address": response.Address, "city": response.City, "state": response.State, "zip": response.Zip, "coordinate.lat": response.Coordinate.Lat, "coordinate.lng": response.Coordinate.Lng}}); err != nil {
		w.WriteHeader(404)
		return
	}

	if err := session.DB("cmpe273_project").C("hello").FindId(response.Id).One(&response); err != nil {
		w.WriteHeader(404)
		return
	}

	location, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", location)
}

func CarRequest(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tId := p.ByName("trip_id")
	checkID, _ := strconv.Atoi(tId)
	var tripData structure.DataStorage
	findTarget := false

	for key, value := range hmap {
		if key == checkID {
			tripData = value
			findTarget = true
		}
	}

	if findTarget == false {
		w.WriteHeader(404)
		return
	}

	var startLat float64
	var startLng float64
	var endLat float64
	var endLng float64
	carRes := structure.CarResponse{}
	response := structure.Response{}

	if tripData.Index == 0 {
		if err := session.DB("cmpe273_project").C("hello").FindId(tripData.Starting_from_location_id).One(&response); err != nil {
			return
		}
		startLat = response.Coordinate.Lat
		startLng = response.Coordinate.Lng

		if err := session.DB("cmpe273_project").C("hello").FindId(tripData.Best_route_location_id[0]).One(&response); err != nil {
			return
		}
		endLat = response.Coordinate.Lat
		endLng = response.Coordinate.Lng
		uberAPI(&carRes, tripData, startLat, startLng, endLat, endLng)
		carRes.Status = "requesting"
		carRes.Starting_from_location_id = tripData.Starting_from_location_id
		carRes.Next_destination_location_id = tripData.Best_route_location_id[0]
	} else if tripData.Index == len(tripData.Best_route_location_id) {
		if err := session.DB("cmpe273_project").C("hello").FindId(tripData.Best_route_location_id[len(tripData.Best_route_location_id)-1]).One(&response); err != nil {
			return
		}
		startLat = response.Coordinate.Lat
		startLng = response.Coordinate.Lng

		if err := session.DB("cmpe273_project").C("hello").FindId(tripData.Starting_from_location_id).One(&response); err != nil {
			return
		}
		endLat = response.Coordinate.Lat
		endLng = response.Coordinate.Lng
		uberAPI(&carRes, tripData, startLat, startLng, endLat, endLng)
		carRes.Status = "requesting"
		carRes.Starting_from_location_id = tripData.Best_route_location_id[len(tripData.Best_route_location_id)-1]
		carRes.Next_destination_location_id = tripData.Starting_from_location_id
	} else if tripData.Index > len(tripData.Best_route_location_id) {
		carRes.Status = "finished"
		carRes.Starting_from_location_id = tripData.Starting_from_location_id
		carRes.Next_destination_location_id = tripData.Starting_from_location_id
	} else {
		if err := session.DB("cmpe273_project").C("hello").FindId(tripData.Best_route_location_id[tripData.Index-1]).One(&response); err != nil {
			return
		}
		startLat = response.Coordinate.Lat
		startLng = response.Coordinate.Lng

		if err := session.DB("cmpe273_project").C("hello").FindId(tripData.Best_route_location_id[tripData.Index]).One(&response); err != nil {
			return
		}
		endLat = response.Coordinate.Lat
		endLng = response.Coordinate.Lng
		uberAPI(&carRes, tripData, startLat, startLng, endLat, endLng)
		carRes.Status = "requesting"
		carRes.Starting_from_location_id = tripData.Best_route_location_id[tripData.Index-1]
		carRes.Next_destination_location_id = tripData.Best_route_location_id[tripData.Index]
	}

	carRes.Id = tripData.Id
	carRes.Best_route_location_id = tripData.Best_route_location_id
	carRes.Total_uber_costs = tripData.Total_uber_costs
	carRes.Total_uber_duration = tripData.Total_uber_duration
	carRes.Total_distance = tripData.Total_distance

	tripData.Index = tripData.Index + 1
	hmap[tripData.Id] = tripData

	trip, _ := json.Marshal(carRes)
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", trip)
}

func DeleteLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}
	userId := bson.ObjectIdHex(id)

	if err := session.DB("cmpe273_project").C("hello").RemoveId(userId); err != nil {
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
}

func main() {
	id = 0
	hmap = make(map[int]structure.DataStorage)

	session, _ = mgo.Dial("mongodb://admin:admin@ds045064.mongolab.com:45064/cmpe273_project")

	defer session.Close()

	router := httprouter.New()
	router.GET("/locations/:id", GetLocation)
	router.GET("/trips/:trip_id", GetTrip)
	router.POST("/locations/", CreateLocation)
	router.POST("/trips/", CreateTrip)
	router.DELETE("/locations/:id", DeleteLocation)
	router.PUT("/locations/:id", UpdateLocation)
	router.PUT("/trips/:trip_id/request", CarRequest)
	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router,
	}
	server.ListenAndServe()
}

func createURL(address string, city string, state string) string {
	var getURL string
	spStr := strings.Split(address, " ")
	for i := 0; i < len(spStr); i++ {
		if i == 0 {
			getURL = spStr[i] + "+"
		} else if i == len(spStr)-1 {
			getURL = getURL + spStr[i] + ","
		} else {
			getURL = getURL + spStr[i] + "+"
		}
	}
	spStr = strings.Split(city, " ")
	for i := 0; i < len(spStr); i++ {
		if i == 0 {
			getURL = getURL + "+" + spStr[i]
		} else if i == len(spStr)-1 {
			getURL = getURL + "+" + spStr[i] + ","
		} else {
			getURL = getURL + "+" + spStr[i]
		}
	}
	getURL = getURL + "+" + state
	return getURL
}

func getLocation(response *structure.Response, formatURL string) {
	urlLeft := "http://maps.google.com/maps/api/geocode/json?address="
	urlRight := "&sensor=false"
	urlFormat := urlLeft + formatURL + urlRight

	getLocation, err := http.Get(urlFormat)
	if err != nil {
		fmt.Println("Get Location Error", err)
		panic(err)
	}

	body, err := ioutil.ReadAll(getLocation.Body)
	if err != nil {
		fmt.Println("Get Location Error", err)
		panic(err)
	}

	var data structure.MyJsonName
	byt := []byte(body)
	if err := json.Unmarshal(byt, &data); err != nil {
		panic(err)
	}
	response.Coordinate.Lat = data.Results[0].Geometry.Location.Lat
	response.Coordinate.Lng = data.Results[0].Geometry.Location.Lng
}

func getBestRoute(tripResponse *structure.TripResponse, tripData *structure.DataStorage, originId bson.ObjectId, targetId []bson.ObjectId) {
	pmtTarget, err := permutation.NewPerm(targetId, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	res := make([][]bson.ObjectId, 0, 0)
	routePrice := make([]int, 0, 0)
	routeDuration := make([]int, 0, 0)
	routeDistance := make([]float64, 0, 0)
	curPrice := 0
	curDuration := 0
	curDistance := 0.0
	for result, err := pmtTarget.Next(); err == nil; result, err = pmtTarget.Next() {
		for i := 0; i <= len(result.([]bson.ObjectId)); i++ {
			var startLat float64
			var startLng float64
			var endLat float64
			var endLng float64
			minPrice := 0
			minDuration := 0
			minDistance := 0.0
			response := structure.Response{}
			if i == 0 {
				if err := session.DB("cmpe273_project").C("hello").FindId(originId).One(&response); err != nil {
					return
				}
				startLat = response.Coordinate.Lat
				startLng = response.Coordinate.Lng

				if err := session.DB("cmpe273_project").C("hello").FindId((result.([]bson.ObjectId))[i]).One(&response); err != nil {
					return
				}
				endLat = response.Coordinate.Lat
				endLng = response.Coordinate.Lng
			} else if i == len(result.([]bson.ObjectId)) {
				if err := session.DB("cmpe273_project").C("hello").FindId((result.([]bson.ObjectId))[i-1]).One(&response); err != nil {
					return
				}
				startLat = response.Coordinate.Lat
				startLng = response.Coordinate.Lng

				if err := session.DB("cmpe273_project").C("hello").FindId(originId).One(&response); err != nil {
					return
				}
				endLat = response.Coordinate.Lat
				endLng = response.Coordinate.Lng
			} else {
				if err := session.DB("cmpe273_project").C("hello").FindId((result.([]bson.ObjectId))[i-1]).One(&response); err != nil {
					return
				}
				startLat = response.Coordinate.Lat
				startLng = response.Coordinate.Lng

				if err := session.DB("cmpe273_project").C("hello").FindId((result.([]bson.ObjectId))[i]).One(&response); err != nil {
					return
				}
				endLat = response.Coordinate.Lat
				endLng = response.Coordinate.Lng
			}

			urlLeft := "https://api.uber.com/v1/estimates/price?"
			// Server token cannot be told to others
			urlRight := "start_latitude=" + strconv.FormatFloat(startLat, 'f', -1, 64) + "&start_longitude=" + strconv.FormatFloat(startLng, 'f', -1, 64) + "&end_latitude=" + strconv.FormatFloat(endLat, 'f', -1, 64) + "&end_longitude=" + strconv.FormatFloat(endLng, 'f', -1, 64) + "&server_token=PLEASE PUT YOUR OWN SERVER TOKEN HERE"
			urlFormat := urlLeft + urlRight

			getPrices, err := http.Get(urlFormat)
			if err != nil {
				fmt.Println("Get Prices Error", err)
				panic(err)
			}

			var data structure.UberAPIResponse
			json.NewDecoder(getPrices.Body).Decode(&data)
			minPrice = data.Prices[0].LowEstimate
			minDuration = data.Prices[0].Duration
			minDistance = data.Prices[0].Distance
			for i := 0; i < len(data.Prices); i++ {
				if minPrice > data.Prices[i].LowEstimate && data.Prices[i].LowEstimate > 0 {
					minPrice = data.Prices[i].LowEstimate
					minDuration = data.Prices[i].Duration
					minDistance = data.Prices[i].Distance
				}
			}
			curPrice = curPrice + minPrice
			curDuration = curDuration + minDuration
			curDistance = curDistance + minDistance
		}

		routePrice = AppendInt(routePrice, curPrice)
		routeDuration = AppendInt(routeDuration, curDuration)
		routeDistance = AppendFloat(routeDistance, curDistance)

		fmt.Println(curPrice)
		fmt.Println(curDuration)
		fmt.Println(curDistance)
		curPrice = 0
		curDuration = 0
		curDistance = 0.0
		res = AppendBsonId(res, result.([]bson.ObjectId))
		fmt.Println(pmtTarget.Index(), result.([]bson.ObjectId))
	}
	index := 0
	curPrice = 1000
	for i := 0; i < len(routePrice); i++ {
		if curPrice > routePrice[i] {
			curPrice = routePrice[i]
			index = i
		}
	}
	fmt.Println("The best route of this trip is: ")
	fmt.Println(res[index])
	fmt.Println(routePrice[index])
	fmt.Println(routeDuration[index])
	fmt.Println(routeDistance[index])

	tripResponse.Best_route_location_id = res[index]
	tripResponse.Total_distance = routeDistance[index]
	tripResponse.Total_uber_costs = routePrice[index]
	tripResponse.Total_uber_duration = routeDuration[index]
	tripData.Id = tripResponse.Id
	tripData.Index = 0
	tripData.Status = tripResponse.Status
	tripData.Starting_from_location_id = tripResponse.Starting_from_location_id
	tripData.Best_route_location_id = res[index]
	tripData.Total_uber_costs = tripResponse.Total_uber_costs
	tripData.Total_uber_duration = tripResponse.Total_uber_duration
	tripData.Total_distance = tripResponse.Total_distance
}

func AppendBsonId(slice [][]bson.ObjectId, data ...[]bson.ObjectId) [][]bson.ObjectId {
	m := len(slice)
	n := m + 1
	if n > cap(slice) {
		newSlice := make([][]bson.ObjectId, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

// Used in bestRoute
func AppendInt(slice []int, data ...int) []int {
	m := len(slice)
	n := m + 1
	if n > cap(slice) {
		newSlice := make([]int, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

// Used in bestRoute
func AppendFloat(slice []float64, data ...float64) []float64 {
	m := len(slice)
	n := m + 1
	if n > cap(slice) {
		newSlice := make([]float64, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func uberAPI(carRes *structure.CarResponse, tripData structure.DataStorage, startLat float64, startLng float64, endLat float64, endLng float64) {
	minPrice := 0
	// Server token should be kept safe from others
	serverToken := "PLEASE PUT YOUR OWN SERVER TOKEN HERE"
	urlLeft := "https://api.uber.com/v1/estimates/price?"
	urlRight := "start_latitude=" + strconv.FormatFloat(startLat, 'f', -1, 64) + "&start_longitude=" + strconv.FormatFloat(startLng, 'f', -1, 64) + "&end_latitude=" + strconv.FormatFloat(endLat, 'f', -1, 64) + "&end_longitude=" + strconv.FormatFloat(endLng, 'f', -1, 64) + "&server_token=" + serverToken
	urlFormat := urlLeft + urlRight
	var userrequest structure.UserRequest

	getPrices, err := http.Get(urlFormat)
	if err != nil {
		fmt.Println("Get Prices Error", err)
		panic(err)
	}

	var data structure.UberAPIResponse
	index := 0

	json.NewDecoder(getPrices.Body).Decode(&data)

	minPrice = data.Prices[0].LowEstimate
	for i := 0; i < len(data.Prices); i++ {
		if minPrice > data.Prices[i].LowEstimate {
			minPrice = data.Prices[i].LowEstimate
			index = i
		}
		userrequest.Product_id = data.Prices[index].ProductID
	}

	urlPath := "https://sandbox-api.uber.com/v1/requests"
	userrequest.Start_latitude = startLat
	userrequest.Start_longitude = startLng
	userrequest.End_latitude = endLat
	userrequest.End_longitude = endLng
	// accessToken cannot be told to others, therefore I leave this line empty
	accessToken := "PLEASE PUT YOUR OWN ACCESSTOKEN HERE"
	requestbody, _ := json.Marshal(userrequest)
	client := &http.Client{}
	req, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(requestbody))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("QueryInfo: http.Get", err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	uberRes := structure.UberResponse{}
	json.Unmarshal(body, &uberRes)

	fmt.Println(uberRes)

	carRes.Uber_wait_time_eta = uberRes.Eta
}

// Generate ID for every different trip
func getID() int {
	if id == 0 {
		for id == 0 {
			id = rand.Intn(10000)
		}
	} else {
		id = id + 1
	}
	return id
}
