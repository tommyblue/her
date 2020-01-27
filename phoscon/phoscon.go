package phoscon

import "github.com/tommyblue/her/her"

import "fmt"

type Server struct {
	token   string
	sensors []Sensor
}

type Sensor struct{}

func New(token string) *Server {
	return &Server{token: token}
}

func (s *Server) Add(sensor her.SensorConf) error {
	fmt.Printf("Adding %v\n", sensor.Value)
	// var subscriptionConfs []her.SubscriptionConf
	// if err := viper.UnmarshalKey("subscriptions", &subscriptionConfs); err != nil {
	// 	return err
	// }

	/*
		package main

	import (
		"encoding/json"
		"fmt"
		"reflect"
		"strings"
	)

	func main() {
		j := `{"state":{"temperature":"likes to perch on rocks","eagle":"bird of prey"},"animals":"none"}`
		s := "state.temperature"
		var result map[string]interface{}
		json.Unmarshal([]byte(j), &result)
		for _, key := range strings.Split(s, ".") {
			if reflect.TypeOf(result[key]) != nil && reflect.TypeOf(result[key]).Kind() == reflect.Map {
				result = result[key].(map[string]interface{})
			} else {
				//fmt.Println("here")
				fmt.Printf("%v\n", result[key])
			}
			//fmt.Println(reflect.TypeOf(result))
		}
		//fmt.Printf("%s", result)
	}
	*/
	return nil
}
