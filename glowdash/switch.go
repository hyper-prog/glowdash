/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2026 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/hyper-prog/smartyaml"
)

type PanelSwitch struct {
	PanelHwDevBased
}

func NewPanelSwitch() *PanelSwitch {
	return &PanelSwitch{
		PanelHwDevBased{
			PanelBase{
				idStr:        "",
				panelType:    Switch,
				title:        "",
				eventtitle:   "",
				subPage:      "",
				thumbImg:     "",
				deviceType:   "",
				hide:         false,
				hasPowerInfo: false,
				index:        0,
			},
			DeviceManipulatorInterface(nil), false, "", 0, 0, 0, 0, 0, 0.0, 0.0,
		},
	}
}

func (p *PanelSwitch) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	p.LoadHwDevConfig(sy, indexInConfig)
	p.InitDeviceManipulator(sy, indexInConfig)
}

func (p PanelSwitch) PanelHtml(withContainer bool) string {
	templ, _ := template.New("PcT").Parse(`
	<div class="badge badge-left" style="max-width: 100%;">
		<div class="label label-s no-radius-bottom-left-diagonal">
			<span class="mr-xs icon-grid icon-grid-xs"><i class="fas fa-microchip"></i></span>
			<div class="label-value-container">
				<p class="text-600 miniature-styles text-nowrap">{{.PTypText}}</p>
			</div>
		</div>
	</div>
	
	<div class="main-container {{if .NoValidInfo}}panelnoinfo{{end}}" data-refid="b-{{.Id}}">
		<div class="main-container-top">
			<div class="circle-avatar-wrapper widget-avatar">
				<div class="circle-avatar large" role="presentation">
					<div class="image" style="background-image: url('/user/{{.ThumbImg}}')"></div>
				</div>
			</div>
			<div class="title-container mt-s">
				<p class="title text-bold body-small-styles">{{.Title}}</p>
			</div>
			{{if .NoValidInfo}}
			<div class="ctrlline-container mt-s">
				<p class="text-600 title text-bold body-small-styles">{{.NoInfoText}}</p>
			</div>
			{{else}}
				{{if .HasPowerInfo}}
				<div class="ctrlline-container mt-s">
					<p class="text-600 title text-bold body-small-styles">
						<i class="fa fa-bolt"></i> {{.Watt}} W
						<i class="fa fa-circle-bolt"></i> {{.Volt}} V
					</p>
				</div>
				{{end}}
			{{end}}
		</div>

		<div class="bottom-slot-container d-flex justify-content-center">
			<button id="b-{{.Id}}-switch" class="align-self-center device-button primary medium jsaction {{if eq .State 0}}inactive{{end}} {{if .NoValidInfo}}noinfo{{end}}">
				<span class="device-action-border">
					<span class="device-action">
						<span class="text-primary icon-grid icon-grid-s">
							<i class="fa fa-power-off"></i>
						</span>
						{{if .HasValidInfo}}
						<span class="indicator {{if eq .InputState 0}}off{{end}}{{if eq .InputState 1}}on{{end}}"></span>
						{{end}}
					</span>
				</span>
			</button>
		</div>
	</div>`)

	pass := struct {
		Title        string
		Id           string
		PTypText     string
		ThumbImg     string
		State        int
		InputState   int
		IpAddress    string
		HasPowerInfo bool
		HasValidInfo bool
		NoValidInfo  bool
		Watt         string
		Volt         string
		NoInfoText   string
	}{
		Title:        p.title,
		Id:           p.idStr,
		PTypText:     T("Device"),
		ThumbImg:     p.thumbImg,
		State:        p.state,
		InputState:   p.inputState,
		IpAddress:    p.deviceIp,
		HasPowerInfo: p.hasPowerInfo,
		HasValidInfo: p.hasValidInfo,
		NoValidInfo:  !p.hasValidInfo,
		Watt:         fmt.Sprintf("%.1f", p.watt),
		Volt:         fmt.Sprintf("%.1f", p.volt),
		NoInfoText:   T("No information"),
	}

	buffer := bytes.Buffer{}
	templ.Execute(&buffer, pass)
	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"widget-card\" tabindex=\"-1\">", p.IdStr()) +
			buffer.String() + "</div>"
	}

	return buffer.String()
}

func (p PanelSwitch) IsActionIdMatch(aId string) bool {
	if "b-"+p.idStr+"-switch" == aId {
		return true
	}
	if "b-"+p.idStr+"-update" == aId {
		return true
	}
	return false
}

func (p *PanelSwitch) DoAction(actionName string, parameters map[string]string) (string, []string, bool) {
	var stateChanged bool = false
	var updatedIds []string = []string{}

	if actionName == "switch" {
		toState := true
		if p.state == 1 {
			toState = false
		}
		r := p.deviceHandler.SwitchTo(p, toState, "swaction")
		if r.ok {
			stateChanged = true
			updatedIds = r.updIds
		}
		time.Sleep(time.Millisecond * 200)
		updatedIds = append(updatedIds, p.QueryDevice()...)
		return "ok", updatedIds, stateChanged
	}

	if actionName == "update" {
		updatedIds = append(updatedIds, p.QueryDevice()...)
		return "ok", updatedIds, stateChanged
	}

	return "ok", updatedIds, stateChanged
}

func (p *PanelSwitch) DoActionFromScheduler(actionName string) []string {
	if actionName == "on" || actionName == "off" {
		toState := false
		if actionName == "on" {
			toState = true
		}
		p.deviceHandler.SwitchTo(p, toState, "swscheduler")
		time.Sleep(time.Millisecond * 200)
		return p.QueryDevice()
	}
	return []string{}
}

func (p *PanelSwitch) QueryDevice() []string {
	var updatedIds []string = []string{}
	queryResult := p.deviceHandler.QuerySwitch(p, "query")
	if !queryResult.ok {
		return []string{p.idStr}
	}

	updatedIds = append(updatedIds,
		p.RefreshHwStatesInRequiredPanelsSwitch(queryResult.state, queryResult.inputstate,
			queryResult.powerMeasured, queryResult.apower, queryResult.voltage)...)

	return updatedIds
}

func (p *PanelSwitch) RefreshHwStatesInRequiredPanelsSwitch(State int, InputState int, PowMet bool, Watt float64, Volt float64) []string {
	var updatedIds []string = []string{}

	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if Panels[i].PanelType() == Switch {
			ps1, ok1 := Panels[i].(*PanelSwitch)
			if ok1 {
				rId := ps1.RefreshHwStateIfMatchSwitch(p.panelType, p.deviceIp, p.inDeviceId, "", State, InputState, PowMet, Watt, Volt, p.idStr)
				if rId != "" {
					updatedIds = append(updatedIds, rId)
				}
			}
			ps2, ok2 := Panels[i].(*PanelToggleSwitch)
			if ok2 {
				rId := ps2.RefreshHwStateIfMatchSwitch(p.panelType, p.deviceIp, p.inDeviceId, "", State, InputState, PowMet, Watt, Volt, p.idStr)
				if rId != "" {
					updatedIds = append(updatedIds, rId)
				}
			}
		}
	}
	return updatedIds
}

func (p *PanelSwitch) RefreshHwStateIfMatchSwitch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int,
	fromScriptName string, State int, InputState int,
	PowMet bool, Watt float64, Volt float64, pId string) string {
	if p.panelType == fromPanelType && p.deviceIp == fromDeviceIp && p.inDeviceId == fromInDeviceId {
		if p.deviceIp == "" && pId != p.idStr {
			return "" // (Probably) independent device without hw info.
		}
		p.state = State
		p.inputState = InputState
		p.hasValidInfo = true
		p.hasPowerInfo = PowMet
		p.watt = Watt
		p.volt = Volt
		return p.idStr
	}
	return ""
}

func (p PanelSwitch) ExposeVariables() map[string]string {

	var m map[string]string = map[string]string{}

	m["Panel.Id"] = p.idStr
	m["Panel.Title"] = p.title
	m["Panel.DeviceType"] = p.deviceType
	m["Panel.SubPage"] = p.subPage
	m["Panel.Index"] = fmt.Sprintf("%d", p.index)

	pwrinfostr := "false"
	if p.hasPowerInfo {
		pwrinfostr = "true"
	}
	m["Panel.PowerInfo"] = pwrinfostr

	m["Panel.DeviceIp"] = p.deviceIp
	m["Panel.TcpPort"] = fmt.Sprintf("%d", p.tcpPort)
	m["Panel.InDeviceId"] = fmt.Sprintf("%d", p.inDeviceId)
	m["Panel.State"] = fmt.Sprintf("%d", p.state)
	m["Panel.InputState"] = fmt.Sprintf("%d", p.inputState)
	m["Panel.Watt"] = fmt.Sprintf("%.2f", p.watt)
	m["Panel.Volt"] = fmt.Sprintf("%.2f", p.volt)
	m["Panel.TextualState"] = ""
	m["Panel.TextualOppositeState"] = ""

	if p.state == 0 {
		m["Panel.TextualState"] = "false"
		m["Panel.TextualOppositeState"] = "true"
	} else if p.state == 1 {
		m["Panel.TextualState"] = "true"
		m["Panel.TextualOppositeState"] = "false"
	}
	return m
}
