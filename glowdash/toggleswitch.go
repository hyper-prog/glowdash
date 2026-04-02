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
}

func NewPanelToggleSwitch() *PanelToggleSwitch {
	return &PanelToggleSwitch{
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
		"", "", "", "",
	}
}

func (p *PanelToggleSwitch) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	p.LoadHwDevConfig(sy, indexInConfig)
	p.InitDeviceManipulator(sy, indexInConfig)
	p.title2 = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/TitleAlt", indexInConfig), p.title)
	p.thumbImg2 = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/ThumbnailAlt", indexInConfig), p.thumbImg)
	p.badge = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/Badge", indexInConfig), "")
	p.badge2 = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/BadgeAlt", indexInConfig), p.badge)
}

func (p *PanelToggleSwitch) PanelHtml(withContainer bool) string {

	BadgeText := getBadgeHtml(p.badge)
	BadgeAltText := getBadgeHtml(p.badge2)

	rawhtml := `
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
		PTypText      string
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
		NoInfoText    string
	}{
		Title:         p.title,
		TitleByState:  titleByState,
		Id:            p.idStr,
		PTypText:      T("Device"),
		ThumbImg:      p.thumbImg,
		ThumbImg2:     p.thumbImg2,
		State:         p.state,
		InputState:    p.inputState,
		IpAddress:     p.deviceIp,
		HasPowerInfo:  p.hasPowerInfo,
		HasValidInfo:  p.hasValidInfo,
		NoValidInfo:   !p.hasValidInfo,
		Watt:          fmt.Sprintf("%.1f", p.watt),
		Volt:          fmt.Sprintf("%.1f", p.volt),
		ThumbImgFront: thumbImgFront,
		ThumbImgBack:  thumbImgBack,
		NoInfoText:    T("No information"),
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

	if actionName == "switch" {
		toState := true
		if p.state == 1 {
			toState = false
		}
		r := p.deviceHandler.SwitchTo(p, toState, "tswaction")
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

func (p *PanelToggleSwitch) DoActionFromScheduler(actionName string) []string {
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

func (p *PanelToggleSwitch) QueryDevice() []string {
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

func (p *PanelToggleSwitch) RefreshHwStatesInRequiredPanelsSwitch(State int, InputState int, PowMet bool, Watt float64, Volt float64) []string {
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

func (p *PanelToggleSwitch) RefreshHwStateIfMatchSwitch(fromPanelType PanelTypes, fromDeviceIp string,
	fromInDeviceId int, fromScriptName string, State int, InputState int,
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

func (p *PanelToggleSwitch) ExposeVariables() map[string]string {

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
