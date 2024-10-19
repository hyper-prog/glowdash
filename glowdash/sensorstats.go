/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
	"net/http"

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
			html += "<td>" + j.SmartJSON.GetStringByPathWithDefault("/lastok", "N.A.") + "</td>"
			html += "<td>" + fmt.Sprintf("%.1f C", j.SmartJSON.GetFloat64ByPathWithDefault("/temp", 0.0)) + "</td>"
			html += "<td>" + fmt.Sprintf("%.0f %%", j.SmartJSON.GetFloat64ByPathWithDefault("/hum", 0.0)) + "</td>"
			html += "<td>" + fmt.Sprintf("%d", int(j.SmartJSON.GetFloat64ByPathWithDefault("/okread", 0.0))) + "</td>"
			html += "<td>" + fmt.Sprintf("%d", int(j.SmartJSON.GetFloat64ByPathWithDefault("/crcerror", 0.0))) + "</td>"
			html += "<td>" + fmt.Sprintf("%d", int(j.SmartJSON.GetFloat64ByPathWithDefault("/insense", 0.0))) + "</td>"
			html += "</tr>"
		}
	}
	html += "</table>"
	return html
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
