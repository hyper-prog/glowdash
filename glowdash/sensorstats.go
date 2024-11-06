/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hyper-prog/smartyaml"
)

type PageSensorStats struct {
	PageBase

	hasValidInfo bool
	hwDeviceIp   string
	hwDevicePort int

	sensors []SensorData
}

func NewPageSensorStats() *PageSensorStats {
	return &PageSensorStats{
		PageBase{
			idStr:      "",
			pageType:   SensorStats,
			title:      "",
			deviceType: "",
			index:      0,
		},
		false, "", 0, []SensorData{},
	}
}

func (p *PageSensorStats) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	if p.deviceType == "smtherm" {
		p.hwDeviceIp = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Pages/[%d]/DeviceIp", indexInConfig), "")
		p.hwDevicePort = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Pages/[%d]/DeviceTcpPort", indexInConfig), 5017)

		if sy.NodeExists(fmt.Sprintf("/GlowDash/Pages/[%d]/Sensors", indexInConfig)) {
			sdefs, _ := sy.GetArrayByPath(fmt.Sprintf("/GlowDash/Pages/[%d]/Sensors", indexInConfig))
			sdl := len(sdefs)
			for i := 0; i < sdl; i++ {
				name := sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Pages/[%d]/Sensors/[%d]/Name", indexInConfig, i), "")
				codename := sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Pages/[%d]/Sensors/[%d]/Code", indexInConfig, i), "")
				if len(name) > 0 && len(codename) > 0 {
					s := SensorData{}
					s.codename = codename
					s.name = name
					p.sensors = append(p.sensors, s)
				}
			}
		}
	}
}

func (p PageSensorStats) PageHtml_smtherm() string {
	html := "<table border=\"1\" class=\"stattable\">"
	html += "<tr><th>Name</th> <th>Last read</th> <th>Last success</th> <th>Temp</th> <th>Hum</th> <th>Succ Read</th> <th>Crc Error</th> <th>Insense data</th></tr>"
	for i := 0; i < len(p.sensors); i++ {

		j := execJsonTcpQuery(p.hwDeviceIp, p.hwDevicePort, fmt.Sprintf("cmd:qstat;sn:%s;", p.sensors[i].codename))
		if j.Success {
			html += "<tr>"
			html += "<td>" + p.sensors[i].name + "</td>"
			html += "<td>" + fmt.Sprintf("%ds ago", int(j.SmartJSON.GetFloat64ByPathWithDefault("/lastread", 0.0))) + "</td>"
			ls := j.SmartJSON.GetStringByPathWithDefault("/lastok", "N.A.")
			html += "<td class=\""
			if ls == "yes" {
				html += "csgreen"
			} else {
				html += "csred"
			}
			html += "\">" + ls + "</td>"
			html += "<td>" + fmt.Sprintf("%.1f C", j.SmartJSON.GetFloat64ByPathWithDefault("/temp", 0.0)) + "</td>"
			html += "<td>" + fmt.Sprintf("%.0f %%", j.SmartJSON.GetFloat64ByPathWithDefault("/hum", 0.0)) + "</td>"
			html += "<td>" + fmt.Sprintf("%d", int(j.SmartJSON.GetFloat64ByPathWithDefault("/okread", 0.0))) + "</td>"
			html += "<td>" + fmt.Sprintf("%d", int(j.SmartJSON.GetFloat64ByPathWithDefault("/crcerror", 0.0))) + "</td>"
			html += "<td>" + fmt.Sprintf("%d", int(j.SmartJSON.GetFloat64ByPathWithDefault("/insense", 0.0))) + "</td>"
			html += "</tr>"
		}
	}
	html += "</table>"

	when, what, dura, day := p.CollectHeaterHistory_smtherm()
	html += "<br/>"

	lastday := -1
	alternate := true
	l := min(len(when), min(len(what), min(len(dura), len(day))))
	if l == 0 {
		html += "<p class=\"whitetext\">There is no heater activity log.</p>"
	} else {
		html += "<table class=\"stattable\">"
		html += "<tr><th>Num</th><th>Date/Time</th><th>Action</th><th>Duration</th></tr>"
		na := 1
		nd := 1
		for i := l - 1; i >= 0; i-- {
			if lastday != day[i] {
				alternate = !alternate
				nd = 1
			}

			html += "<tr class=\""
			if alternate {
				html += "altcolor"
			} else {
				html += "normcolor"
			}
			html += "\">"
			html += "<td>" + fmt.Sprintf("%d / %d",na,nd) + "</td>"
			html += "<td>" + when[i] + "</td>"
			html += "<td class=\""
			if what[i] == "Start heating" {
				html += "csorange"
				nd++
			}
			if what[i] == "Stop heating" {
				html += "csdeepblue"
			}
			html += "\">" + what[i] + "</td>"
			html += "<td class=\""
			if dura[i] != "" {
				html += "csred"
			}
			html += "\">" + dura[i] + "</td>"
			html += "</tr>"
			lastday = day[i]
			na++
		}
		html += "</table>"
	}

	return html
}

func (p PageSensorStats) CollectHeaterHistory_smtherm() ([]string, []string, []string, []int) {
	var when []string = []string{}
	var what []string = []string{}
	var dura []string = []string{}
	var day []int = []int{}

	var tm time.Time
	var lastHasStart bool = false
	var lastStart time.Time

	j := execJsonTcpQuery(p.hwDeviceIp, p.hwDevicePort, "cmd:qhshis;")
	if j.Success {
		arr, _ := j.SmartJSON.GetArrayByPath("$.hswhist")
		alen := len(arr)

		for mi := 0; mi < alen; mi++ {
			if subarr, isArray := arr[mi].([]interface{}); isArray {
				f0, isFloat0 := subarr[0].(float64)
				f1, isFloat1 := subarr[1].(float64)

				if isFloat0 && isFloat1 {
					tm = time.Unix(int64(f0), 0)
					when = append(when, tm.Format("2006-01-02 15:04:05"))

					whatstr := "unknown"
					durastr := ""
					if int(f1) == 1 {
						whatstr = "Start heating"
						lastHasStart = true
						lastStart = tm
					}
					if int(f1) == 2 {
						whatstr = "Stop heating"
						if lastHasStart {
							durastr = fmt.Sprintf("%d minute", int(tm.Sub(lastStart).Seconds()/60))
						}
					}
					what = append(what, whatstr)
					dura = append(dura, durastr)
					day = append(day, tm.Day())
				}
			}
		}
	}

	return when, what, dura, day
}

func (p PageSensorStats) PageHtml(withContainer bool, r *http.Request) string {
	html := ""

	if p.deviceType == "smtherm" {
		html += p.PageHtml_smtherm()
	}

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"fullpage-content\" tabindex=\"-1\">", p.IdStr()) +
			html + "</div>"
	}

	return html
}
