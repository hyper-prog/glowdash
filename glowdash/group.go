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

	"github.com/hyper-prog/smartyaml"
)

type PanelGroup struct {
	PanelBase

	panelCornerTitle string
	subPageTo        string
}

func NewPanelGroup() *PanelGroup {
	return &PanelGroup{
		PanelBase{
			idStr:       "",
			panelType:   Group,
			title:       "",
			eventtitle:  "",
			subPage:     "",
			thumbImg:    "",
			deviceType:  "",
			hasPoweInfo: false,
			index:       0,
		},
		"Group", "",
	}
}

func (p PanelGroup) SubTo() string {
	return p.subPageTo
}

func (p *PanelGroup) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	p.panelCornerTitle = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/CornerTitle", indexInConfig), "Group")
	p.subPageTo = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/SubPageTo", indexInConfig), "")
}

func (p PanelGroup) PanelHtml(withContainer bool) string {
	templ, _ := template.New("PcT").Parse(`
    	<div class="badge badge-left" style="max-width: 100%;">
    	    <div class="label label-s no-radius-bottom-left-diagonal">
    	        <span class="mr-xs icon-grid icon-grid-xs"><i class="fas fa-door-open"></i></span>
    	        <div class="label-value-container">
    	            <p class="text-600 miniature-styles text-nowrap">{{.CornerTitle}}</p>
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
				<button id="b-{{.Id}}" class="align-self-center device-button primary folder large inactive jsaction">
					<span class="device-action-border">
						<span class="device-action">
							<span class="text-primary icon-grid icon-grid-s">
								<i class="fa fa-folder foldercolor foldericon"></i>
							</span>
						</span>
					</span>
				</button>	
    	    </div>
    	</div>
		`)

	pass := struct {
		Title       string
		Id          string
		ThumbImg    string
		CornerTitle string
	}{
		Title:       p.title,
		Id:          p.idStr,
		ThumbImg:    p.thumbImg,
		CornerTitle: p.panelCornerTitle,
	}

	buffer := bytes.Buffer{}
	templ.Execute(&buffer, pass)

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"widget-card\" tabindex=\"-1\">", p.IdStr()) +
			buffer.String() + "</div>"
	}

	return buffer.String()
}

func (p PanelGroup) DoAction(actionName string, parameters map[string]string) (string, []string, bool) {
	var updatedIds []string = []string{}

	updatedIds = append(updatedIds, p.idStr)

	return "ok", updatedIds, false
}

func (p *PanelGroup) QueryDevice() []string {
	var updatedIds []string = []string{}

	return updatedIds
}

func (p *PanelGroup) SetHwDeviceId(id int) {

}

func (p *PanelGroup) RefreshHwStateIfMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int) string {
	return p.idStr
}

func (p PanelGroup) ExposeVariables() map[string]string {

	var m map[string]string = map[string]string{}

	m["Panel.Id"] = p.idStr
	m["Panel.Title"] = p.title
	m["Panel.DeviceType"] = p.deviceType
	m["Panel.SubPage"] = p.subPage
	m["Panel.Index"] = fmt.Sprintf("%d", p.index)

	m["Panel.PowerInfo"] = "false"

	m["Panel.PanelCornerTitle"] = p.panelCornerTitle
	m["Panel.SubPageTo"] = p.subPageTo

	return m
}
