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
				idStr:       "",
				panelType:   Switch,
				title:       "",
				eventtitle:  "",
				subPage:     "",
				thumbImg:    "",
				deviceType:  "",
				hide:        false,
				hasPoweInfo: false,
				index:       0,
			},
			false, "", 0, 0, 0, 0.0, 0.0,
		},
	}
}

func (p *PanelSwitch) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	if p.deviceType == "Shelly" {
		p.deviceIp = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/DeviceIp", indexInConfig), "")
		p.inDeviceId = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/InDeviceId", indexInConfig), 0)
	}

}

func (p PanelSwitch) PanelHtml(withContainer bool) string {
	templ, _ := template.New("PcT").Parse(`
	<div class="badge badge-left" style="max-width: 100%;">
		<div class="label label-s no-radius-bottom-left-diagonal">
			<span class="mr-xs icon-grid icon-grid-xs"><i class="fas fa-microchip"></i></span>
			<div class="label-value-container">
				<p class="text-600 miniature-styles text-nowrap">Device</p>
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
				<p class="text-600 title text-bold body-small-styles">No information</p>
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
		ThumbImg     string
		State        int
		InputState   int
		IpAddress    string
		HasPowerInfo bool
		HasValidInfo bool
		NoValidInfo  bool
		Watt         string
		Volt         string
	}{
		Title:        p.title,
		Id:           p.idStr,
		ThumbImg:     p.thumbImg,
		State:        p.state,
		InputState:   p.inputState,
		IpAddress:    p.deviceIp,
		HasPowerInfo: p.hasPoweInfo,
		HasValidInfo: p.hasValidInfo,
		NoValidInfo:  !p.hasValidInfo,
		Watt:         fmt.Sprintf("%.1f", p.watt),
		Volt:         fmt.Sprintf("%.1f", p.volt),
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

func (p PanelSwitch) DoAction(actionName string, parameters map[string]string) (string, []string, bool) {
	var stateChanged bool = false
	var updatedIds []string = []string{}
	if p.deviceType == "Shelly" && p.deviceIp != "" {
		if actionName == "switch" {
			tostr := "true"
			if p.state == 1 {
				tostr = "false"
			}
			execUrl := fmt.Sprintf("http://%s/rpc/Switch.Set?id=%d&on=%s", p.deviceIp, p.inDeviceId, tostr)
			ro := execJsonHttpQuery(execUrl)
			if !ro.Success {
				p.InvalidateInfo()
			}
			stateChanged = true
			time.Sleep(time.Millisecond * 200)
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
		if actionName == "update" {
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}

	}
	return "ok", updatedIds, stateChanged
}

func (p PanelSwitch) DoActionFromScheduler(actionName string) []string {
	if p.deviceType == "Shelly" && p.deviceIp != "" {
		if actionName == "on" {
			execUrl := fmt.Sprintf("http://%s/rpc/Switch.Set?id=%d&on=true", p.deviceIp, p.inDeviceId)
			ro := execJsonHttpQuery(execUrl)
			if !ro.Success {
				p.InvalidateInfo()
			}
			time.Sleep(time.Millisecond * 200)
			return p.QueryDevice()
		}
		if actionName == "off" {
			execUrl := fmt.Sprintf("http://%s/rpc/Switch.Set?id=%d&on=false", p.deviceIp, p.inDeviceId)
			ro := execJsonHttpQuery(execUrl)
			if !ro.Success {
				p.InvalidateInfo()
			}
			time.Sleep(time.Millisecond * 200)
			return p.QueryDevice()
		}
	}
	return []string{}
}

func (p PanelSwitch) QueryDevice() []string {
	var updatedIds []string = []string{}

	if p.deviceType == "Shelly" && p.deviceIp != "" {
		execUrl := fmt.Sprintf("http://%s/rpc/Switch.GetStatus?id=%d", p.deviceIp, p.inDeviceId)
		jhq := execJsonHttpQuery(execUrl)
		if !jhq.Success {
			p.InvalidateInfo()
			return []string{p.idStr}
		}
		bstate := jhq.SmartJSON.GetBoolByPathWithDefault("/output", false)

		var state int
		var inputstate int
		var powerMeasured bool = false
		var apower float64 = 0.0
		var voltage float64 = 0.0
		if bstate {
			state = 1
		} else {
			state = 0
		}

		if jhq.SmartJSON.NodeExists("/apower") && jhq.SmartJSON.NodeExists("/voltage") {
			str1 := ""
			str2 := ""
			apower, str1 = jhq.SmartJSON.GetFloat64ByPath("/apower")
			voltage, str2 = jhq.SmartJSON.GetFloat64ByPath("/voltage")
			if str1 == "float64" && str2 == "float64" && apower >= 0.0 && voltage >= 0.0 {
				powerMeasured = true
			}
		}

		execUrl = fmt.Sprintf("http://%s/rpc/Input.GetStatus?id=%d", p.deviceIp, p.inDeviceId)
		jhq2 := execJsonHttpQuery(execUrl)
		if !jhq2.Success {
			p.InvalidateInfo()
			return []string{p.idStr}
		}
		istate := jhq2.SmartJSON.GetBoolByPathWithDefault("/state", false)
		if istate {
			inputstate = 1
		} else {
			inputstate = 0
		}
		updatedIds = append(updatedIds, p.RefreshHwStatesInRequiredPanelsSwitch(state, inputstate, powerMeasured, apower, voltage)...)
	}

	return updatedIds
}

func (p *PanelSwitch) RefreshHwStatesInRequiredPanelsSwitch(State int, InputState int, PowMet bool, Watt float64, Volt float64) []string {
	var updatedIds []string = []string{}

	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if Panels[i].PanelType() == Switch {
			ps, ok := Panels[i].(*PanelSwitch)
			if ok {
				rId := ps.RefreshHwStateIfMatchSwitch(p.panelType, p.deviceIp, p.inDeviceId, "", State, InputState, PowMet, Watt, Volt)
				if rId != "" {
					updatedIds = append(updatedIds, rId)
				}
			}
		}
	}
	return updatedIds
}

func (p *PanelSwitch) RefreshHwStateIfMatchSwitch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int, PowMet bool, Watt float64, Volt float64) string {
	if p.panelType == fromPanelType && p.deviceIp == fromDeviceIp && p.inDeviceId == fromInDeviceId {
		p.state = State
		p.inputState = InputState
		p.hasValidInfo = true
		p.hasPoweInfo = PowMet
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
	if p.hasPoweInfo {
		pwrinfostr = "true"
	}
	m["Panel.PowerInfo"] = pwrinfostr

	m["Panel.DeviceIp"] = p.deviceIp
	m["Panel.InDeviceId"] = fmt.Sprintf("%d", p.inDeviceId)
	m["Panel.State"] = fmt.Sprintf("%d", p.state)
	m["Panel.InputState"] = fmt.Sprintf("%d", p.inputState)
	m["Panel.Watt"] = fmt.Sprintf("%.2f", p.watt)
	m["Panel.Volt"] = fmt.Sprintf("%.2f", p.volt)

	return m
}
