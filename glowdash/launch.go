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

type PanelLaunch struct {
	PanelBase

	panelCornerTitle        string
	pageIdentifier          string
	buttonFontImageCssClass string
}

func NewPanelLaunch() *PanelLaunch {
	return &PanelLaunch{
		PanelBase{
			idStr:       "",
			panelType:   Launch,
			title:       "",
			subPage:     "",
			thumbImg:    "",
			deviceType:  "",
			hasPoweInfo: false,
			index:       0,
		},
		"Page", "", "fa-launch",
	}
}

func (p PanelLaunch) LaunchTo() string {
	return p.pageIdentifier
}

func (p *PanelLaunch) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	p.panelCornerTitle = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/CornerTitle", indexInConfig), "Launch")
	p.buttonFontImageCssClass = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/ButtonFontImageCssClass", indexInConfig), "fa-launch")
	p.pageIdentifier = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/LaunchTo", indexInConfig), "")
}

func (p PanelLaunch) PanelHtml(withContainer bool) string {
	templ, _ := template.New("PcT").Parse(`
    	<div class="badge badge-left" style="max-width: 100%;">
    	    <div class="label label-s no-radius-bottom-left-diagonal">
    	        <span class="mr-xs icon-grid icon-grid-xs"><i class="fas fa-rocket"></i></span>
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
				<button id="b-{{.Id}}" class="align-self-center device-button primary medium inactive jsaction">
					<span class="device-action-border">
						<span class="device-action">
							<span class="text-primary icon-grid icon-grid-s">
								<i class="fa {{.ButtonFontImageCssClass}}"></i>
							</span>
						</span>
					</span>
				</button>	
    	    </div>
    	</div>
		`)

	pass := struct {
		Title                   string
		Id                      string
		ThumbImg                string
		CornerTitle             string
		ButtonFontImageCssClass string
	}{
		Title:                   p.title,
		Id:                      p.idStr,
		ThumbImg:                p.thumbImg,
		CornerTitle:             p.panelCornerTitle,
		ButtonFontImageCssClass: p.buttonFontImageCssClass,
	}

	buffer := bytes.Buffer{}
	templ.Execute(&buffer, pass)

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"widget-card\" tabindex=\"-1\">", p.IdStr()) +
			buffer.String() + "</div>"
	}

	return buffer.String()
}

func (p PanelLaunch) DoAction(actionName string,parameters map[string]string) (string, []string) {
	var updatedIds []string = []string{}

	updatedIds = append(updatedIds, p.idStr)

	return "ok", updatedIds
}

func (p *PanelLaunch) QueryDevice() []string {
	var updatedIds []string = []string{}

	return updatedIds
}

func (p *PanelLaunch) SetHwDeviceId(id int) {

}

func (p *PanelLaunch) RefreshHwStateIfMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int) string {
	return p.idStr
}

func (p PanelLaunch) ExposeVariables() map[string]string {

	var m map[string]string = map[string]string{}

	m["Panel.Id"] = p.idStr
	m["Panel.Title"] = p.title
	m["Panel.SubPage"] = p.subPage
	m["Panel.Index"] = fmt.Sprintf("%d", p.index)

	m["Panel.PowerInfo"] = "false"

	m["Panel.PanelCornerTitle"] = p.panelCornerTitle
	return m
}
