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
	"io/ioutil"
	"log"

	"github.com/hyper-prog/smartyaml"
)

type PanelAction struct {
	PanelBase

	Commands      string
	RelatedPanels []string
}

func NewPanelAction() *PanelAction {
	return &PanelAction{
		PanelBase{
			idStr:       "",
			panelType:   Action,
			title:       "",
			eventtitle:  "",
			subPage:     "",
			thumbImg:    "",
			deviceType:  "",
			hide:        false,
			hasPoweInfo: false,
			index:       0,
		},
		"", []string{},
	}
}

func (p *PanelAction) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	p.Commands = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/Commands", indexInConfig), "")
	if sy.NodeExists(fmt.Sprintf("/GlowDash/Panels/[%d]/CommandFile", indexInConfig)) {
		commandFile := sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/CommandFile", indexInConfig), "")
		commandFileProgram, commandFileErr := ioutil.ReadFile(commandFile)
		if commandFileErr != nil {
			log.Printf("Error, cannot read external program file: %s\n", commandFileErr.Error())
		} else {
			p.Commands = string(commandFileProgram)
		}
	}
}

func (p PanelAction) PanelHtml(withContainer bool) string {
	templ, _ := template.New("PcT").Parse(`
	<div class="badge badge-left" style="max-width: 100%;">
		<div class="label label-s no-radius-bottom-left-diagonal">
			<span class="mr-xs icon-grid icon-grid-xs"><i class="fas fa-program2"></i></span>
			<div class="label-value-container">
				<p class="text-600 miniature-styles text-nowrap">Action</p>
			</div>
		</div>
	</div>

	<div class="main-container">
		<div class="main-container-top">
			<div class="circle-avatar-wrapper widget-avatar">
				<div class="circle-avatar large" role="presentation">
					<div class="image" style="background-image: url('/user/{{.ThumbImg}}')"></div>
				</div>
			</div>
			<div class="title-container mt-s">
				<p class="title text-bold body-small-styles">{{.Title}}</p>
			</div>
		</div>

		<div class="bottom-slot-container d-flex justify-content-center">
			<button id="b-{{.Id}}-run" class="align-self-center device-button primary medium jsaction {{if eq .State 0}}inactive{{end}}">
				<span class="device-action-border">
					<span class="device-action">
						<span class="text-primary icon-grid icon-grid-s">
							<i class="fa fa-action"></i>
						</span>
					</span>
				</span>
			</button>
		</div>
	</div>`)

	pass := struct {
		Title    string
		Id       string
		ThumbImg string
		State    int
	}{
		Title:    p.title,
		Id:       p.idStr,
		ThumbImg: p.thumbImg,
		State:    0,
	}

	buffer := bytes.Buffer{}
	templ.Execute(&buffer, pass)

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"widget-card\" tabindex=\"-1\">", p.IdStr()) +
			buffer.String() + "</div>"
	}

	return buffer.String()
}

func (p PanelAction) IsActionIdMatch(aId string) bool {
	if "b-"+p.idStr+"-run" == aId {
		return true
	}
	if "b-"+p.idStr+"-update" == aId {
		return true
	}
	return false
}

func (p PanelAction) DoAction(actionName string, parameters map[string]string) (string, []string, bool) {
	var stateChanged bool = false
	var updatedIds []string = []string{}
	if actionName == "run" {
		p.RelatedPanels = []string{}
		initVariables := map[string]string{}
		initVariables["ActionPanel.RunType"] = "UserAction"
		initVariables["ActionPanel.Title"] = p.title
		initVariables["ActionPanel.Id"] = p.idStr
		initVariables["ActionPanel.DeviceType"] = p.deviceType
		ExecuteCommands(p.Commands, initVariables, &(p.RelatedPanels))
		if len(p.RelatedPanels) > 0 {
			stateChanged = true
		}
		updatedIds = append(updatedIds, p.QueryDevice()...)
	}
	if actionName == "update" {
		updatedIds = append(updatedIds, p.QueryDevice()...)
	}
	return "ok", updatedIds, stateChanged
}

func (p PanelAction) DoActionFromScheduler(actionName string) []string {
	if actionName == "run" {
		p.RelatedPanels = []string{}
		initVariables := map[string]string{}
		initVariables["ActionPanel.RunType"] = "ScheduledTask"
		initVariables["ActionPanel.Title"] = p.title
		initVariables["ActionPanel.Id"] = p.idStr
		initVariables["ActionPanel.DeviceType"] = p.deviceType
		ExecuteCommands(p.Commands, initVariables, &(p.RelatedPanels))
		return p.QueryDevice()
	}
	return []string{}
}

func (p *PanelAction) QueryDevice() []string {
	return append(getUpdatedIdsFromRelatedPanels(p.RelatedPanels), p.idStr)
}

func (p *PanelAction) SetHwDeviceId(id int) {

}

func (p *PanelAction) RefreshHwStateIfMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int) string {
	return p.idStr
}

func (p PanelAction) ExposeVariables() map[string]string {

	var m map[string]string = map[string]string{}

	m["Panel.Id"] = p.idStr
	m["Panel.Title"] = p.title
	m["Panel.DeviceType"] = p.deviceType
	m["Panel.SubPage"] = p.subPage
	m["Panel.Index"] = fmt.Sprintf("%d", p.index)

	m["Panel.PowerInfo"] = "false"
	return m
}
