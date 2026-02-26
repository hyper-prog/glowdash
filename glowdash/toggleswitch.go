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
	"strings"
	"time"

	"github.com/hyper-prog/smartyaml"
)

type PanelToggleSwitch struct {
	PanelHwDevBased

	title2    string
	thumbImg2 string
	badge     string
	badge2    string

	customquerycode string
	customsetcode   string
}

func NewPanelToggleSwitch() *PanelToggleSwitch {
	return &PanelToggleSwitch{
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
		"", "", "", "", "", "",
	}
}

func (p *PanelToggleSwitch) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	p.title2 = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/TitleAlt", indexInConfig), p.title)
	p.thumbImg2 = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/ThumbnailAlt", indexInConfig), p.thumbImg)
	p.badge = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/Badge", indexInConfig), "")
	p.badge2 = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/BadgeAlt", indexInConfig), p.badge)

	p.customquerycode = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/CustomQueryCode", indexInConfig), "")
	p.customsetcode = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/CustomSetCode", indexInConfig), "")

	if p.deviceType == "Shelly" {
		p.deviceIp = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/DeviceIp", indexInConfig), "")
		p.inDeviceId = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/InDeviceId", indexInConfig), 0)
	}

}

func (p *PanelToggleSwitch) PanelHtml(withContainer bool) string {

	BadgeText := getBadgeHtml(p.badge)
	BadgeAltText := getBadgeHtml(p.badge2)

	rawhtml := `
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
			<div id="tglshows-{{.Id}}-tmw" class="circle-avatar-wrapper widget-avatar">
				<div id="tglshows-{{.Id}}-tpp" class="circle-avatar large" role="presentation"> <!-- circle-avatar-flip -->
					<div class="image circle-avatar-face front" style="background-image: url('/user/{{.ThumbImgFront}}')"></div>
					<div class="image circle-avatar-face back" style="background-image: url('/user/{{.ThumbImgBack}}')"></div>
				</div>
				___BADGEOVERLAYTEXTBYSTATE___
			</div>
			<div class="title-container mt-s">
				<p class="title text-bold body-small-styles">{{.TitleByState}}</p>
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
			<button id="b-{{.Id}}-switch"
					data-tgshid="tglshows-{{.Id}}"
			        class="align-self-center device-button primary medium jsaction tglswbtn inactive {{if .NoValidInfo}}noinfo{{end}}">
				<span class="device-action-border">
					<span class="device-action">
						<span class="text-primary icon-grid icon-grid-s">
							<i class="fa fa-change-toggle"></i>
						</span>
					</span>
				</span>
			</button>
		</div>
	</div>`

	titleByState := ""
	thumbImgFront := ""
	thumbImgBack := ""

	rawhtml = strings.ReplaceAll(rawhtml, "___BADGEOVERLAYTEXT1___", BadgeText)
	rawhtml = strings.ReplaceAll(rawhtml, "___BADGEOVERLAYTEXT2___", BadgeAltText)

	if p.state == 0 {
		titleByState = p.title
		rawhtml = strings.ReplaceAll(rawhtml, "___BADGEOVERLAYTEXTBYSTATE___", BadgeText)
		thumbImgFront = p.thumbImg
		thumbImgBack = p.thumbImg2
	} else {
		titleByState = p.title2
		rawhtml = strings.ReplaceAll(rawhtml, "___BADGEOVERLAYTEXTBYSTATE___", BadgeAltText)
		thumbImgFront = p.thumbImg2
		thumbImgBack = p.thumbImg
	}

	templ, _ := template.New("PcT").Parse(rawhtml)

	pass := struct {
		Title         string
		TitleByState  string
		Id            string
		ThumbImg      string
		ThumbImg2     string
		ThumbImgFront string
		ThumbImgBack  string
		State         int
		InputState    int
		IpAddress     string
		HasPowerInfo  bool
		HasValidInfo  bool
		NoValidInfo   bool
		Watt          string
		Volt          string
	}{
		Title:         p.title,
		TitleByState:  titleByState,
		Id:            p.idStr,
		ThumbImg:      p.thumbImg,
		ThumbImg2:     p.thumbImg2,
		State:         p.state,
		InputState:    p.inputState,
		IpAddress:     p.deviceIp,
		HasPowerInfo:  p.hasPoweInfo,
		HasValidInfo:  p.hasValidInfo,
		NoValidInfo:   !p.hasValidInfo,
		Watt:          fmt.Sprintf("%.1f", p.watt),
		Volt:          fmt.Sprintf("%.1f", p.volt),
		ThumbImgFront: thumbImgFront,
		ThumbImgBack:  thumbImgBack,
	}

	buffer := bytes.Buffer{}
	templ.Execute(&buffer, pass)

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"widget-card\" tabindex=\"-1\">", p.IdStr()) +
			buffer.String() + "</div>"
	}

	return buffer.String()
}

func (p *PanelToggleSwitch) IsActionIdMatch(aId string) bool {
	if "b-"+p.idStr+"-switch" == aId {
		return true
	}
	if "b-"+p.idStr+"-update" == aId {
		return true
	}
	return false
}

func (p *PanelToggleSwitch) DoAction(actionName string, parameters map[string]string) (string, []string, bool) {
	var stateChanged bool = false
	var updatedIds []string = []string{}

	if actionName == "switch" && p.customsetcode != "" {
		code, ok := ProgramLibrary[p.customsetcode]
		if ok {
			relatedPanels := []string{}
			initVariables := p.ExposeVariables()
			initVariables["ToggleSwitchPanel.Title"] = p.title
			initVariables["ToggleSwitchPanel.Id"] = p.idStr
			initVariables["ToggleSwitchPanel.DeviceType"] = p.deviceType
			initVariables["ToggleSwitchPanel.ActionName"] = actionName
			initVariables["ReqiredStateText"] = initVariables["Panel.TextualOppositeState"]
			results := ExecuteCommands(code, initVariables, &relatedPanels)
			if DebugLevel >= 2 {
				fmt.Printf("Custom set code \"%s\" executed for panel %s, result: %s\n", p.customsetcode, p.title, results["Return"])
			}
			if results["Return"] == "error" {
				p.InvalidateInfo()
			}
			stateChanged = true
			time.Sleep(time.Millisecond * 200)
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
		return "ok", updatedIds, stateChanged
	}

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

func (p *PanelToggleSwitch) DoActionFromScheduler(actionName string) []string {

	if (actionName == "on" || actionName == "off") && p.customsetcode != "" {
		code, ok := ProgramLibrary[p.customsetcode]
		if ok {
			relatedPanels := []string{}
			initVariables := p.ExposeVariables()
			initVariables["ToggleSwitchPanel.Title"] = p.title
			initVariables["ToggleSwitchPanel.Id"] = p.idStr
			initVariables["ToggleSwitchPanel.DeviceType"] = p.deviceType
			initVariables["ToggleSwitchPanel.ActionName"] = actionName
			if actionName == "on" {
				initVariables["ReqiredStateText"] = "true"
			}
			if actionName == "off" {
				initVariables["ReqiredStateText"] = "false"
			}
			results := ExecuteCommands(code, initVariables, &relatedPanels)
			if results["Return"] == "error" {
				p.InvalidateInfo()
			}
			time.Sleep(time.Millisecond * 200)
		}
		return p.QueryDevice()
	}

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

func (p *PanelToggleSwitch) QueryDevice() []string {
	var updatedIds []string = []string{}

	if p.customquerycode != "" {
		code, ok := ProgramLibrary[p.customquerycode]
		if ok {
			initVariables := p.ExposeVariables()
			initVariables["ToggleSwitchPanel.Title"] = p.title
			initVariables["ToggleSwitchPanel.Id"] = p.idStr
			initVariables["ToggleSwitchPanel.DeviceType"] = p.deviceType
			initVariables["ToggleSwitchPanel.ActionName"] = "update"

			resInt := 0
			results := ExecuteCommands(code, initVariables, &updatedIds)
			if DebugLevel >= 2 {
				fmt.Printf("Custom query code \"%s\" executed for panel %s, result: %s\n", p.customquerycode, p.title, results["Return"])
			}
			if results["Return"] == "error" {
				p.InvalidateInfo()
				return []string{p.idStr}
			}
			if results["Return"] == "true" {
				resInt = 1
			}
			updatedIds = append(updatedIds, p.RefreshHwStatesInRequiredPanelsSwitch(resInt, 0, false, 0.0, 0.0)...)
		}
		return updatedIds
	}

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

		updatedIds = append(updatedIds, p.RefreshHwStatesInRequiredPanelsSwitch(state, inputstate, powerMeasured, apower, voltage)...)
	}

	return updatedIds
}

func (p *PanelToggleSwitch) RefreshHwStatesInRequiredPanelsSwitch(State int, InputState int, PowMet bool, Watt float64, Volt float64) []string {
	var updatedIds []string = []string{}

	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if Panels[i].PanelType() == Switch {
			ps1, ok1 := Panels[i].(*PanelSwitch)
			if ok1 {
				rId := ps1.RefreshHwStateIfMatchSwitch(p.panelType, p.deviceIp, p.inDeviceId, "", State, InputState, PowMet, Watt, Volt)
				if rId != "" {
					updatedIds = append(updatedIds, rId)
				}
			}
			ps2, ok2 := Panels[i].(*PanelToggleSwitch)
			if ok2 {
				rId := ps2.RefreshHwStateIfMatchSwitch(p.panelType, p.deviceIp, p.inDeviceId, "", State, InputState, PowMet, Watt, Volt)
				if rId != "" {
					updatedIds = append(updatedIds, rId)
				}
			}
		}
	}
	return updatedIds
}

func (p *PanelToggleSwitch) RefreshHwStateIfMatchSwitch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int, PowMet bool, Watt float64, Volt float64) string {
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

func (p *PanelToggleSwitch) ExposeVariables() map[string]string {

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
