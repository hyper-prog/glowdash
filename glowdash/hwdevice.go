/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2026 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"

	"github.com/hyper-prog/smartyaml"
)

type DeviceHardwareInterface interface {
	IdStr() string
	Title() string
	EventTitle() string
	DeviceType() string
	DeviceIp() string
	InDeviceId() int
	UnitId() int
	TcpPort() int

	State() int
	InputState() int
	HasValidInfo() bool
	HasPowerInfo() bool
	Watt() float64
	Volt() float64

	ExposeVariables() map[string]string

	SetHwDeviceId(id int)
	InvalidateInfo()
}

type PanelHwDevBased struct {
	PanelBase

	deviceHandler DeviceManipulatorInterface

	hasValidInfo bool

	deviceIp   string
	inDeviceId int
	unitId     int
	tcpPort    int

	state      int
	inputState int
	watt       float64
	volt       float64
}

type HwDevTrunk struct {
	hasValidInfo bool
	hasPowerInfo bool

	deviceIp   string
	inDeviceId int
	unitId     int
	tcpPort    int

	state      int
	inputState int
	watt       float64
	volt       float64
}

func (p *PanelHwDevBased) DeviceIp() string {
	return p.deviceIp
}

func (p PanelHwDevBased) InDeviceId() int {
	return p.inDeviceId
}

func (p *PanelHwDevBased) UnitId() int {
	return p.unitId
}

func (p *PanelHwDevBased) TcpPort() int {
	return p.tcpPort
}

func (p *PanelHwDevBased) State() int {
	return p.state
}

func (p *PanelHwDevBased) InputState() int {
	return p.inputState
}

func (p *PanelHwDevBased) HasValidInfo() bool {
	return p.hasValidInfo
}

func (p *PanelHwDevBased) HasPowerInfo() bool {
	return p.hasPowerInfo
}

func (p *PanelHwDevBased) Watt() float64 {
	return p.watt
}

func (p *PanelHwDevBased) Volt() float64 {
	return p.volt
}

func (p *PanelHwDevBased) SetHwDeviceId(id int) {
	p.inDeviceId = id
}

func (p *PanelHwDevBased) SetUnitId(id int) {
	p.unitId = id
}

func (p *PanelHwDevBased) SetTcpPort(port int) {
	p.tcpPort = port
}

func (p *PanelHwDevBased) InitDeviceManipulator(sy smartyaml.SmartYAML, indexInConfig int) {
	if p.deviceType == "Custom" {
		p.deviceHandler = newCustomDevice(
			sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/CustomSetCode", indexInConfig), ""),
			sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/CustomQueryCode", indexInConfig), ""),
		)
	}

	if p.deviceType == "Shelly" {
		p.deviceHandler = newShellyDevice()
	}

	if p.deviceType == "ModbusTCP" {
		p.deviceHandler = newModbusTCPDevice()
	}
}

func (p *PanelHwDevBased) LoadHwDevConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	if p.deviceType == "Shelly" || p.deviceType == "ModbusTCP" {
		p.deviceIp = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/DeviceIp", indexInConfig), "")
		p.inDeviceId = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/InDeviceId", indexInConfig), 0)

		if p.deviceType == "Shelly" {
			p.tcpPort = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/TcpPort", indexInConfig), 80)
		}

		if p.deviceType == "ModbusTCP" {
			p.tcpPort = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/TcpPort", indexInConfig), 502)
			p.unitId = sy.GetIntegerByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/UnitId", indexInConfig), 1)
		}
	}
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

func (p PanelHwDevBased) IsIpAddressMatch(deviceIp string) bool {
	if p.deviceIp == deviceIp {
		return true
	}
	return false
}

func (p *PanelHwDevBased) RefreshHwStateIfMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int) string {
	if p.panelType == fromPanelType && p.deviceIp == fromDeviceIp && p.inDeviceId == fromInDeviceId {
		p.state = State
		p.inputState = InputState
		p.hasValidInfo = true
		p.hasPowerInfo = false
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

// -------------------HwDevTrunk-------------------

func (d HwDevTrunk) IdStr() string {
	return ""
}

func (d HwDevTrunk) Title() string {
	return ""
}

func (d HwDevTrunk) EventTitle() string {
	return ""
}

func (d HwDevTrunk) DeviceType() string {
	return ""
}

func (d HwDevTrunk) DeviceIp() string {
	return d.deviceIp
}

func (d HwDevTrunk) InDeviceId() int {
	return d.inDeviceId
}

func (d HwDevTrunk) UnitId() int {
	return d.unitId
}

func (d HwDevTrunk) TcpPort() int {
	return d.tcpPort
}

func (d HwDevTrunk) State() int {
	return d.state
}

func (d HwDevTrunk) InputState() int {
	return d.inputState
}

func (d HwDevTrunk) HasValidInfo() bool {
	return d.hasValidInfo
}

func (d HwDevTrunk) HasPowerInfo() bool {
	return d.hasPowerInfo
}

func (d HwDevTrunk) Watt() float64 {
	return d.watt
}

func (d HwDevTrunk) Volt() float64 {
	return d.volt
}

func (d HwDevTrunk) ExposeVariables() map[string]string {
	return map[string]string{}
}

func (d *HwDevTrunk) SetHwDeviceId(id int) {
	d.inDeviceId = id
}

func (d *HwDevTrunk) InvalidateInfo() {
	d.hasValidInfo = false
}
