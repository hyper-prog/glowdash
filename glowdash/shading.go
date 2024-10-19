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

type PanelShading struct {
	PanelHwDevBased

	coverNamedState      string
	disablePosIndicator  bool
	enablePowerIndicator bool
}

func NewPanelShading() *PanelShading {
	return &PanelShading{
		PanelHwDevBased{
			PanelBase{
				idStr:       "",
				panelType:   Shading,
				title:       "",
				subPage:     "",
				thumbImg:    "",
				deviceType:  "",
				hide:        false,
				hasPoweInfo: false,
				index:       0,
			},
			false, "", 0, 0, 0, 0.0, 0.0,
		}, "unknown", false, false,
	}
}

func (p *PanelShading) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	if p.deviceType == "Shelly" {
		p.deviceIp = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/DeviceIp", indexInConfig), "")
		p.inDeviceId = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/InDeviceId", indexInConfig), 0)
		p.disablePosIndicator, _ = sy.GetBoolByPath(fmt.Sprintf("/GlowDash/Panels/[%d]/DisablePosIndicator", indexInConfig))
		p.enablePowerIndicator, _ = sy.GetBoolByPath(fmt.Sprintf("/GlowDash/Panels/[%d]/EnablePowerIndicator", indexInConfig))
	}
}

func (p PanelShading) PanelHtml(withContainer bool) string {
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
    	        <div class="title-container mt-xs">
    	            <p class="title text-bold body-small-styles">{{.Title}}</p>
    	        </div>
				{{if .HasPowerInfo}}
				<div class="ctrlline-container mt-s">
					<p class="text-600 title text-bold body-small-styles">
						<i class="fa fa-bolt"></i> {{.Watt}} W
						<i class="fa fa-circle-bolt"></i> {{.Volt}} V
					</p>
				</div>
				{{end}}
    	    </div>

			{{if .NoValidInfo}}
			<div class="ctrlline-container mt-s">
				<p class="text-600 title text-bold body-small-styles">No information</p>
			</div>
			{{else}}
				{{if .PosIndicator}}
				<div class="bottom-slot-container d-flex justify-content-center">
    	        	<button id="b-{{.Id}}" class="align-self-center shader-button">
   	            		<div class="shader-cover" style="position: absolute;top: 0px; left: 0px; width: 100%; height: {{.StateInv}}%;"></div>
   	            		<span>{{.State}}%</span>
    	        	</button>
    	    	</div>
				{{end}}
			{{end}}

    	    <div class="bottom-slot-container d-flex justify-content-center">
				<button id="b-{{.Id}}-down" 
						data-grpid="b-{{.Id}}"
						class="align-self-center device-button primary medium jsaction inactive zcombomove {{if .ClosingState}}displaynone{{end}} {{if .NoValidInfo}}noinfo{{end}}">
    	            <span class="device-action-border">
    	                <span class="device-action">
    	                    <span class="text-primary icon-grid icon-grid-s">
    	                        <i class="fa fa-down"></i>
    	                    </span>
    	                </span>
    	            </span>
    	        </button>
				<button id="b-{{.Id}}-stop" 
						data-grpid="b-{{.Id}}"
						class="align-self-center device-button primary medium jsaction inactive zcombostop {{if .MoveState}}justmove{{else}}displaynone{{end}} {{if .NoValidInfo}}noinfo{{end}}">
    	            <span class="device-action-border">
    	                <span class="device-action">
    	                    <span class="text-primary icon-grid icon-grid-s {{if .MoveState}}animated-border-box{{end}}">
    	                        <i class="fa fa-stop"></i>
    	                    </span>
    	                </span>
    	            </span>
    	        </button>
				<button id="b-{{.Id}}-up" 
						data-grpid="b-{{.Id}}"
				        class="align-self-center device-button primary medium jsaction inactive zcombomove {{if .OpeningState}}displaynone{{end}} {{if .NoValidInfo}}noinfo{{end}}">
    	            <span class="device-action-border">
    	                <span class="device-action">
    	                    <span class="text-primary icon-grid icon-grid-s">
    	                        <i class="fa fa-up"></i>
    	                    </span>
    	                </span>
    	            </span>
    	        </button>
    	    </div>
    	</div>`)

	var moveState bool = false
	if p.coverNamedState == "open" || p.coverNamedState == "closed" || p.coverNamedState == "stopped" {
		moveState = false
	}
	if p.coverNamedState == "opening" || p.coverNamedState == "closing" {
		moveState = true
	}

	pass := struct {
		Title        string
		Id           string
		ThumbImg     string
		State        int
		StateInv     int
		IpAddress    string
		HasPowerInfo bool
		PosIndicator bool
		NamedState   string
		MoveState    bool
		OpeningState bool
		ClosingState bool
		HasValidInfo bool
		NoValidInfo  bool
		Watt         string
		Volt         string
	}{
		Title:        p.title,
		Id:           p.idStr,
		ThumbImg:     p.thumbImg,
		State:        p.state,
		StateInv:     100 - p.state,
		IpAddress:    p.deviceIp,
		HasPowerInfo: p.hasPoweInfo && p.enablePowerIndicator,
		PosIndicator: !p.disablePosIndicator,
		NamedState:   p.coverNamedState,
		MoveState:    moveState,
		OpeningState: p.coverNamedState == "opening",
		ClosingState: p.coverNamedState == "closing",
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

func (p PanelShading) IsActionIdMatch(aId string) bool {
	if "b-"+p.idStr == aId {
		return true
	}
	if "b-"+p.idStr+"-up" == aId {
		return true
	}
	if "b-"+p.idStr+"-stop" == aId {
		return true
	}
	if "b-"+p.idStr+"-down" == aId {
		return true
	}
	if "b-"+p.idStr+"-update" == aId {
		return true
	}
	if "b-"+p.idStr+"-movupdate" == aId {
		return true
	}
	return false
}

func (p PanelShading) WaitUntilStateIsEqual(state string, waitMillisec int) {
	qstr := state
	for qstr == state {
		jhq := execJsonHttpQuery(fmt.Sprintf("http://%s/rpc/Cover.GetStatus?id=%d", p.deviceIp, p.inDeviceId))
		if !jhq.Success {
			p.InvalidateInfo()
			return
		}
		qstr = jhq.SmartJSON.GetStringByPathWithDefault("/state", "")
		time.Sleep(time.Millisecond * time.Duration(waitMillisec))
	}
}

func (p PanelShading) WaitUntilStateIsMoving(waitMillisec int, maxWaitMillisec int) {
	qstr := "opening"
	allWait := 0
	for qstr == "opening" || qstr == "closing" {
		jhq := execJsonHttpQuery(fmt.Sprintf("http://%s/rpc/Cover.GetStatus?id=%d", p.deviceIp, p.inDeviceId))
		if !jhq.Success {
			p.InvalidateInfo()
			return
		}
		qstr = jhq.SmartJSON.GetStringByPathWithDefault("/state", "")
		time.Sleep(time.Millisecond * time.Duration(waitMillisec))
		allWait += waitMillisec
		if allWait >= maxWaitMillisec {
			return
		}
	}
}

func (p PanelShading) DoActionCoverUp() {
	execUrl := fmt.Sprintf("http://%s/rpc/Cover.Open?id=%d", p.deviceIp, p.inDeviceId)
	ro := execJsonHttpQuery(execUrl)
	if !ro.Success {
		p.InvalidateInfo()
		return
	}
	time.Sleep(time.Millisecond * 500) //Wait a little time to let the device do the operation
}

func (p PanelShading) DoActionCoverDown() {
	execUrl := fmt.Sprintf("http://%s/rpc/Cover.Close?id=%d", p.deviceIp, p.inDeviceId)
	ro := execJsonHttpQuery(execUrl)
	if !ro.Success {
		p.InvalidateInfo()
		return
	}
	time.Sleep(time.Millisecond * 500) //Wait a little time to let the device do the operation
}

func (p PanelShading) DoActionCoverStop() {
	execUrl := fmt.Sprintf("http://%s/rpc/Cover.Stop?id=%d", p.deviceIp, p.inDeviceId)
	ro := execJsonHttpQuery(execUrl)
	if !ro.Success {
		p.InvalidateInfo()
		return
	}
	time.Sleep(time.Millisecond * 200) //Wait a little time to let the device do the operation
}

func (p PanelShading) DoAction(actionName string, parameters map[string]string) (string, []string) {
	var updatedIds []string = []string{}

	if p.deviceType == "Shelly" && p.deviceIp != "" {
		if actionName == "up" {
			p.DoActionCoverUp()
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
		if actionName == "down" {
			p.DoActionCoverDown()
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
		if actionName == "stop" {
			p.DoActionCoverStop()
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
		if actionName == "update" {
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
		if actionName == "movupdate" {
			p.WaitUntilStateIsMoving(1000, 2000)
			updatedIds = append(updatedIds, p.QueryDevice()...)
		}
	}
	return "ok", updatedIds
}

func (p PanelShading) DoActionFromScheduler(actionName string) []string {
	if p.deviceType == "Shelly" && p.deviceIp != "" {
		if actionName == "open" {
			p.DoActionCoverUp()
			return p.QueryDevice()
		}
		if actionName == "close" {
			p.DoActionCoverDown()
			return p.QueryDevice()
		}
	}
	return []string{}
}

func (p PanelShading) QueryDevice() []string {
	var updatedIds []string = []string{}

	if p.deviceType == "Shelly" && p.deviceIp != "" {
		execUrl := fmt.Sprintf("http://%s/rpc/Cover.GetStatus?id=%d", p.deviceIp, p.inDeviceId)
		jhq := execJsonHttpQuery(execUrl)
		if !jhq.Success {
			p.InvalidateInfo()
			return []string{p.idStr}
		}
		current_pos := jhq.SmartJSON.GetFloat64ByPathWithDefault("/current_pos", 0.0)
		current_ns := jhq.SmartJSON.GetStringByPathWithDefault("/state", "")

		var powerMeasured bool = false
		var apower float64 = 0.0
		var voltage float64 = 0.0

		if jhq.SmartJSON.NodeExists("/apower") && jhq.SmartJSON.NodeExists("/voltage") {
			str1 := ""
			str2 := ""
			apower, str1 = jhq.SmartJSON.GetFloat64ByPath("/apower")
			voltage, str2 = jhq.SmartJSON.GetFloat64ByPath("/voltage")
			if str1 == "float64" && str2 == "float64" && apower >= 0.0 && voltage >= 0.0 {
				powerMeasured = true
			}
		}

		updatedIds = append(updatedIds, p.RefreshHwStatesInRequiredPanelsCover(int(current_pos), current_ns, powerMeasured, apower, voltage)...)
	}
	return updatedIds
}

func (p *PanelShading) RefreshHwStatesInRequiredPanelsCover(State int, coverNamedState string, PowMet bool, Watt float64, Volt float64) []string {
	var updatedIds []string = []string{}

	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if Panels[i].PanelType() == Shading {
			ps, ok := Panels[i].(*PanelShading)
			if ok {
				rId := ps.RefreshHwStateIfMatchCover(p.panelType, p.deviceIp, p.inDeviceId, "", State, coverNamedState, PowMet, Watt, Volt)
				if rId != "" {
					updatedIds = append(updatedIds, rId)
				}
			}
		}
	}
	return updatedIds
}

func (p *PanelShading) RefreshHwStateIfMatchCover(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, coverNamedState string, PowMet bool, Watt float64, Volt float64) string {
	if p.panelType == fromPanelType && p.deviceIp == fromDeviceIp && p.inDeviceId == fromInDeviceId {
		p.state = State
		p.coverNamedState = coverNamedState
		p.hasValidInfo = true
		p.hasPoweInfo = PowMet
		p.watt = Watt
		p.volt = Volt
		return p.idStr
	}
	return ""
}

func (p PanelShading) ExposeVariables() map[string]string {

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
	m["Panel.NamedState"] = p.coverNamedState
	m["Panel.Watt"] = fmt.Sprintf("%.2f", p.watt)
	m["Panel.Volt"] = fmt.Sprintf("%.2f", p.volt)

	return m
}
