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
	"strconv"
	"strings"

	"github.com/hyper-prog/smartyaml"
)

type PanelScheduleShortcut struct {
	PanelBase

	scheduleName string
}

func NewPanelScheduleShortcut() *PanelScheduleShortcut {
	return &PanelScheduleShortcut{
		PanelBase{
			idStr:       "",
			panelType:   ScheduleShortcut,
			title:       "",
			eventtitle:  "",
			subPage:     "",
			thumbImg:    "",
			deviceType:  "",
			hide:        false,
			hasPoweInfo: false,
			index:       0,
		},
		"",
	}
}

func (p *PanelScheduleShortcut) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	p.scheduleName = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/ScheduleName", indexInConfig), "--unknown-schedule-name--")
	if p.scheduleName != "--unknown-schedule-name--" {
		p.title = p.scheduleName
	}
}

func (p PanelScheduleShortcut) PanelHtml(withContainer bool) string {
	templ, _ := template.New("PcT").Parse(`
	<div class="badge badge-left" style="max-width: 100%;">
		<div class="label label-s no-radius-bottom-left-diagonal">
			<span class="mr-xs icon-grid icon-grid-xs"><i class="fas fa-sched3"></i></span>
			<div class="label-value-container">
				<p class="text-600 miniature-styles text-nowrap">Schedule</p>
			</div>
		</div>
	</div>

	<div class="main-container {{if .MissingSchedule}}panelnoinfo{{end}}" data-refid="b-{{.Id}}">
		<div class="main-container-top">
			<div class="placeholdersp"></div>

			{{if .ConnShedule}}
			<div class="title-container mt-s">
				<p class="title text-bold body-small-styles">{{.Title}}</p>
			</div>
			{{end}}

			{{if .MissingSchedule}}
			<div class="ctrlline-container mt-s">
				<p class="text-600 title text-bold body-small-styles">No information</p>
			</div>
			{{end}}

			__CLOCKSELECTOR__
			__DAYS__		

		</div>
		<div class="bottom-slot-container d-flex justify-content-center">
			<button id="b-{{.Id}}-toggle" class="align-self-center device-button primary medium jsaction {{if .Disabled}}inactive{{end}} {{if .MissingSchedule}}inactive noinfo{{end}}">
				<span class="device-action-border">
					<span class="device-action">
						<span class="text-primary icon-grid icon-grid-s">
							<i class="fa fa-sched1"></i>
						</span>
					</span>
				</span>
			</button>
		</div>

	</div>`)

	var connectedSchedule bool = false
	s := getScheduleByName(p.scheduleName)
	if s.name == p.scheduleName {
		connectedSchedule = true
	}

	pass := struct {
		Title           string
		Id              string
		ThumbImg        string
		ConnShedule     bool
		MissingSchedule bool
		Enabled         bool
		Disabled        bool
	}{
		Title:           p.title,
		Id:              p.idStr,
		ThumbImg:        p.thumbImg,
		ConnShedule:     connectedSchedule,
		MissingSchedule: !connectedSchedule,
		Enabled:         connectedSchedule && s.enabled,
		Disabled:        connectedSchedule && !s.enabled,
	}

	buffer := bytes.Buffer{}
	templ.Execute(&buffer, pass)

	ostr := buffer.String()
	if connectedSchedule {
		ostr = strings.ReplaceAll(ostr, "__CLOCKSELECTOR__", htmlClockPicker("clksel"+p.IdStr(), s.hour, s.min, false, "jsfiredcs", p.idStr))
		ostr = strings.ReplaceAll(ostr, "__DAYS__", htmlScheduleDays(s, "oneletter", false))
	} else {
		ostr = strings.ReplaceAll(ostr, "__CLOCKSELECTOR__", "")
		ostr = strings.ReplaceAll(ostr, "__DAYS__", "")
	}

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"widget-card\" tabindex=\"-1\">", p.IdStr()) +
			ostr + "</div>"
	}

	return ostr
}

func (p *PanelScheduleShortcut) SetHwDeviceId(id int) {

}

func (p *PanelScheduleShortcut) RefreshHwStateIfMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int) string {
	return p.idStr
}

func (p PanelScheduleShortcut) IsActionIdMatch(aId string) bool {
	if "b-"+p.idStr+"-toggle" == aId {
		return true
	}
	if "b-"+p.idStr+"-updateclock" == aId {
		return true
	}
	if "b-"+p.idStr+"-update" == aId {
		return true
	}
	return false
}

func (p PanelScheduleShortcut) RequiredActionParameters(actionName string) []string {
	if actionName == "updateclock" {
		return []string{"toclock"}
	}
	return []string{}
}

func (p PanelScheduleShortcut) DoAction(actionName string, parameters map[string]string) (string, []string, bool) {
	var stateChanged bool = false
	var updatedIds []string = []string{}
	if actionName == "toggle" {
		idx := getScheduleIndex(p.scheduleName)
		if idx >= 0 {
			s := getScheduleByIndex(idx)
			s.enabled = !s.enabled
			updateSchedule(idx, s)

			stateChanged = true
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
	}
	if actionName == "updateclock" {
		idx := getScheduleIndex(p.scheduleName)
		if idx >= 0 {
			s := getScheduleByIndex(idx)
			if len(parameters["toclock"]) > 3 && parameters["toclock"][0] == 'h' {
				hmtxt := strings.Split(parameters["toclock"][1:], "m")
				if len(hmtxt) == 2 {
					hv, herr := strconv.Atoi(hmtxt[0])
					mv, merr := strconv.Atoi(hmtxt[1])
					if herr == nil && merr == nil {
						s.hour = hv
						s.min = mv
						updateSchedule(idx, s)
						stateChanged = true
					}
				}
			}
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
	}
	if actionName == "update" {
		updatedIds = append(updatedIds, p.QueryDevice()...)
	}
	return "ok", updatedIds, stateChanged
}

func (p *PanelScheduleShortcut) QueryDevice() []string {
	var updatedIds []string = []string{}

	updatedIds = append(updatedIds, p.idStr)

	return updatedIds
}

func (p PanelScheduleShortcut) ExposeVariables() map[string]string {

	var m map[string]string = map[string]string{}

	m["Panel.Id"] = p.idStr
	m["Panel.Title"] = p.title
	m["Panel.SubPage"] = p.subPage
	m["Panel.Index"] = fmt.Sprintf("%d", p.index)
	return m
}
