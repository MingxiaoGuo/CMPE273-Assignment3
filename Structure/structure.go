package structure

import "gopkg.in/mgo.v2/bson"

type Request struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type Response struct {
	Id         bson.ObjectId `json:"id" bson:"_id"`
	Name       string        `json:"name" bson:"name"`
	Address    string        `json:"address" bson:"address"`
	City       string        `json:"city" bson:"city"`
	State      string        `json:"state" bson:"state"`
	Zip        string        `json:"zip" bson:"zip"`
	Coordinate struct {
		Lat float64 `json:"lat" bson:"lat"`
		Lng float64 `json:"lng" bson:"lng"`
	} `json:"coordinate" bson:"coordinate"`
}

type MyJsonName struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type UberAPIResponse struct {
	Prices []struct {
		CurrencyCode    string  `json:"currency_code"`
		DisplayName     string  `json:"display_name"`
		Distance        float64 `json:"distance"`
		Duration        int     `json:"duration"`
		Estimate        string  `json:"estimate"`
		HighEstimate    int     `json:"high_estimate"`
		LowEstimate     int     `json:"low_estimate"`
		ProductID       string  `json:"product_id"`
		SurgeMultiplier int     `json:"surge_multiplier"`
	} `json:"prices"`
}

type TripRequest struct {
	Starting_from_location_id bson.ObjectId
	Location_ids              []bson.ObjectId
}

type TripResponse struct {
	Id                        int
	Status                    string
	Starting_from_location_id bson.ObjectId
	Best_route_location_id    []bson.ObjectId
	Total_uber_costs          int
	Total_uber_duration       int
	Total_distance            float64
}

type DataStorage struct {
	Id                        int
	Index                     int
	Status                    string
	Starting_from_location_id bson.ObjectId
	Best_route_location_id    []bson.ObjectId
	Total_uber_costs          int
	Total_uber_duration       int
	Total_distance            float64
}

type UberRequestResponse struct {
	RequestID       string  `json:"request_id"`
	Status          string  `json:"status"`
	Vehicle         string  `json:"vehicle"`
	Driver          string  `json:"driver"`
	Location        string  `json:"location"`
	ETA             int     `json:"eta"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
}

type CarResponse struct {
	Id                           int
	Status                       string
	Starting_from_location_id    bson.ObjectId
	Next_destination_location_id bson.ObjectId
	Best_route_location_id       []bson.ObjectId
	Total_uber_costs             int
	Total_uber_duration          int
	Total_distance               float64
	Uber_wait_time_eta           int
}

type UserRequest struct {
	Product_id      string  `json:"product_id"`
	Start_latitude  float64 `json:"start_latitude"`
	Start_longitude float64 `json:"start_longitude"`
	End_latitude    float64 `json:"end_latitude"`
	End_longitude   float64 `json:"end_longitude"`
}

type UberResponse struct {
	Driver          interface{} `json:"driver"`
	Eta             int         `json:"eta"`
	Location        interface{} `json:"location"`
	RequestID       string      `json:"request_id"`
	Status          string      `json:"status"`
	SurgeMultiplier float64     `json:"surge_multiplier"`
	Vehicle         interface{} `json:"vehicle"`
}
