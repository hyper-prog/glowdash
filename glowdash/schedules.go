/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hyper-prog/smartjson"
)

type Schedule struct {
	name    string
	enabled bool
	oneshot bool
	lastrun string

	hour int
	min  int

	dayMon bool
	dayTue bool
	dayWed bool
	dayThu bool
	dayFri bool
	daySat bool
	daySun bool

	actionType  string
	actionId    string
	actionParam string
}

var days_oneletter = map[int]string{
	0: "M",
	1: "T",
	2: "W",
	3: "T",
	4: "F",
	5: "S",
	6: "S",
}

var days_short = map[int]string{
	0: "Mon",
	1: "Tue",
	2: "Wed",
	3: "Thu",
	4: "Fri",
	5: "Sat",
	6: "Sun",
}

var subActionDisplayText = map[string]string{
	"on":    "Switch On",
	"off":   "Switch Off",
	"run":   "Run",
	"open":  "Open",
	"close": "Close",
}

var scheduleMutex sync.Mutex
var schedules []Schedule
var schedulesUnsaved bool = false
var schedulesAutosaveState int = 0
var schedulesAutosaveLimit int = 5

func nullSchedule() Schedule {
	return Schedule{"", false, false, "", 0, 0, false, false, false, false, false, false, false, "", "", ""}
}

func countSchedules() int {
	return len(schedules)
}

func getScheduleByIndex(index int) Schedule {
	if index < 0 || index >= len(schedules) {
		return nullSchedule()
	}
	return schedules[index]
}

func getScheduleIndex(name string) int {
	for i := 0; i < len(schedules); i++ {
		if schedules[i].name == name {
			return i
		}
	}
	return -1
}

func getScheduleByName(name string) Schedule {
	return getScheduleByIndex(getScheduleIndex(name))
}

func addSchedule(s Schedule) {
	scheduleMutex.Lock()
	schedules = append(schedules, s)
	schedulesUnsaved = true
	scheduleMutex.Unlock()
}

func addScheduleLowlevel(s Schedule) {
	schedules = append(schedules, s)
}

func updateSchedule(index int, s Schedule) {
	scheduleMutex.Lock()
	if index >= 0 && index < len(schedules) {
		schedules[index] = s
	}
	schedulesUnsaved = true
	scheduleMutex.Unlock()
}

func scheduleMoveUp(index int) {
	scheduleMutex.Lock()
	if index > 0 && index < len(schedules) {
		s := schedules[index-1]
		schedules[index-1] = schedules[index]
		schedules[index] = s
		schedulesUnsaved = true
	}
	scheduleMutex.Unlock()
}

func scheduleMoveDown(index int) {
	scheduleMutex.Lock()
	if index >= 0 && index < len(schedules)-1 {
		s := schedules[index+1]
		schedules[index+1] = schedules[index]
		schedules[index] = s
		schedulesUnsaved = true
	}
	scheduleMutex.Unlock()
}

func SetScheduleOnOffByName(name string, toState bool) {
	scheduleMutex.Lock()
	idx := getScheduleIndex(name)
	if idx >= 0 && idx < len(schedules) {
		schedules[idx].enabled = toState
		schedulesUnsaved = true
	}
	scheduleMutex.Unlock()
}

func removeSchedule(index int) {
	scheduleMutex.Lock()
	removeScheduleInLock(index)
	scheduleMutex.Unlock()
}

func removeScheduleInLock(index int) {
	if index >= 0 && index < len(schedules) {
		schedules = append(schedules[:index], schedules[index+1:]...)
		schedulesUnsaved = true
	}
}

func FireSchedule(index int) {
	for i := 0; i < len(Panels); i++ {
		if Panels[i].IdStr() == schedules[index].actionId {
			if DebugLevel > 0 {
				fmt.Printf("Scheduler execute: Panel(%s) - %s\n", schedules[index].actionId, schedules[index].actionParam)
			}
			refreshIds := Panels[i].DoActionFromScheduler(schedules[index].actionParam)
			if len(refreshIds) > 0 {
				panelUpdateRequestSSE(refreshIds)
			}
		}
	}
}

func CheckScheduleDayEnabled(n time.Time, s Schedule) bool {
	if n.Weekday() == time.Monday && s.dayMon {
		return true
	}
	if n.Weekday() == time.Tuesday && s.dayTue {
		return true
	}
	if n.Weekday() == time.Wednesday && s.dayWed {
		return true
	}
	if n.Weekday() == time.Thursday && s.dayTue {
		return true
	}
	if n.Weekday() == time.Friday && s.dayFri {
		return true
	}
	if n.Weekday() == time.Saturday && s.daySat {
		return true
	}
	if n.Weekday() == time.Sunday && s.daySun {
		return true
	}
	return false
}

func CheckSchedules() {
	current_time := time.Now()

	schedulesAutosaveState++
	if schedulesAutosaveState > (schedulesAutosaveLimit - 1) {
		schedulesAutosaveState = 0
		SaveSchedulesIfRequired()
	}

	scheduleMutex.Lock()
	for i := 0; i < len(schedules); i++ {
		if schedules[i].enabled {
			if schedules[i].hour == current_time.Hour() && schedules[i].min == current_time.Minute() {
				if CheckScheduleDayEnabled(current_time, schedules[i]) {
					FireSchedule(i)
					if schedules[i].oneshot {
						removeScheduleInLock(i)
					} else {
						schedules[i].lastrun = fmt.Sprintf("%d-%d-%d %d:%d", current_time.Year(), current_time.Month(), current_time.Day(),
							current_time.Hour(), current_time.Minute())
					}
				}
			}
		}
	}
	scheduleMutex.Unlock()
}

func schedulesGetJson() string {
	scheduleMutex.Lock()
	o := "{\"schedules\":["
	haspre := false
	for i := 0; i < len(schedules); i++ {
		if schedules[i].oneshot {
			continue
		}

		if haspre {
			o += ","
		}
		o += "{"
		o += "\"name\":\"" + schedules[i].name + "\","
		o += "\"enabled\": " + TrueFalseTextFromBool(schedules[i].enabled) + ","
		o += "\"hour\": " + fmt.Sprintf("%d", schedules[i].hour) + ","
		o += "\"min\": " + fmt.Sprintf("%d", schedules[i].min) + ","

		o += "\"mon\": " + TrueFalseTextFromBool(schedules[i].dayMon) + ","
		o += "\"tue\": " + TrueFalseTextFromBool(schedules[i].dayTue) + ","
		o += "\"wed\": " + TrueFalseTextFromBool(schedules[i].dayWed) + ","
		o += "\"thu\": " + TrueFalseTextFromBool(schedules[i].dayThu) + ","
		o += "\"fri\": " + TrueFalseTextFromBool(schedules[i].dayFri) + ","
		o += "\"sat\": " + TrueFalseTextFromBool(schedules[i].daySat) + ","
		o += "\"sun\": " + TrueFalseTextFromBool(schedules[i].daySun) + ","

		o += "\"at\":\"" + schedules[i].actionType + "\","
		o += "\"ai\":\"" + schedules[i].actionId + "\","
		o += "\"ap\":\"" + schedules[i].actionParam + "\","
		o += "\"lr\":\"" + schedules[i].lastrun + "\""

		o += "}"
		haspre = true
	}
	o += "]}"
	scheduleMutex.Unlock()
	return o
}

func LowerUpperCase(luc bool, str string) string {
	if luc {
		return strings.ToUpper(str)
	}
	return strings.ToLower(str)
}

func schedulesGetDbStr() string {
	o := ""
	scheduleMutex.Lock()
	for i := 0; i < len(schedules); i++ {
		if schedules[i].oneshot {
			continue
		}
		if schedules[i].enabled {
			o += "1;"
		} else {
			o += "0;"
		}
		o += schedules[i].name + ";"
		o += fmt.Sprintf("%d;%d", schedules[i].hour, schedules[i].min) + ";"

		o += LowerUpperCase(schedules[i].dayMon, "m")
		o += LowerUpperCase(schedules[i].dayTue, "t")
		o += LowerUpperCase(schedules[i].dayWed, "w")
		o += LowerUpperCase(schedules[i].dayThu, "t")
		o += LowerUpperCase(schedules[i].dayFri, "f")
		o += LowerUpperCase(schedules[i].daySat, "s")
		o += LowerUpperCase(schedules[i].daySun, "s")
		o += ";"
		o += schedules[i].actionType + ":" + schedules[i].actionId + ":" + schedules[i].actionParam + ";"
		o += schedules[i].lastrun + "\n"
	}
	scheduleMutex.Unlock()
	return o
}

func SaveSchedulesToFileJson() {
	f, err := os.Create(StateConfigDirectory + "/schedules.json")
	if err != nil {
		fmt.Println("Cannot create schedules.json")
		return
	}
	defer f.Close()

	if DebugLevel > 0 {
		fmt.Println("Writing schedules.json")
	}

	_, err = f.Write([]byte(schedulesGetJson()))
	if err != nil {
		fmt.Println("Cannot write schedules.json")
		return
	}
	schedulesUnsaved = false
}

func SaveSchedulesToFileDb() {
	f, err := os.Create(StateConfigDirectory + "/schedules.db")
	if err != nil {
		fmt.Println("Cannot create schedules.db")
		return
	}
	defer f.Close()

	if DebugLevel > 0 {
		fmt.Println("Writing schedules.db")
	}

	_, err = f.Write([]byte(schedulesGetDbStr()))
	if err != nil {
		fmt.Println("Cannot write schedules.db")
		return
	}
	schedulesUnsaved = false
}

func ReadSchedulesFromFileJson() {
	content, err := ioutil.ReadFile(StateConfigDirectory + "/schedules.json")
	if err == nil {
		sj, parserror := smartjson.ParseJSON(content)
		if parserror == nil {
			scheduleMutex.Lock()
			schedules = []Schedule{}
			schedulesarray, _ := sj.GetArrayByPath("/schedules")
			sl := len(schedulesarray)
			for i := 0; i < sl; i++ {
				var s Schedule
				s.name = sj.GetStringByPathWithDefault(fmt.Sprintf("/schedules/[%d]/name", i), "")
				s.enabled = sj.GetBoolByPathWithDefault(fmt.Sprintf("/schedules/[%d]/enabled", i), false)
				s.hour = int(sj.GetFloat64ByPathWithDefault(fmt.Sprintf("/schedules/[%d]/hour", i), 0.0))
				s.min = int(sj.GetFloat64ByPathWithDefault(fmt.Sprintf("/schedules/[%d]/min", i), 0.0))

				s.dayMon = sj.GetBoolByPathWithDefault(fmt.Sprintf("/schedules/[%d]/mon", i), false)
				s.dayTue = sj.GetBoolByPathWithDefault(fmt.Sprintf("/schedules/[%d]/tue", i), false)
				s.dayWed = sj.GetBoolByPathWithDefault(fmt.Sprintf("/schedules/[%d]/wed", i), false)
				s.dayThu = sj.GetBoolByPathWithDefault(fmt.Sprintf("/schedules/[%d]/thu", i), false)
				s.dayFri = sj.GetBoolByPathWithDefault(fmt.Sprintf("/schedules/[%d]/fri", i), false)
				s.daySat = sj.GetBoolByPathWithDefault(fmt.Sprintf("/schedules/[%d]/sat", i), false)
				s.daySun = sj.GetBoolByPathWithDefault(fmt.Sprintf("/schedules/[%d]/sun", i), false)

				s.actionType = sj.GetStringByPathWithDefault(fmt.Sprintf("/schedules/[%d]/at", i), "")
				s.actionId = sj.GetStringByPathWithDefault(fmt.Sprintf("/schedules/[%d]/ai", i), "")
				s.actionParam = sj.GetStringByPathWithDefault(fmt.Sprintf("/schedules/[%d]/ap", i), "")

				s.lastrun = sj.GetStringByPathWithDefault(fmt.Sprintf("/schedules/[%d]/lr", i), "")

				if s.name != "" && s.actionType != "" && s.actionId != "" && s.actionParam != "" {
					addScheduleLowlevel(s)
				}
			}
			scheduleMutex.Unlock()
		}
	}
	schedulesUnsaved = false
}

func ReadSchedulesFromFileDb() {
	content, err := ioutil.ReadFile(StateConfigDirectory + "/schedules.db")
	if err == nil {
		scheduleMutex.Lock()
		lines := strings.Split(string(content), "\n")
		schedules = []Schedule{}
		for _, line := range lines {
			s := Schedule{}
			lineparts := strings.Split(line, ";")
			if len(lineparts) < 6 {
				continue
			}

			if lineparts[0] == "0" {
				s.enabled = false
			}
			if lineparts[0] == "1" {
				s.enabled = true
			}
			if len(lineparts[1]) < 1 {
				continue
			}
			s.name = lineparts[1]

			h, herr := strconv.Atoi(strings.Trim(lineparts[2], " "))
			if herr != nil {
				continue
			}
			s.hour = h
			m, merr := strconv.Atoi(strings.Trim(lineparts[3], " "))
			if merr != nil {
				continue
			}
			s.min = m

			if lineparts[4][0] == 'M' {
				s.dayMon = true
			} else if lineparts[4][0] == 'm' {
				s.dayMon = false
			} else {
				continue
			}

			if lineparts[4][1] == 'T' {
				s.dayTue = true
			} else if lineparts[4][1] == 't' {
				s.dayTue = false
			} else {
				continue
			}

			if lineparts[4][2] == 'W' {
				s.dayWed = true
			} else if lineparts[4][2] == 'w' {
				s.dayWed = false
			} else {
				continue
			}

			if lineparts[4][3] == 'T' {
				s.dayThu = true
			} else if lineparts[4][3] == 't' {
				s.dayThu = false
			} else {
				continue
			}

			if lineparts[4][4] == 'F' {
				s.dayFri = true
			} else if lineparts[4][4] == 'f' {
				s.dayFri = false
			} else {
				continue
			}

			if lineparts[4][5] == 'S' {
				s.daySat = true
			} else if lineparts[4][5] == 's' {
				s.daySat = false
			} else {
				continue
			}

			if lineparts[4][6] == 'S' {
				s.daySun = true
			} else if lineparts[4][6] == 's' {
				s.daySun = false
			} else {
				continue
			}

			actparts := strings.Split(lineparts[5], ":")
			if len(actparts) != 3 {
				continue
			}
			s.actionType = actparts[0]
			s.actionId = actparts[1]
			s.actionParam = actparts[2]

			if len(lineparts) > 5 && lineparts[6] != "" {
				s.lastrun = lineparts[6]
			}

			if s.name != "" && s.actionType != "" && s.actionId != "" && s.actionParam != "" {
				addScheduleLowlevel(s)
			}
		}
		scheduleMutex.Unlock()
	}
	schedulesUnsaved = false
}

func SaveSchedulesIfRequired() {
	if schedulesUnsaved {
		SaveSchedulesToFileDb()
	}
}
