/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

type PanelHwDevBased struct {
	PanelBase

	hasValidInfo bool

	deviceIp   string
	inDeviceId int

	state      int
	inputState int
	watt       float64
	volt       float64
}

func (p *PanelHwDevBased) SetHwDeviceId(id int) {
	p.inDeviceId = id
}

func (p *PanelHwDevBased) RefreshHwStatesInRequiredPanels(State int, InputState int) []string {
	var updatedIds []string = []string{}

	pc := len(Panels)
	for i := 0; i < pc; i++ {
		rId := Panels[i].RefreshHwStateIfMatch(p.panelType, p.deviceIp, p.inDeviceId, "", State, InputState)
		if rId != "" {
			updatedIds = append(updatedIds, rId)
		}
	}
	return updatedIds
}

func (p PanelHwDevBased) IsHwMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int) bool {
	if p.panelType == fromPanelType && p.deviceIp == fromDeviceIp && p.inDeviceId == fromInDeviceId {
		return true
	}
	return false
}

func (p *PanelHwDevBased) RefreshHwStateIfMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int) string {
	if p.panelType == fromPanelType && p.deviceIp == fromDeviceIp && p.inDeviceId == fromInDeviceId {
		p.state = State
		p.inputState = InputState
		p.hasValidInfo = true
		p.hasPoweInfo = false
		p.watt = 0.0
		p.volt = 0.0
		return p.idStr
	}
	return ""
}

func (p *PanelHwDevBased) InvalidateInfo() {
	p.hasValidInfo = false
}

func UpdateFirstHwPanel(pt PanelTypes, ip string, id int) []string {
	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if Panels[i].IsHwMatch(pt, ip, id) {
			return Panels[i].QueryDevice()
		}
	}
	return []string{}
}
