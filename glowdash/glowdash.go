/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hyper-prog/smartyaml"
)

type PanelTypes int

const (
	Group            PanelTypes = 0
	Switch           PanelTypes = 1
	Shading          PanelTypes = 2
	Action           PanelTypes = 3
	Script           PanelTypes = 4
	Thermostat       PanelTypes = 5
	ThermostatSwitch PanelTypes = 6
	Sensors          PanelTypes = 7
	Launch           PanelTypes = 8
	ScheduleShortcut PanelTypes = 9
	Unknown          PanelTypes = 99
)

type PanelBase struct {
	idStr       string
	panelType   PanelTypes
	title       string
	subPage     string
	thumbImg    string
	deviceType  string
	hide        bool
	hasPoweInfo bool
	index       int
}

type PanelInterface interface {
	Title() string
	PanelType() PanelTypes
	IdStr() string
	Sub() string

	SubTo() string
	LaunchTo() string

	Index() int
	SetIndex(int)

	SetHwDeviceId(int)

	LoadBaseConfig(smartyaml.SmartYAML, int)
	LoadCustomConfig(smartyaml.SmartYAML, int)

	PanelHtml(bool) string
	IsHide() bool
	IsActionIdMatch(string) bool
	GetActionIdFromUrl(full string) string
	RequiredActionParameters(string) []string
	HandleActionEvent(*ActionResponse, string, map[string]string)
	DoAction(string, map[string]string) (string, []string)
	DoActionFromScheduler(string) []string
	QueryDevice() []string
	IsHwMatch(PanelTypes, string, int) bool
	RefreshHwStateIfMatch(PanelTypes, string, int, string, int, int) string
	ExposeVariables() map[string]string
	InvalidateInfo()
}

type PageTypes int

const (
	Settings     PageTypes = 0
	ScheduleEdit PageTypes = 1
	SensorStats  PageTypes = 2
	SensorGraph  PageTypes = 3
	UnknownPage  PageTypes = 99
)

type PageBase struct {
	idStr      string
	pageType   PageTypes
	title      string
	deviceType string
	index      int
}

type PageInterface interface {
	Title() string
	PageType() PageTypes
	IdStr() string

	Index() int
	SetIndex(int)

	LoadBaseConfig(smartyaml.SmartYAML, int)
	LoadCustomConfig(smartyaml.SmartYAML, int)

	PageHtml(bool, *http.Request) string

	IsActionIdMatch(string) bool
	GetActionIdFromUrl(full string) string
	RequiredActionParameters(string) []string
	HandleActionEvent(*ActionResponse, string, map[string]string)
}

var DashboardTitle string = "GlowDash"
var DebugLevel = 0
var configFileName string = ""
var ReadWindInfo bool = false
var WindInfoPollInterval int64 = 3600
var LastWindInfo WindInfo
var StaticFilesDirectory string
var UserFilesDirectory string
var StateConfigDirectory string
var WebServerPort string
var WebUseSSE int = 0
var WebSSEPort int = 8080
var CommUseSSE int = 0
var CommSSEHost string = ""
var CommSSEPort int = 8085
var BackgroudDevQueryNetDialerTimeout time.Duration = time.Duration(1200) * time.Millisecond
var BackgroudDevQueryNetKeepaliveTimeout time.Duration = time.Duration(1200) * time.Millisecond
var AssetVer string = "100"

var Panels []PanelInterface
var Pages []PageInterface
var ProgramLibrary map[string]string = map[string]string{}

func readConfig(yamlfile string) bool {
	Panels = []PanelInterface{}
	Pages = []PageInterface{}
	confYamlData, confYamlFileErr := ioutil.ReadFile(yamlfile)
	if confYamlFileErr != nil {
		log.Printf("Error, cannot read %s file: %s\n", yamlfile, confYamlFileErr.Error())
		return true
	}
	configYAML, confYAMLError := smartyaml.ParseYAML(confYamlData)
	if confYAMLError != nil {
		log.Printf("Error, configuration file has not valid YAML: %s\n", confYAMLError.Error())
		return true
	}
	WebServerPort = configYAML.GetStringByPathWithDefault("/GlowDash/WebServerPort", "80")
	portval, err := strconv.Atoi(WebServerPort)
	if err != nil || portval < 1 || portval > 65534 {
		log.Printf("Not valid server port, fallback to 80\n")
		WebServerPort = "80"
	}

	WebUseSSE = int(configYAML.GetIntegerByPathWithDefault("/GlowDash/WebUseSSE", 0))
	WebSSEPort = int(configYAML.GetIntegerByPathWithDefault("/GlowDash/WebSSEPort", 8080))
	CommUseSSE = int(configYAML.GetIntegerByPathWithDefault("/GlowDash/CommUseSSE", 0))
	CommSSEHost = configYAML.GetStringByPathWithDefault("/GlowDash/CommSSEHost", "127.0.0.1")
	CommSSEPort = int(configYAML.GetIntegerByPathWithDefault("/GlowDash/CommSSEPort", 8085))
	WindInfoPollInterval = int64(configYAML.GetIntegerByPathWithDefault("/GlowDash/WindInfoPollInterval", 3600))
	DashboardTitle = configYAML.GetStringByPathWithDefault("/GlowDash/DashboardTitle", "GlowDash")
	DebugLevel = configYAML.GetIntegerByPathWithDefault("/GlowDash/DebugLevel", 0)
	StaticFilesDirectory = configYAML.GetStringByPathWithDefault("/GlowDash/StaticDirectory", "")
	UserFilesDirectory = configYAML.GetStringByPathWithDefault("/GlowDash/UserDirectory", "")
	StateConfigDirectory = configYAML.GetStringByPathWithDefault("/GlowDash/StateConfigDirectory", ".")
	ReadWindInfo, _ = configYAML.GetBoolByPath("/GlowDash/ReadWindInfo")
	WeatherSource.Provider = configYAML.GetStringByPathWithDefault("/GlowDash/WeatherSource/Provider", "")
	WeatherSource.ApiKey = configYAML.GetStringByPathWithDefault("/GlowDash/WeatherSource/ApiKey", "")
	WeatherSource.Location = configYAML.GetStringByPathWithDefault("/GlowDash/WeatherSource/Location", "")
	AssetVer = configYAML.GetStringByPathWithDefault("/GlowDash/AssetVer", AssetVer)

	BackgroudDevQueryNetDialerTimeout = time.Duration(configYAML.GetIntegerByPathWithDefault("/GlowDash/BackDevDialerTimeout", 1200)) * time.Millisecond
	BackgroudDevQueryNetKeepaliveTimeout = time.Duration(configYAML.GetIntegerByPathWithDefault("/GlowDash/BackDevKeepaliveTimeout", 1200)) * time.Millisecond

	if !strings.HasSuffix(StaticFilesDirectory, "/") {
		StaticFilesDirectory += "/"
	}
	if !strings.HasSuffix(UserFilesDirectory, "/") {
		UserFilesDirectory += "/"
	}

	if configYAML.NodeExists("/GlowDash/CommandLibrary") {
		librarydefs, _ := configYAML.GetArrayByPath("/GlowDash/CommandLibrary")
		ll := len(librarydefs)
		for i := 0; i < ll; i++ {
			name := configYAML.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/CommandLibrary/[%d]/Name", i), "")
			code := configYAML.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/CommandLibrary/[%d]/Code", i), "")
			if configYAML.NodeExists(fmt.Sprintf("/GlowDash/CommandLibrary/[%d]/CodeFile", i)) {
				codeFile := configYAML.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/CommandLibrary/[%d]/CodeFile", i), "")
				codeFileCode, codeFileErr := ioutil.ReadFile(codeFile)
				if codeFileErr != nil {
					log.Printf("Error, cannot read external program file: %s\n", codeFileErr.Error())
				} else {
					code = string(codeFileCode)
				}
			}
			ProgramLibrary[name] = code
		}
	}

	paneldefs, _ := configYAML.GetArrayByPath("/GlowDash/Panels")
	cl := len(paneldefs)
	for i := 0; i < cl; i++ {
		var p PanelInterface = nil

		typ := configYAML.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/PanelType", i), "")

		if typ == "Switch" {
			p = NewPanelSwitch()
		}
		if typ == "Shading" {
			p = NewPanelShading()
		}
		if typ == "Thermostat" {
			p = NewPanelThermostat()
		}
		if typ == "ThermostatSwitch" {
			p = NewPanelThermostatSwitch()
		}
		if typ == "Sensors" {
			p = NewPanelSensors()
		}
		if typ == "Script" {
			p = NewPanelScript()
		}
		if typ == "Action" {
			p = NewPanelAction()
		}
		if typ == "Group" {
			p = NewPanelGroup()
		}
		if typ == "Launch" {
			p = NewPanelLaunch()
		}
		if typ == "ScheduleShortcut" {
			p = NewPanelScheduleShortcut()
		}

		if p != nil {
			p.LoadBaseConfig(configYAML, i)
			p.LoadCustomConfig(configYAML, i)

			Panels = append(Panels, p)
		}
	}

	pagedefs, _ := configYAML.GetArrayByPath("/GlowDash/Pages")
	pl := len(pagedefs)
	for i := 0; i < pl; i++ {
		var p PageInterface = nil

		typ := configYAML.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Pages/[%d]/PageType", i), "")

		if typ == "ScheduleEdit" {
			p = NewPageScheduleEdit()
		}
		if typ == "SensorStats" {
			p = NewPageSensorStats()
		}
		if typ == "SensorGraph" {
			p = NewPageSensorGraph()
		}

		if p != nil {
			p.LoadBaseConfig(configYAML, i)
			p.LoadCustomConfig(configYAML, i)

			Pages = append(Pages, p)
		}
	}

	//Saving indexes to speed up searches.
	pc := len(Panels)
	for i := 0; i < pc; i++ {
		Panels[i].SetIndex(i)
	}

	return false
}

func panelUpdateRequestSSE(panelIds []string) {
	if DebugLevel > 2 {
		fmt.Printf("Send SSE message to refresh panelId: %s \n", strings.Join(panelIds, ","))
	}
	sendSSENotify("panelupd=refreshId(" + strings.Join(panelIds, ",") + ")")
}

func InvalidateInfoOnAllDevice(sub string) {
	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if Panels[i].Sub() == sub {
			if Panels[i].PanelType() == Switch ||
				Panels[i].PanelType() == Shading ||
				Panels[i].PanelType() == Script ||
				Panels[i].PanelType() == Thermostat ||
				Panels[i].PanelType() == ThermostatSwitch ||
				Panels[i].PanelType() == Sensors {
				Panels[i].InvalidateInfo()
			}
		}
	}
}

func QueryAllDevice(sub string) {
	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if Panels[i].Sub() == sub {
			if Panels[i].PanelType() == Switch ||
				Panels[i].PanelType() == Shading ||
				Panels[i].PanelType() == Script ||
				Panels[i].PanelType() == Thermostat ||
				Panels[i].PanelType() == ThermostatSwitch ||
				Panels[i].PanelType() == Sensors {
				Panels[i].QueryDevice()
			}
		}
	}
}

func getDashboard(sub string) string {
	html := htmlStart()
	html += htmlHeaderLine(sub)
	html += htmlPanels(sub)
	html += htmlEnd()
	return html
}

func getPage(w http.ResponseWriter, r *http.Request, sub string) {
	if DebugLevel > 1 {
		fmt.Printf("Req: /%s\n", sub)
	}

	InvalidateInfoOnAllDevice(sub)
	response := getDashboard(sub)
	io.WriteString(w, response)
}

func getCustomPage(w http.ResponseWriter, r *http.Request, page string) {
	if DebugLevel > 1 {
		fmt.Printf("ReqCustomPage: /page/%s\n", page)
	}

	r.ParseForm()
	response := htmlStart()
	response += htmlHeaderLine(page)
	response += htmlCustomPage(page, r)
	response += htmlEnd()
	io.WriteString(w, response)
}

func GetPanelById(id string) PanelInterface {
	for i := 0; i < len(Panels); i++ {
		if id == Panels[i].IdStr() {
			return Panels[i]
		}
	}
	return nil
}

func getAction(w http.ResponseWriter, r *http.Request) {
	if DebugLevel > 0 {
		fmt.Printf("FireACTION-url: %s\n", r.URL.Path)
	}
	aId := r.URL.Path[8:]
	res := newActionResponse()

	r.ParseForm()
	res.setResultString("error")
	panelcnt := len(Panels)
	for i := 0; i < panelcnt; i++ {
		if Panels[i].IsActionIdMatch(aId) {
			rp := Panels[i].RequiredActionParameters(Panels[i].GetActionIdFromUrl(aId))
			ps := map[string]string{}
			for _, pname := range rp {
				ps[pname] = r.Form.Get(pname)
			}
			Panels[i].HandleActionEvent(&res, Panels[i].GetActionIdFromUrl(aId), ps)
		}
	}
	pagecnt := len(Pages)
	for i := 0; i < pagecnt; i++ {
		if Pages[i].IsActionIdMatch(aId) {
			Pages[i].HandleActionEvent(&res, Pages[i].GetActionIdFromUrl(aId), map[string]string{})
		}
	}
	w.Header().Add("Content-Type", "application/json")
	io.WriteString(w, res.getResponseString())
}

func getStatic(w http.ResponseWriter, r *http.Request, stype string) {
	if DebugLevel > 0 {
		fmt.Printf("ReqSTATIC/USER: %s\n", r.URL.Path)
	}

	p := ""
	if stype == "static" {
		p = StaticFilesDirectory + r.URL.Path[8:]
	}
	if stype == "user" {
		p = UserFilesDirectory + r.URL.Path[6:]
	}

	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		http.Error(w, "404 Not Found", 404)
		return
	}

	if strings.HasSuffix(p, ".css") {
		w.Header().Add("Content-Type", "text/css")
	}
	if strings.HasSuffix(p, ".js") {
		w.Header().Add("Content-Type", "text/javascript")
	}

	if stype == "static" || stype == "user" || strings.HasSuffix(p, ".css") || strings.HasSuffix(p, ".js") {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Expires", time.Now().Add(time.Hour*24).Format(http.TimeFormat))
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	}

	response, _ := ioutil.ReadFile(p)
	io.WriteString(w, string(response))
}

type httpRouter struct{}

func (router *httpRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/static/") {
		getStatic(w, r, "static")
		return
	}
	if strings.HasPrefix(r.URL.Path, "/user/") {
		getStatic(w, r, "user")
		return
	}
	if strings.HasPrefix(r.URL.Path, "/action/") {
		getAction(w, r)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/subpage/") {
		getPage(w, r, r.URL.Path[9:])
		return
	}
	if strings.HasPrefix(r.URL.Path, "/page/") {
		getCustomPage(w, r, r.URL.Path[6:])
		return
	}
	if r.URL.Path == "/" {
		getPage(w, r, "")
		return
	}
	http.Error(w, "404 Not Found", 404)
}

func gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	_ = <-quit

	if DebugLevel > 0 {
		fmt.Printf("Saving schedules\n")
	}
	SaveSchedulesIfRequired()
	os.Exit(0)
}

func schedulerRunner() {
	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()
	done := make(chan bool)
	last_hour := -1
	last_min := -1
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			if last_min != t.Minute() || last_hour != t.Hour() {
				last_hour = t.Hour()
				last_min = t.Minute()
				CheckSchedules()
			}
		}
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Error: You must pass a configuration YAML file name as parameter.")
		return
	}

	configFileName = os.Args[1]
	StaticFilesDirectory = ""
	readConfig(configFileName)

	LastWindInfo.RequestTime = time.Date(2000, 1, 1, 8, 00, 00, 100, time.Local)
	mime.AddExtensionType(".css", "text/css")

	ReadSchedulesFromFileDb()

	var myrouter httpRouter
	if DebugLevel > 0 {
		fmt.Printf("Start server\n")
	}
	go gracefulShutdown()
	go schedulerRunner()
	err := http.ListenAndServe(":"+WebServerPort, &myrouter)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server closed\n")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
