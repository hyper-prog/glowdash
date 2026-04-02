/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2026 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"

	"strconv"
)

type DeviceTypeCustom struct {
	DeviceTypeUnspecified

	customquerycode string
	customsetcode   string
}

func newCustomDevice(csetcode string, cquerycode string) DeviceTypeCustom {
	return DeviceTypeCustom{
		customsetcode:   csetcode,
		customquerycode: cquerycode,
	}
}

// --------------------------------- Custom scripted device methods ----------------------------------

func (d DeviceTypeCustom) SwitchTo(p DeviceHardwareInterface, toState bool, from string) SwitchSetResult {
	sr := SwitchSetResult{
		ok:     false,
		state:  0,
		updIds: []string{},
	}

	baseNameStr := "Switch"
	if from == "swaction" || from == "swscheduler" {
		baseNameStr = "SwitchPanel"
	}
	if from == "tswaction" || from == "tswscheduler" {
		baseNameStr = "ToggleSwitchPanel"
	}

	relatedPanels := []string{}
	initVariables := p.ExposeVariables()
	initVariables[baseNameStr+".Title"] = p.Title()
	initVariables[baseNameStr+".Id"] = p.IdStr()
	initVariables[baseNameStr+".DeviceType"] = p.DeviceType()
	initVariables[baseNameStr+".ActionName"] = "switch"
	initVariables["RequiredStateText"] = "false"
	if toState {
		initVariables["RequiredStateText"] = "true"
	}

	if from == "swaction" {
		GlowdashConsole.Write(T("Set switch \"{{title}}\" by custom code \"{{code}}\" to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "code": d.customsetcode, "state": T(initVariables["RequiredStateText"])}))
	}
	if from == "swscheduler" {
		GlowdashConsole.Write(T("Scheduled set switch \"{{title}}\" by custom code \"{{code}}\" to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "code": d.customsetcode, "state": T(initVariables["RequiredStateText"])}))
	}
	if from == "tswaction" {
		GlowdashConsole.Write(T("Set toggle switch \"{{title}}\" by custom code \"{{code}}\" to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "code": d.customsetcode, "state": T(initVariables["RequiredStateText"])}))
	}
	if from == "tswscheduler" {
		GlowdashConsole.Write(T("Scheduled set toggle switch \"{{title}}\" by custom code \"{{code}}\" to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "code": d.customsetcode, "state": T(initVariables["RequiredStateText"])}))
	}

	code, ok := ProgramLibrary[d.customsetcode]
	if !ok {
		sr.ok = false
		return sr
	}

	results := ExecuteCommands(code, initVariables, &relatedPanels)
	if DebugLevel >= 2 {
		fmt.Printf("Custom set code \"%s\" executed for panel %s, result: %s\n", d.customsetcode, p.Title(), results["Return"])
	}
	if results["Return"] == "error" {
		GlowdashConsole.Write(T("ERROR: The last operation failed to complete"))
		p.InvalidateInfo()
	}

	sr.state = 0
	if toState {
		sr.state = 1
	}

	sr.ok = true
	sr.updIds = append([]string{p.IdStr()}, getUpdatedIdsFromRelatedPanels(relatedPanels)...)
	return sr
}

func (d DeviceTypeCustom) QuerySwitch(p DeviceHardwareInterface, from string) SwitchQueryResult {
	qr := SwitchQueryResult{
		ok:            false,
		state:         0,
		inputstate:    0,
		powerMeasured: false,
		apower:        0.0,
		voltage:       0.0,
	}

	code, ok := ProgramLibrary[d.customquerycode]
	if !ok {
		qr.ok = false
		return qr
	}

	baseNameStr := "Switch"
	if from == "swaction" || from == "swscheduler" {
		baseNameStr = "SwitchPanel"
	}
	if from == "tswaction" || from == "tswscheduler" {
		baseNameStr = "ToggleSwitchPanel"
	}

	relatedPanels := []string{}
	initVariables := p.ExposeVariables()
	initVariables[baseNameStr+".Title"] = p.EventTitle()
	initVariables[baseNameStr+".Id"] = p.IdStr()
	initVariables[baseNameStr+".DeviceType"] = p.DeviceType()
	initVariables[baseNameStr+".ActionName"] = "update"

	results := ExecuteCommands(code, initVariables, &relatedPanels)
	if DebugLevel >= 2 {
		fmt.Printf("Custom query code \"%s\" executed for panel %s, result: %s\n", d.customquerycode, p.EventTitle(), results["Return"])
	}
	if results["Return"] == "error" {
		p.InvalidateInfo()
		qr.ok = false
		return qr
	}
	if results["Return"] == "true" {
		qr.state = 1
	}

	qr.inputstate = 0
	if inputstatestr, istrok := results["Return.InputState"]; istrok {
		if isv, iserr := strconv.Atoi(inputstatestr); iserr == nil {
			qr.inputstate = isv
		}
	}

	qr.powerMeasured = false
	qr.ok = true
	return qr
}
