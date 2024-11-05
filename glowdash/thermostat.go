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
	"time"

	"github.com/hyper-prog/smartyaml"
)

type PanelThermostat struct {
	PanelBase

	workingOn     bool
	tartgetTemp   float32
	referenceTemp float32
	heatingOn     bool

	hasValidInfo bool
	hwDeviceIp   string
	hwDevicePort int
}

func NewPanelThermostat() *PanelThermostat {
	return &PanelThermostat{
		PanelBase{
			idStr:       "",
			panelType:   Thermostat,
			title:       "",
			subPage:     "",
			thumbImg:    "",
			deviceType:  "",
			hasPoweInfo: false,
			index:       0,
		},
		false, 20.0, 0.0, false, false, "", 0,
	}
}

func NewPanelThermostatSwitch() *PanelThermostat {
	return &PanelThermostat{
		PanelBase{
			idStr:       "",
			panelType:   ThermostatSwitch,
			title:       "",
			subPage:     "",
			thumbImg:    "",
			deviceType:  "",
			hasPoweInfo: false,
			index:       0,
		},
		false, 20.0, 0.0, false, false, "", 0,
	}
}

func (p *PanelThermostat) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	if p.deviceType == "smtherm" {
		p.hwDeviceIp = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/DeviceIp", indexInConfig), "")
		p.hwDevicePort = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/DeviceTcpPort", indexInConfig), 5017)
	}
}

func (p PanelThermostat) PanelHtml(withContainer bool) string {
	var templ *template.Template
	if p.panelType == Thermostat {
		templ, _ = template.New("PcT").Parse(`
		<div class="badge badge-left" style="max-width: 100%;">
			<div class="label label-s no-radius-bottom-left-diagonal">
				<span class="mr-xs icon-grid icon-grid-xs"><i class="fas fa-microchip"></i></span>
				<div class="label-value-container">
					<p class="text-600 miniature-styles text-nowrap">Device</p>
				</div>
			</div>
		</div>
	
		<div class="thermostatpanel main-container {{if .NoValidInfo}}panelnoinfo{{end}}" data-refid="b-{{.Id}}">
			<div class="main-container-top">
				<div class="gauge-container">
    				<div class="gauge">
        				<svg viewBox="0 0 100 60">
			            	<path d="M 10 50 A 40 40 0 0 1 90 50" fill="none" stroke="#eee" stroke-width="7" />
			            	<path id="thermostatTempArc-{{.Id}}" d="M 10 50 A 40 40 0 0 1 10 50" fill="none" stroke="blue" stroke-width="7" />
                        	<text x="10" y="58" text-anchor="middle">5°C</text>
            				<text x="90" y="58" text-anchor="middle">30°C</text>
        				</svg>
    				</div>
       				<div class="temperature-display {{if .HasValidInfo}}{{if .WorkingOn}}gauge-uncarved{{end}}{{end}}" data-fragval="30" data-mid="{{.Id}}" id="thermostatTemperatureDisplay-{{.Id}}">
					{{if .NoValidInfo}}
					--
					{{else}}
						{{if .WorkingOn}}
							{{.TargetTempStr}}
						{{else}}
							<i class="fa fa-power-off"></i>
						{{end}}
					{{end}}
					</div>
				</div>

				{{if .ShowTitle}}
				<div class="title-container mt-s">
					<p class="title text-bold body-small-styles">{{.Title}}</p>
				</div>
				{{end}}

				{{if .HasValidInfo}}
					<div class="ctrlline-container mt-s">
						<p class="text-600 title text-bold body-small-styles">
							<i class="fa fa-temp"></i> {{.ReferenceTempStr}} C
							&nbsp;&nbsp;
							{{if .HeatingOn}}
								<i class="fa fa-heaton"></i>
							{{else}}
								<i class="fa fa-heatoff"></i>
							{{end}}
						</p>
					</div>
				{{end}}
			</div>
		    
			<div class="bottom-slot-container d-flex justify-content-center">
    	    	<button id="b-{{.Id}}-down" 
							data-grpid="b-{{.Id}}"
							class="align-self-center device-button primary medium jsaction inactive thermobtn {{if .ButtonsDisabled}}noinfo{{end}}"
							{{if .ButtonsDisabled}}disabled{{end}}>
    	            	<span class="device-action-border">
    	                	<span class="device-action">
    	                    	<span class="text-primary icon-grid icon-grid-s">
    	                        	<i class="fa fa-tdown"></i>
    	                    	</span>
    	                	</span>
    	            	</span>
    	    	</button>
				<button id="b-{{.Id}}-up" 
						data-grpid="b-{{.Id}}"
			        	class="align-self-center device-button primary medium jsaction inactive thermobtn {{if .ButtonsDisabled}}noinfo{{end}}" 
						{{if .ButtonsDisabled}}disabled{{end}}>
    	            	<span class="device-action-border">
    	                	<span class="device-action">
    	                    	<span class="text-primary icon-grid icon-grid-s">
    	                        	<i class="fa fa-tup"></i>
    	                    	</span>
    	                	</span>
    	            	</span>
    	    	</button>			
    		</div>
		</div>`)
	}

	if p.panelType == ThermostatSwitch {
		templ, _ = template.New("PcT").Parse(`
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
			</div>

			<div class="bottom-slot-container d-flex justify-content-center">
				<button id="b-{{.Id}}-switch" class="align-self-center device-button primary medium jsaction {{if .WorkingOff}}inactive{{end}} {{if .NoValidInfo}}inactive noinfo{{end}}">
					<span class="device-action-border">
						<span class="device-action">
							<span class="text-primary icon-grid icon-grid-s">
								<i class="fa fa-heater"></i>
							</span>
						</span>
					</span>
				</button>
			</div>
		</div>`)
	}

	pass := struct {
		Title            string
		ShowTitle        bool
		Id               string
		ThumbImg         string
		HasValidInfo     bool
		NoValidInfo      bool
		TargetTemp       float32
		TargetTempStr    string
		ReferenceTemp    float32
		ReferenceTempStr string
		HeatingOn        bool
		WorkingOn        bool
		WorkingOff       bool
		ButtonsDisabled  bool
	}{
		Title:            p.title,
		ShowTitle:        false,
		Id:               p.idStr,
		ThumbImg:         p.thumbImg,
		HasValidInfo:     p.hasValidInfo,
		NoValidInfo:      !p.hasValidInfo,
		TargetTemp:       p.tartgetTemp,
		TargetTempStr:    fmt.Sprintf("%.1f", p.tartgetTemp),
		ReferenceTemp:    p.referenceTemp,
		ReferenceTempStr: fmt.Sprintf("%.1f", p.referenceTemp),
		HeatingOn:        p.heatingOn,
		WorkingOn:        p.workingOn,
		WorkingOff:       !p.workingOn,
		ButtonsDisabled:  !p.workingOn || !p.hasValidInfo,
	}

	buffer := bytes.Buffer{}
	templ.Execute(&buffer, pass)
	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"widget-card\" tabindex=\"-1\">", p.IdStr()) +
			buffer.String() + "</div>"
	}

	return buffer.String()
}

func (p *PanelThermostat) SetHwDeviceId(id int) {

}

func (p *PanelThermostat) RefreshHwStateIfMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int) string {
	return p.idStr
}

func (p *PanelThermostat) InvalidateInfo() {
	p.hasValidInfo = false
}

func (p PanelThermostat) IsActionIdMatch(aId string) bool {
	if "b-"+p.idStr == aId {
		return true
	}
	if "b-"+p.idStr+"-update" == aId {
		return true
	}
	if "b-"+p.idStr+"-switch" == aId {
		return true
	}
	if strings.HasPrefix(aId, "b-"+p.idStr+"-tts/") {
		return true
	}
	return false
}

func (p PanelThermostat) DoAction(actionName string, parameters map[string]string) (string, []string, bool) {
	var stateChanged bool = false
	var updatedIds []string = []string{}

	if p.deviceType == "smtherm" && p.hwDeviceIp != "" {

		if actionName == "update" {
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}

		if actionName == "switch" {
			actstr := "on"
			if p.workingOn {
				actstr = "off"
			}

			execTcpQuery(p.hwDeviceIp, p.hwDevicePort, fmt.Sprintf("cmd:stw;work:%s;", actstr))
			stateChanged = true
			time.Sleep(time.Millisecond * 500)
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}

		if strings.HasPrefix(actionName, "tts/") {
			fv, err := strconv.ParseFloat(actionName[4:], 32)
			if err == nil {
				execTcpQuery(p.hwDeviceIp, p.hwDevicePort, fmt.Sprintf("cmd:stt;ttemp:%.1f;", fv))
				stateChanged = true
				time.Sleep(time.Millisecond * 500)
			}
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
	}
	return "ok", updatedIds, stateChanged
}

func (p PanelThermostat) DoActionFromScheduler(actionName string) []string {
	if p.deviceType == "smtherm" && p.hwDeviceIp != "" {
		f, converr := strconv.ParseFloat(actionName, 8)
		if converr == nil {
			execTcpQuery(p.hwDeviceIp, p.hwDevicePort, fmt.Sprintf("cmd:stt;ttemp:%.1f;", f))
			time.Sleep(time.Millisecond * 500)
			return p.QueryDevice()
		}
	}
	return []string{}
}

func (p PanelThermostat) QueryDevice() []string {
	var updatedIds []string = []string{}
	if p.deviceType == "smtherm" {
		j := execJsonTcpQuery(p.hwDeviceIp, p.hwDevicePort, "cmd:qtt;")
		if j.Success {
			s_won := j.SmartJSON.GetStringByPathWithDefault("/working", "")
			f_tt := j.SmartJSON.GetFloat64ByPathWithDefault("/target_temp", -100.0)
			f_rt := j.SmartJSON.GetFloat64ByPathWithDefault("/reference_temp", -100.0)
			s_ho := j.SmartJSON.GetStringByPathWithDefault("/heating_state", "")
			if s_won != "" && f_tt > -100 && f_rt > -100 && (s_ho == "on" || s_ho == "off") {
				b_ho := false
				if s_ho == "on" {
					b_ho = true
				}
				won := false
				if s_won == "on" {
					won = true
				}
				updatedIds = append(updatedIds, p.RefreshHwStatesInRequiredPanelsThermostat(won, float32(f_tt), float32(f_rt), b_ho)...)
			}
		} else {
			p.InvalidateInfo()
			updatedIds = append(updatedIds, p.idStr)
		}
	}
	return updatedIds
}

func (p *PanelThermostat) RefreshHwStatesInRequiredPanelsThermostat(won bool, tt float32, rt float32, ho bool) []string {
	var updatedIds []string = []string{}

	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if Panels[i].PanelType() == Thermostat || Panels[i].PanelType() == ThermostatSwitch {
			ps, ok := Panels[i].(*PanelThermostat)
			if ok {
				rId := ps.RefreshHwStateIfMatchThermostat(p.panelType, p.hwDeviceIp, p.hwDevicePort, won, tt, rt, ho)
				if rId != "" {
					updatedIds = append(updatedIds, rId)
				}
			}
		}
	}
	return updatedIds
}

func (p *PanelThermostat) RefreshHwStateIfMatchThermostat(fromPanelType PanelTypes, fromDeviceIp string, fromDevicePort int, won bool, tt float32, rt float32, ho bool) string {
	if (fromPanelType == Thermostat || fromPanelType == ThermostatSwitch) &&
		(p.panelType == Thermostat || p.panelType == ThermostatSwitch) &&
		p.hwDeviceIp == fromDeviceIp && p.hwDevicePort == fromDevicePort {
		p.tartgetTemp = tt
		p.referenceTemp = rt
		p.heatingOn = ho
		p.workingOn = won
		p.hasValidInfo = true
		return p.idStr
	}
	return ""
}

func (p PanelThermostat) ExposeVariables() map[string]string {

	var m map[string]string = map[string]string{}

	m["Panel.Id"] = p.idStr
	m["Panel.Title"] = p.title
	m["Panel.DeviceType"] = p.deviceType
	m["Panel.SubPage"] = p.subPage
	m["Panel.Index"] = fmt.Sprintf("%d", p.index)
	return m
}
