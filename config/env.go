package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

type envClient struct {
	IDs   string
	IPs   string
	Ports string
}

func envParseProxy(input Proxy) Proxy {
	confStruct := reflect.ValueOf(&input).Elem()
	numFields := confStruct.NumField()
	for i := 0; i < numFields; i++ {
		fieldName := confStruct.Type().Field(i).Name
		evName := "PROXY" + "_" + strings.ToUpper(fieldName)

		evVal, evExists := os.LookupEnv(evName)
		if evExists {
			confStruct.FieldByName(fieldName).SetString(evVal)
		}
	}
	return input
}

func envParseClients() (output []Client) {
	envData := envClient{}
	confStruct := reflect.ValueOf(&envData).Elem()
	numFields := confStruct.NumField()
	for i := 0; i < numFields; i++ {
		fieldName := confStruct.Type().Field(i).Name
		evName := "PROXY" + "_" + "CLIENT" + "_" + strings.ToUpper(fieldName)

		evVal, evExists := os.LookupEnv(evName)
		if evExists {
			confStruct.FieldByName(fieldName).SetString(evVal)
		}
	}
	clients := strings.Split(envData.IDs, "|")
	ips := strings.Split(envData.IPs, "|")
	ports := strings.Split(envData.Ports, "|")

	for i := range clients {
		output = append(output, Client{
			ID:      strings.Fields(clients[i])[0],
			Address: strings.Fields(ips[i])[0],
			Ports: func() []int {
				if len(ports) <= i || ports[i] == "" {
					return []int{}
				}
				fields := strings.Split(ports[i], ",")
				ints := make([]int, len(fields))
				for i, s := range fields {
					ints[i], _ = strconv.Atoi(s)
				}
				return ints
			}(),
		})
	}
	return output
}
