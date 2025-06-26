/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2025 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/hyper-prog/smartjson"
)

type WeatherSourceType struct {
	Provider string
	ApiKey   string
	Location string
}

type WindInfo struct {
	RequestTime time.Time
	Windspeed   float64
	GustSpeed   float64
}

type JsonHttpQuery struct {
	Success      bool
	ErrorMessage string
	QueryUrl     string
	SmartJSON    smartjson.SmartJSON
}

type JsonTcpQuery struct {
	Success      bool
	ErrorMessage string
	QueryString  string
	SmartJSON    smartjson.SmartJSON
}

func TrueFalseTextFromBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func IfTrue(b bool, text string) string {
	if b {
		return text
	}
	return ""
}

func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func Base64Decode(s string) string {
	se, _ := base64.StdEncoding.DecodeString(s)
	return string(se)
}

var WeatherSource WeatherSourceType = WeatherSourceType{"", "", ""}

func GetWindInfo() WindInfo {
	if WeatherSource.Provider == "weatherapi.com" && WeatherSource.ApiKey != "" && WeatherSource.Location != "" {
		return GetWindInfo_WeatherApiCom()
	}
	return WindInfo{time.Now(), 0.0, 0.0}
}

func GetWindInfo_WeatherApiCom() WindInfo {
	wq := execJsonHttpQuery("http://api.weatherapi.com/v1/current.json?key=" +
		WeatherSource.ApiKey + "&q=" + WeatherSource.Location + "&aqi=no")
	if !wq.Success {
		return WindInfo{time.Now(), 0.0, 0.0}
	}
	wind := wq.SmartJSON.GetFloat64ByPathWithDefault("$.current.wind_kph", 0.0)
	gust := wq.SmartJSON.GetFloat64ByPathWithDefault("$.current.gust_kph", 0.0)
	return WindInfo{time.Now(), wind, gust}
}

func execJsonHttpQuery(url string) JsonHttpQuery {
	start := time.Now()
	jhq := JsonHttpQuery{}

	if DebugLevel > 0 {
		fmt.Printf("CALL -> %s\n", url)
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   BackgroudDevQueryNetDialerTimeout,
				KeepAlive: BackgroudDevQueryNetKeepaliveTimeout,
			}).Dial,
		},
	}

	res, err := client.Get(url)
	if err != nil {
		if DebugLevel > 1 {
			fmt.Printf("Error making http request: %s\n", err)
		}
		jhq.Success = false
		jhq.ErrorMessage = fmt.Sprintf("Error making http request: %s\n", err)
		return jhq
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		if DebugLevel > 1 {
			fmt.Printf("Error reading http result: %s\n", err)
		}
		jhq.Success = false
		jhq.ErrorMessage = fmt.Sprintf("Error reading http result: %s\n", err)
		return jhq
	}

	if DebugLevel > 1 {
		fmt.Printf("RESPONSE -> %s\nElapsed time: %s\n", string(body), time.Since(start))
	}

	sj, error := smartjson.ParseJSON(body)
	if error != nil {
		if DebugLevel > 1 {
			fmt.Printf("Error parsing json result.\n")
		}
		jhq.Success = false
		return jhq
	}
	jhq.SmartJSON = sj
	jhq.Success = true
	jhq.ErrorMessage = ""
	return jhq
}

func execTcpQuery(ip string, port int, sendData string) []byte {
	start := time.Now()

	if DebugLevel > 1 {
		fmt.Printf("CALL -> %s:%d\n", ip, port)
	}

	dialer := &net.Dialer{
		Timeout:   BackgroudDevQueryNetDialerTimeout,
		KeepAlive: BackgroudDevQueryNetKeepaliveTimeout,
	}

	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		if DebugLevel > 0 {
			fmt.Printf("Dial failed: %s", err)
		}
		return []byte{}
	}
	defer conn.Close()
	if DebugLevel > 1 {
		fmt.Printf("SEND -> %s\n", sendData)
	}

	_, err = conn.Write([]byte(sendData))
	if err != nil {
		if DebugLevel > 0 {
			fmt.Printf("Write to server failed: %s", err)
		}
		return []byte{}
	}

	var buf bytes.Buffer
	io.Copy(&buf, conn)

	if DebugLevel > 1 {
		fmt.Printf("RESPONSE -> %s\nElapsed time: %s\n Size: %d\n", buf.String(), time.Since(start), buf.Len())
	}
	return buf.Bytes()
}

func execTcpSend(ip string, port int, sendData string) {
	if DebugLevel > 1 {
		fmt.Printf("CALL -> %s:%d\n", ip, port)
	}

	dialer := &net.Dialer{
		Timeout:   BackgroudDevQueryNetDialerTimeout,
		KeepAlive: BackgroudDevQueryNetKeepaliveTimeout,
	}

	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		if DebugLevel > 0 {
			fmt.Printf("Dial failed: %s", err)
		}
		return
	}
	defer conn.Close()
	if DebugLevel > 1 {
		fmt.Printf("SEND -> %s\n", sendData)
	}

	_, err = conn.Write([]byte(sendData))
	if err != nil {
		if DebugLevel > 0 {
			fmt.Printf("Write to server failed: %s", err)
		}
	}
}

func execJsonTcpQuery(ip string, port int, sendData string) JsonTcpQuery {
	jtq := JsonTcpQuery{}

	jtq.QueryString = sendData

	response := execTcpQuery(ip, port, sendData)
	if len(response) == 0 {
		jtq.Success = false
		jtq.ErrorMessage = "Empty string received"
		return jtq
	}

	sj, error := smartjson.ParseJSON(response)
	if error != nil {
		if DebugLevel > 0 {
			fmt.Printf("Error parsing json result: %s\n", error)
		}
		jtq.Success = false
		jtq.ErrorMessage = "Error parsing received string as json"
		return jtq
	}
	jtq.SmartJSON = sj
	jtq.Success = true
	jtq.ErrorMessage = ""
	return jtq
}

type ActionResponse struct {
	resultStr string
	commands  []string
}

func newActionResponse() ActionResponse {
	return ActionResponse{"unknown", []string{}}
}

func (ar *ActionResponse) setResultString(rs string) *ActionResponse {
	ar.resultStr = rs
	return ar
}

func (ar *ActionResponse) addCommandArg0(cmd string) *ActionResponse {
	ar.commands = append(ar.commands, cmd)
	return ar
}

func (ar *ActionResponse) addCommandArg1(cmd string, arg1 string) *ActionResponse {
	ar.commands = append(ar.commands, cmd+":"+Base64Encode(arg1))
	return ar
}

func (ar *ActionResponse) addCommandArg2(cmd string, arg1 string, arg2 string) *ActionResponse {
	ar.commands = append(ar.commands, cmd+":"+Base64Encode(arg1)+":"+Base64Encode(arg2))
	return ar
}

func (ar *ActionResponse) getResponseString() string {
	cmdpart := ""
	for i := 0; i < len(ar.commands); i++ {
		if i > 0 {
			cmdpart += ","
		}
		cmdpart += "\"" + ar.commands[i] + "\""
	}
	return fmt.Sprintf("{\"result\":\"%s\",\"cmds\":[%s]}\n", ar.resultStr, cmdpart)
}

func sendSSENotify(message string) {
	if CommUseSSE == 0 {
		return
	}
	execTcpSend(CommSSEHost, CommSSEPort, message)
}

func min(a int,b int ) int {
	if a < b {
		return a
	}
	return b
}
