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

type PanelScript struct {
	PanelHwDevBased

	scriptName string
}

func NewPanelScript() *PanelScript {
	return &PanelScript{
		PanelHwDevBased{
			PanelBase{
				idStr:       "",
				panelType:   Script,
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
		"",
	}
}

func (p *PanelScript) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	if p.deviceType == "Shelly" {
		p.deviceIp = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/DeviceIp", indexInConfig), "")
		p.inDeviceId = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/InDeviceId", indexInConfig), 0)
		p.scriptName = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/ScriptName", indexInConfig), "")
	}
}

func (p PanelScript) PanelHtml(withContainer bool) string {
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
			{{end}}
			{{if .HasPowerInfo}}
			<div class="ctrlline-container mt-s">
				<p class="text-600 title text-bold body-small-styles">
					<i class="fa fa-bolt"></i> 0 W
					<i class="fa fa-circle-bolt"></i> 230V
				</p>
			</div>
			{{end}}
		</div>

		<div class="bottom-slot-container d-flex justify-content-center">
			<button id="b-{{.Id}}-switch" class="align-self-center device-button primary medium jsaction {{if eq .State 0}}inactive{{end}} {{if .NoValidInfo}}inactive noinfo{{end}}">
				<span class="device-action-border">
					<span class="device-action">
						<span class="text-primary icon-grid icon-grid-s">
							<i class="fa fa-script"></i>
						</span>
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
		IpAddress    string
		HasPowerInfo bool
		HasValidInfo bool
		NoValidInfo  bool
	}{
		Title:        p.title,
		Id:           p.idStr,
		ThumbImg:     p.thumbImg,
		State:        p.state,
		IpAddress:    p.deviceIp,
		HasPowerInfo: p.hasPoweInfo,
		HasValidInfo: p.hasValidInfo,
		NoValidInfo:  !p.hasValidInfo,
	}

	buffer := bytes.Buffer{}
	templ.Execute(&buffer, pass)

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"widget-card\" tabindex=\"-1\">", p.IdStr()) +
			buffer.String() + "</div>"
	}

	return buffer.String()
}

func (p PanelScript) IsActionIdMatch(aId string) bool {
	if "b-"+p.idStr+"-switch" == aId {
		return true
	}
	if "b-"+p.idStr+"-update" == aId {
		return true
	}
	return false
}

func (p PanelScript) DoAction(actionName string, parameters map[string]string) (string, []string, bool) {
	var stateChanged bool = false
	var updatedIds []string = []string{}
	if p.deviceType == "Shelly" && p.deviceIp != "" && p.scriptName != "" && p.inDeviceId > -1 {
		if actionName == "switch" {
			actstr := "Start"
			if p.state == 1 {
				actstr = "Stop"
			}
			execUrl := fmt.Sprintf("http://%s/rpc/Script.%s?id=%d", p.deviceIp, actstr, p.inDeviceId)
			ro := execJsonHttpQuery(execUrl)
			if !ro.Success {
				p.InvalidateInfo()
			}
			stateChanged = true
			time.Sleep(time.Millisecond * 500)
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
		if actionName == "update" {
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
	}
	return "ok", updatedIds, stateChanged
}

func (p PanelScript) DoActionFromScheduler(actionName string) []string {
	if p.deviceType == "Shelly" && p.deviceIp != "" && p.scriptName != "" && p.inDeviceId > -1 {
		if actionName == "start" {
			execUrl := fmt.Sprintf("http://%s/rpc/Script.Start?id=%d", p.deviceIp, p.inDeviceId)
			ro := execJsonHttpQuery(execUrl)
			if !ro.Success {
				p.InvalidateInfo()
			}
			time.Sleep(time.Millisecond * 500)
			return p.QueryDevice()
		}
		if actionName == "stop" {
			execUrl := fmt.Sprintf("http://%s/rpc/Script.Stop?id=%d", p.deviceIp, p.inDeviceId)
			ro := execJsonHttpQuery(execUrl)
			if !ro.Success {
				p.InvalidateInfo()
			}
			time.Sleep(time.Millisecond * 500)
			return p.QueryDevice()
		}
	}
	return []string{}
}

func (p *PanelScript) QueryDevice() []string {
	var updatedIds []string = []string{}

	if p.deviceType == "Shelly" && p.deviceIp != "" && p.scriptName != "" {
		execUrl := fmt.Sprintf("http://%s/rpc/Script.List", p.deviceIp)

		jhq := execJsonHttpQuery(execUrl)
		if !jhq.Success {
			p.InvalidateInfo()
			return []string{p.idStr}
		}
		scriptcount := jhq.SmartJSON.GetCountDescendantsByPath("/scripts")
		for i := 0; i < scriptcount; i++ {
			sn := jhq.SmartJSON.GetStringByPathWithDefault(fmt.Sprintf("/scripts/[%d]/name", i), "")
			if sn == p.scriptName {
				Panels[p.Index()].SetHwDeviceId(int(jhq.SmartJSON.GetFloat64ByPathWithDefault(fmt.Sprintf("/scripts/[%d]/id", i), -1.0)))
				run, _ := jhq.SmartJSON.GetBoolByPath(fmt.Sprintf("/scripts/[%d]/running", i))
				var state int
				if run {
					state = 1
				} else {
					state = 0
				}
				updatedIds = append(updatedIds, p.RefreshHwStatesInRequiredScriptPanels(state)...)
				break
			}
		}
	}
	return updatedIds
}

func (p *PanelScript) RefreshHwStatesInRequiredScriptPanels(State int) []string {
	var updatedIds []string = []string{}

	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if Panels[i].PanelType() == Script {
			ps, ok := Panels[i].(*PanelScript)
			if ok {
				rId := ps.RefreshHwStateIfMatchScriptPanel(p.panelType, p.deviceIp, p.inDeviceId, p.scriptName, State)
				if rId != "" {
					updatedIds = append(updatedIds, rId)
				}
			}
		}
	}
	return updatedIds
}

func (p *PanelScript) RefreshHwStateIfMatchScriptPanel(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int) string {
	if p.panelType == fromPanelType && p.deviceIp == fromDeviceIp && p.scriptName == fromScriptName {
		p.state = State
		p.inputState = 0
		p.hasValidInfo = true
		return p.idStr
	}
	return ""
}

func (p PanelScript) ExposeVariables() map[string]string {

	var m map[string]string = map[string]string{}

	m["Panel.Id"] = p.idStr
	m["Panel.Title"] = p.title
	m["Panel.DeviceType"] = p.deviceType
	m["Panel.SubPage"] = p.subPage
	m["Panel.Index"] = fmt.Sprintf("%d", p.index)

	m["Panel.PowerInfo"] = "false"

	m["Panel.DeviceIp"] = p.deviceIp
	m["Panel.ScriptName"] = p.scriptName
	m["Panel.InDeviceId"] = fmt.Sprintf("%d", p.inDeviceId)
	m["Panel.State"] = fmt.Sprintf("%d", p.state)
	m["Panel.InputState"] = fmt.Sprintf("%d", p.inputState)

	return m
}
