/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hyper-prog/smartyaml"
)

type PageSensorGraph struct {
	PageBase

	hasValidInfo bool
	hwDeviceIp   string
	hwDevicePort int

	sensors []SensorData
}

var length_names []string = []string{
	"1 hour",
	"3 hour",
	"6 hour",
	"12 hour",
	"1 day",
	"2 days",
	"3 days",
}
var length_values []int = []int{12, 36, 72, 144, 288, 576, 867}
var offset_names []string = []string{
	"Until now",
	"Until 3 hour ago",
	"Until 6 hour ago",
	"Until 12 hour ago",
	"Until 1 day ago",
	"Until 2 days ago",
	"Until 3 days ago",
}
var offset_values []int = []int{0, 36, 72, 144, 288, 576, 867}

func NewPageSensorGraph() *PageSensorGraph {
	return &PageSensorGraph{
		PageBase{
			idStr:      "",
			pageType:   SensorGraph,
			title:      "",
			deviceType: "",
			index:      0,
		},
		false, "", 0, []SensorData{},
	}
}

func (p *PageSensorGraph) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
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

func (p PageSensorGraph) CollectSensorHistory_smtherm(length int, offset int, temphum string) (string, string, string, string) {
	datablock := ""
	datanames := ""

	dsx := ""
	dsy := ""

	var starttimeunix int64 = 0
	var starttime time.Time
	timemin := 0
	timemax := 0
	thmin := 999
	thmax := -999
	for i := 0; i < len(p.sensors); i++ {
		dsx = ""
		dsy = ""
		j := execJsonTcpQuery(p.hwDeviceIp, p.hwDevicePort, fmt.Sprintf("cmd:qhis;sn:%s;off:%d;len:%d;", p.sensors[i].codename, offset, length))
		if j.Success {
			starttimeunix = int64(j.SmartJSON.GetFloat64ByPathWithDefault("/st", 0))
			starttime = time.Unix(starttimeunix, 0)

			arr, _ := j.SmartJSON.GetArrayByPath("$.d")
			alen := len(arr)

			for mi := 0; mi < alen; mi++ {
				if subarr, isArray := arr[mi].([]interface{}); isArray {
					f0, isFloat0 := subarr[0].(float64)
					var f1 float64
					var isFloat1 bool
					if temphum == "temp" {
						f1, isFloat1 = subarr[1].(float64)
					}
					if temphum == "hum" {
						f1, isFloat1 = subarr[2].(float64)
					}

					if isFloat0 && isFloat1 {
						if mi > 0 {
							dsx += ","
							dsy += ","
						}

						dsx += "'" + starttime.Add(time.Duration(f0)*time.Second).Format("2006-01-02 15:04:05") + "'"

						if temphum == "temp" {
							dsy += fmt.Sprintf("%.1f", f1)
						}
						if temphum == "hum" {
							dsy += fmt.Sprintf("%.0f", f1)
						}

						if mi == 0 {
							timemin = int(f0)
						}
						if mi == alen-1 {
							timemax = int(f0)
						}
						if thmin > int(f1) {
							thmin = int(f1)
						}
						if thmax < int(f1) {
							thmax = int(f1)
						}
					}
				}
			}

			datablock += fmt.Sprintf("var tr%d={x:[%s],y:[%s],name:'%s',type: 'scatter',line: {shape: 'spline'}};\n", i, dsx, dsy, p.sensors[i].name)

			if i > 0 {
				datanames += ","
			}
			datanames += fmt.Sprintf("tr%d", i)
		}
	}

	timemin_str := starttime.Add(time.Duration(timemin) * time.Second).Format("2006-01-02 15:04:05")
	timemax_str := starttime.Add(time.Duration(timemax) * time.Second).Format("2006-01-02 15:04:05")

	return datablock, datanames, fmt.Sprintf("'%d','%d'", timemin_str, timemax_str), fmt.Sprintf("%d,%d", thmin-1, thmax+1)
}

func (p PageSensorGraph) PageHtml(withContainer bool, r *http.Request) string {

	datablock := ""
	datanames := ""
	rangex := "0,1"
	rangey := "0,1"

	length := 72
	offset := 0
	temphum := "temp"

	get_length, err := strconv.Atoi(r.Form.Get("len"))
	if err == nil {
		length = get_length
	}
	get_offset, err := strconv.Atoi(r.Form.Get("offset"))
	if err == nil {
		offset = get_offset
	}
	if r.Form.Get("temphum") == "temp" || r.Form.Get("temphum") == "hum" {
		temphum = r.Form.Get("temphum")
	}

	if p.deviceType == "smtherm" {
		datablock, datanames, rangex, rangey = p.CollectSensorHistory_smtherm(length, offset, temphum)
	}

	ytitle := ""

	html := ""
	html += "<script src=\"/static/plotly.min.js\"></script>"
	html += "<form method=\"get\">"
	//html += "Q len:" + fmt.Sprintf("%d", length) + " off: " + fmt.Sprintf("%d", offset) + " th:" + temphum

	html += "<div class=\"graph-select-control-panel\">"
	html += "<select name=\"temphum\" class=\"custom-select\">"
	html += "<option value=\"temp\" "
	if temphum == "temp" {
		html += "selected"
		ytitle = "Temperature (Celsius)"
	}
	html += ">Temperature</option>"
	html += "<option value=\"hum\" "
	if temphum == "hum" {
		html += "selected"
		ytitle = "Humidity (Percent)"
	}
	html += ">Humidity</option>"
	html += "</select>"

	html += "<select name=\"len\" class=\"custom-select\">"
	for i, n := range length_names {
		html += "<option value=\"" + fmt.Sprintf("%d", length_values[i]) + "\""
		if length_values[i] == length {
			html += "selected"
		}
		html += ">" + n + "</option>"
	}
	html += "</select>"

	html += "<select name=\"offset\" class=\"custom-select\">"
	for i, n := range offset_names {
		html += "<option value=\"" + fmt.Sprintf("%d", offset_values[i]) + "\""
		if offset_values[i] == offset {
			html += "selected"
		}
		html += ">" + n + "</option>"
	}
	html += "</select>"

	html += "<input type=\"submit\" name=\"updategraph\" value=\"Show\" class=\" custom-submit\">"
	//html += "<div class=\"clearboth\"></div>"
	html += "</div>"
	html += "</form>"
	html += "<div class=\"sensorpage justify-content-center\" data-refid=\"\">"
	html += "<div id=\"grafplot\" style=\"width:100%;max-width:1024px;margin:auto;\"></div>"
	html += "<script>\n"

	html += datablock

	html += `var data = [` + datanames + `];
	         const layout = {
	             xaxis: {range: [` + rangex + `], title: "Time",color: "white"},
	             yaxis: {range: [` + rangey + `], title: "` + ytitle + `",color: "white"},
	             title: {
					text: "` + p.title + `",
					font: {
						color: "white"
					}
				 },
				 automargin: true,
				 autosize: true,
	             paper_bgcolor: 'rgba(50, 50, 100, 0.2)',
	             plot_bgcolor: 'rgba(50, 50, 120, 0.2)',
				 legend_font_color: 'white',
				 legend: {
					font: {
						color: "white",
						size: 15
					}
				 }
	         };
			 Plotly.newPlot("grafplot", data, layout);`
	html += "</script>"

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"fullpage-content\" tabindex=\"-1\">", p.IdStr()) +
			html + "</div>"
	}

	return html
}
