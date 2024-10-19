/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/hyper-prog/smartyaml"
)

type SensorData struct {
	name     string
	codename string
	temp     float32
	hum      float32
}

type PanelSensors struct {
	PanelBase

	hasValidInfo bool
	hwDeviceIp   string
	hwDevicePort int
	sensors      []SensorData
}

func NewPanelSensors() *PanelSensors {
	return &PanelSensors{
		PanelBase{
			idStr:       "",
			panelType:   Sensors,
			title:       "",
			subPage:     "",
			thumbImg:    "",
			deviceType:  "",
			hide:        false,
			hasPoweInfo: false,
			index:       0,
		},
		false, "", 0, []SensorData{},
	}
}

func (p *PanelSensors) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	if p.deviceType == "smtherm" {
		p.hwDeviceIp = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/DeviceIp", indexInConfig), "")
		p.hwDevicePort = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/DeviceTcpPort", indexInConfig), 5017)

		if sy.NodeExists(fmt.Sprintf("/GlowDash/Panels/[%d]/Sensors", indexInConfig)) {
			sdefs, _ := sy.GetArrayByPath(fmt.Sprintf("/GlowDash/Panels/[%d]/Sensors", indexInConfig))
			sdl := len(sdefs)
			for i := 0; i < sdl; i++ {
				name := sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/Sensors/[%d]/Name", indexInConfig, i), "")
				codename := sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/Sensors/[%d]/Code", indexInConfig, i), "")
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

func (p PanelSensors) PanelHtml(withContainer bool) string {
	templ, _ := template.New("PcT").Parse(`
	<div class="badge badge-left" style="max-width: 100%;">
		<div class="label label-s no-radius-bottom-left-diagonal">
			<span class="mr-xs icon-grid icon-grid-xs"><i class="fas fa-microchip"></i></span>
			<div class="label-value-container">
				<p class="text-600 miniature-styles text-nowrap">Device</p>
			</div>
		</div>
	</div>
	
	<div class="sensorpanel main-container {{if .NoValidInfo}}panelnoinfo{{end}}" data-refid="b-{{.Id}}">
		<div class="main-container-top">
			{{if .ShowTitle}}
			<div class="title-container mt-s">
				<p class="title text-bold body-small-styles">{{.Title}}</p>
			</div>
			{{end}}

			__SENSOR_BLOCK_PLACEHOLDER__

		</div>
	</div>`)

	in_html := ""
	for _, s := range p.sensors {
		if p.hasValidInfo {
			in_html +=
				"<div class=\"title-container mt-xs\">" +
					"<p class=\"title text-bold body-small-styles\">" + s.name + "</p>" +
					"</div>" +
					"<div class=\"ctrlline-container mt-xxs width90percent\">" +
					"<p class=\"text-600 title text-bold body-small-styles\">" +
					"<i class=\"fa fa-temp\"></i>&nbsp;" + fmt.Sprintf("%.1f", s.temp) + "C" +
					"&nbsp;&nbsp;&nbsp;&nbsp;" +
					"<i class=\"fa fa-hum\"></i>&nbsp;" + fmt.Sprintf("%.0f", s.hum) + "%" +
					"</p>" +
					"</div>"
		} else {
			in_html +=
				"<div class=\"title-container mt-xs\">" +
					"<p class=\"title text-bold body-small-styles\">" + s.name + "</p>" +
					"</div>" +
					"<div class=\"ctrlline-container mt-xxs width90percent\">&nbsp;-&nbsp;&nbsp;&nbsp;&nbsp;-&nbsp;</div>"
		}
	}

	pass := struct {
		Title        string
		ShowTitle    bool
		Id           string
		ThumbImg     string
		HasValidInfo bool
		NoValidInfo  bool
	}{
		Title:        p.title,
		ShowTitle:    false,
		Id:           p.idStr,
		ThumbImg:     p.thumbImg,
		HasValidInfo: p.hasValidInfo,
		NoValidInfo:  !p.hasValidInfo,
	}

	buffer := bytes.Buffer{}
	templ.Execute(&buffer, pass)

	ostr := strings.ReplaceAll(buffer.String(), "__SENSOR_BLOCK_PLACEHOLDER__", in_html)
	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"widget-card\" tabindex=\"-1\">", p.IdStr()) +
			ostr + "</div>"
	}

	return ostr
}

func (p *PanelSensors) SetHwDeviceId(id int) {

}

func (p *PanelSensors) RefreshHwStateIfMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int) string {
	return p.idStr
}

func (p *PanelSensors) InvalidateInfo() {
	p.hasValidInfo = false
}

func (p PanelSensors) IsActionIdMatch(aId string) bool {
	if "b-"+p.idStr == aId {
		return true
	}
	if "b-"+p.idStr+"-update" == aId {
		return true
	}
	return false
}

func (p PanelSensors) DoAction(actionName string,parameters map[string]string) (string, []string) {
	var updatedIds []string = []string{}
	if p.deviceType == "smtherm" && p.hwDeviceIp != "" {
		if actionName == "update" {
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
	}
	return "ok", updatedIds
}

func (p PanelSensors) QueryDevice() []string {
	var updatedIds []string = []string{}
	if p.deviceType == "smtherm" {
		j := execJsonTcpQuery(p.hwDeviceIp, p.hwDevicePort, "cmd:qas;")
		if j.Success {
			var sensors []SensorData = []SensorData{}
			sd_array, _ := j.SmartJSON.GetArrayByPath("/sensors")
			sc := len(sd_array)
			for i := 0; i < sc; i++ {
				name := j.SmartJSON.GetStringByPathWithDefault(fmt.Sprintf("/sensors/[%d]/name", i), "")
				temp := j.SmartJSON.GetFloat64ByPathWithDefault(fmt.Sprintf("/sensors/[%d]/temp", i), -100.0)
				hum := j.SmartJSON.GetFloat64ByPathWithDefault(fmt.Sprintf("/sensors/[%d]/hum", i), -100.0)
				if len(name) > 0 && temp > -100 && hum > -100 {
					sensors = append(sensors, SensorData{"", name, float32(temp), float32(hum)})
				}
			}
			updatedIds = append(updatedIds, p.RefreshHwStatesInRequiredPanelsSensors(sensors)...)
		} else {
			p.InvalidateInfo()
			updatedIds = append(updatedIds, p.idStr)
		}
	}
	return updatedIds
}

func (p *PanelSensors) RefreshHwStatesInRequiredPanelsSensors(sensors []SensorData) []string {
	var updatedIds []string = []string{}

	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if Panels[i].PanelType() == Sensors {
			ps, ok := Panels[i].(*PanelSensors)
			if ok {
				rId := ps.RefreshHwStateIfMatchSensors(p.panelType, p.hwDeviceIp, p.hwDevicePort, sensors)
				if rId != "" {
					updatedIds = append(updatedIds, rId)
				}
			}
		}
	}
	return updatedIds
}

func (p *PanelSensors) RefreshHwStateIfMatchSensors(fromPanelType PanelTypes, fromDeviceIp string, fromDevicePort int, sensors []SensorData) string {
	if p.panelType == fromPanelType && p.hwDeviceIp == fromDeviceIp && p.hwDevicePort == fromDevicePort {

		c := len(p.sensors)
		for _, s := range sensors {
			for i := 0; i < c; i++ {
				if p.sensors[i].codename == s.codename {
					p.sensors[i].temp = s.temp
					p.sensors[i].hum = s.hum
					p.hasValidInfo = true
					break
				}
			}
		}
		return p.idStr
	}

	return ""
}

func (p PanelSensors) ExposeVariables() map[string]string {

	var m map[string]string = map[string]string{}

	m["Panel.Id"] = p.idStr
	m["Panel.Title"] = p.title
	m["Panel.DeviceType"] = p.deviceType
	m["Panel.SubPage"] = p.subPage
	m["Panel.Index"] = fmt.Sprintf("%d", p.index)

	return m
}
