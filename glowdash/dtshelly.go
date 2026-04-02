/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2026 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"

	"time"
)

type DeviceTypeShelly struct {
	DeviceTypeUnspecified
}

func newShellyDevice() DeviceTypeShelly {
	return DeviceTypeShelly{}
}

// --------------------------------- Shelly device methods ---------------------------------

func (d DeviceTypeShelly) DeviceHttpRequestAddr(p DeviceHardwareInterface) string {
	if p.TcpPort() == 80 {
		return fmt.Sprintf("http://%s", p.DeviceIp())
	}
	if p.TcpPort() == 443 {
		return fmt.Sprintf("https://%s", p.DeviceIp())
	}
	return fmt.Sprintf("http://%s:%d", p.DeviceIp(), p.TcpPort())
}

func (d DeviceTypeShelly) SwitchTo(p DeviceHardwareInterface, toState bool, from string) SwitchSetResult {
	sr := SwitchSetResult{
		ok:     false,
		state:  0,
		updIds: []string{},
	}

	if p.DeviceIp() == "" {
		p.InvalidateInfo()
		sr.ok = false
		return sr
	}

	tostr := "false"
	if toState {
		tostr = "true"
	}

	if from == "swaction" {
		GlowdashConsole.Write(T("Set Shelly switch \"{{title}}\" to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "state": T(tostr)}))
	}
	if from == "swscheduler" {
		GlowdashConsole.Write(T("Scheduled set Shelly switch \"{{title}}\" to &lt;{{sts}}&gt;",
			map[string]any{"title": p.EventTitle(), "sts": T(tostr)}))
	}
	if from == "tswaction" {
		GlowdashConsole.Write(T("Set Shelly toggle switch \"{{title}}\" to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "state": T(tostr)}))
	}
	if from == "tswscheduler" {
		GlowdashConsole.Write(T("Scheduled set Shelly toggle switch \"{{title}}\" to &lt;{{sts}}&gt;",
			map[string]any{"title": p.EventTitle(), "sts": T(tostr)}))
	}

	execUrl := fmt.Sprintf("%s/rpc/Switch.Set?id=%d&on=%s", d.DeviceHttpRequestAddr(p), p.InDeviceId(), tostr)
	ro := execJsonHttpQuery(execUrl)
	if !ro.Success {
		GlowdashConsole.Write(T("ERROR: The last operation failed to complete"))
		p.InvalidateInfo()
		sr.ok = false
		return sr
	}

	sr.state = 0
	if toState {
		sr.state = 1
	}
	sr.ok = true
	sr.updIds = []string{p.IdStr()}
	return sr
}

func (d DeviceTypeShelly) PerformThis(p DeviceHardwareInterface, fnc string, from string) PerformThisResult {
	pr := PerformThisResult{
		ok:     false,
		state:  0,
		updIds: []string{},
	}

	if p.DeviceIp() == "" {
		p.InvalidateInfo()
		pr.ok = false
		return pr
	}

	if fnc == "up" {
		if from == "action" {
			GlowdashConsole.Write(T("Set shading \"{{title}}\" to &lt;{{tst}}&gt;",
				map[string]any{"title": p.EventTitle(), "tst": T("up")}))
		}
		if from == "scheduler" {
			GlowdashConsole.Write(T("Scheduled set Shelly shading \"{{title}}\" to &lt;{{tst}}&gt;",
				map[string]any{"title": p.EventTitle(), "tst": T("open")}))
		}
		execUrl := fmt.Sprintf("%s/rpc/Cover.Open?id=%d", d.DeviceHttpRequestAddr(p), p.InDeviceId())
		ro := execJsonHttpQuery(execUrl)
		if !ro.Success {
			GlowdashConsole.Write(T("ERROR: The last operation failed to complete"))
			pr.ok = false
			p.InvalidateInfo()
			return pr
		}
		time.Sleep(time.Millisecond * 500) //Wait a little time to let the device do the operation
		pr.ok = true
		pr.updIds = []string{p.IdStr()}
		return pr
	}

	if fnc == "down" {
		if from == "action" {
			GlowdashConsole.Write(T("Set shading \"{{title}}\" to &lt;{{tst}}&gt;",
				map[string]any{"title": p.EventTitle(), "tst": T("down")}))
		}
		if from == "scheduler" {
			GlowdashConsole.Write(T("Scheduled set Shelly shading \"{{title}}\" to &lt;{{tst}}&gt;",
				map[string]any{"title": p.EventTitle(), "tst": T("close")}))

		}
		execUrl := fmt.Sprintf("%s/rpc/Cover.Close?id=%d", d.DeviceHttpRequestAddr(p), p.InDeviceId())
		ro := execJsonHttpQuery(execUrl)
		if !ro.Success {
			GlowdashConsole.Write(T("ERROR: The last operation failed to complete"))
			pr.ok = false
			p.InvalidateInfo()
			return pr
		}
		time.Sleep(time.Millisecond * 500) //Wait a little time to let the device do the operation
		pr.ok = true
		pr.updIds = []string{p.IdStr()}
		return pr
	}

	if fnc == "stop" {
		if from == "action" {
			GlowdashConsole.Write(T("Set shading \"{{title}}\" to &lt;{{tst}}&gt;",
				map[string]any{"title": p.EventTitle(), "tst": T("stop")}))
		}
		if from == "scheduler" {
			// No scheduler action for stop function, as it doesn't make sense to schedule a stop command for a shading device.
		}
		execUrl := fmt.Sprintf("%s/rpc/Cover.Stop?id=%d", d.DeviceHttpRequestAddr(p), p.InDeviceId())
		ro := execJsonHttpQuery(execUrl)
		if !ro.Success {
			GlowdashConsole.Write(T("ERROR: The last operation failed to complete"))
			pr.ok = false
			p.InvalidateInfo()
			return pr
		}
		time.Sleep(time.Millisecond * 500) //Wait a little time to let the device do the operation
		pr.ok = true
		pr.updIds = []string{p.IdStr()}
		return pr
	}
	return pr
}

func (d DeviceTypeShelly) ScriptTo(p DeviceHardwareInterface, scriptName string, fnc string, from string) PerformThisResult {
	pr := PerformThisResult{
		ok:     false,
		state:  0,
		updIds: []string{},
	}

	toStr := ""
	if fnc == "start" {
		toStr = "Start"
	}
	if fnc == "stop" {
		toStr = "Stop"
	}

	if toStr != "Start" && toStr != "Stop" {
		pr.ok = false
		return pr
	}

	if p.DeviceIp() == "" || scriptName == "" || p.InDeviceId() < 0 {
		p.InvalidateInfo()
		pr.ok = false
		return pr
	}

	if from == "action" {
		GlowdashConsole.Write(T("Set Shelly script \"{{title}}\" state to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "state": T(toStr)}))
	}
	if from == "scheduler" {
		GlowdashConsole.Write(T("Scheduled set Shelly script \"{{title}}\" state to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "state": T(toStr)}))
	}

	execUrl := fmt.Sprintf("%s/rpc/Script.%s?id=%d", d.DeviceHttpRequestAddr(p), toStr, p.InDeviceId())
	ro := execJsonHttpQuery(execUrl)
	if !ro.Success {
		GlowdashConsole.Write(T("ERROR: The last operation failed to complete"))
		p.InvalidateInfo()
		pr.ok = false
		return pr
	}

	pr.ok = true
	return pr
}

func (d DeviceTypeShelly) QuerySwitch(p DeviceHardwareInterface, from string) SwitchQueryResult {
	qr := SwitchQueryResult{
		ok:            false,
		state:         0,
		inputstate:    0,
		powerMeasured: false,
		apower:        0.0,
		voltage:       0.0,
	}

	if p.DeviceIp() == "" {
		p.InvalidateInfo()
		qr.ok = false
		return qr
	}

	execUrl := fmt.Sprintf("%s/rpc/Switch.GetStatus?id=%d", d.DeviceHttpRequestAddr(p), p.InDeviceId())
	jhq := execJsonHttpQuery(execUrl)
	if !jhq.Success {
		p.InvalidateInfo()
		qr.ok = false
		return qr
	}

	relaystate := jhq.SmartJSON.GetBoolByPathWithDefault("/output", false)
	if relaystate {
		qr.state = 1
	} else {
		qr.state = 0
	}

	if jhq.SmartJSON.NodeExists("/apower") && jhq.SmartJSON.NodeExists("/voltage") {
		str1 := ""
		str2 := ""
		qr.apower, str1 = jhq.SmartJSON.GetFloat64ByPath("/apower")
		qr.voltage, str2 = jhq.SmartJSON.GetFloat64ByPath("/voltage")
		if str1 == "float64" && str2 == "float64" && qr.apower >= 0.0 && qr.voltage >= 0.0 {
			qr.powerMeasured = true
		}
	}

	execUrl = fmt.Sprintf("%s/rpc/Input.GetStatus?id=%d", d.DeviceHttpRequestAddr(p), p.InDeviceId())
	jhq2 := execJsonHttpQuery(execUrl)
	if !jhq2.Success {
		p.InvalidateInfo()
		qr.ok = false
		return qr
	}
	istate := jhq2.SmartJSON.GetBoolByPathWithDefault("/state", false)
	if istate {
		qr.inputstate = 1
	} else {
		qr.inputstate = 0
	}
	qr.ok = true
	return qr
}

func (d DeviceTypeShelly) QueryShader(p DeviceHardwareInterface, queryExtInfo bool, from string) ShaderQueryResult {
	qr := ShaderQueryResult{
		ok:            false,
		position:      0.0,
		namedState:    "unknown",
		powerMeasured: false,
		apower:        0.0,
		voltage:       0.0,
	}

	if p.DeviceIp() == "" {
		p.InvalidateInfo()
		qr.ok = false
		return qr
	}

	execUrl := fmt.Sprintf("%s/rpc/Cover.GetStatus?id=%d", d.DeviceHttpRequestAddr(p), p.InDeviceId())
	jhq := execJsonHttpQuery(execUrl)
	if !jhq.Success {
		p.InvalidateInfo()
		qr.ok = false
		return qr
	}

	qr.position = jhq.SmartJSON.GetFloat64ByPathWithDefault("/current_pos", 0.0)
	qr.namedState = jhq.SmartJSON.GetStringByPathWithDefault("/state", "")

	if queryExtInfo && jhq.SmartJSON.NodeExists("/apower") && jhq.SmartJSON.NodeExists("/voltage") {
		str1 := ""
		str2 := ""
		qr.apower, str1 = jhq.SmartJSON.GetFloat64ByPath("/apower")
		qr.voltage, str2 = jhq.SmartJSON.GetFloat64ByPath("/voltage")
		if str1 == "float64" && str2 == "float64" && qr.apower >= 0.0 && qr.voltage >= 0.0 {
			qr.powerMeasured = true
		}
	}
	qr.ok = true
	return qr
}

func (d DeviceTypeShelly) QueryScript(p DeviceHardwareInterface, scriptName string, from string) ScriptQueryResult {
	sr := ScriptQueryResult{
		ok:    false,
		state: 0,
	}

	if p.DeviceIp() == "" || scriptName == "" {
		p.InvalidateInfo()
		sr.ok = false
		return sr
	}

	execUrl := fmt.Sprintf("%s/rpc/Script.List", d.DeviceHttpRequestAddr(p))
	jhq := execJsonHttpQuery(execUrl)
	if !jhq.Success {
		p.InvalidateInfo()
		sr.ok = false
		return sr
	}
	scriptcount := jhq.SmartJSON.GetCountDescendantsByPath("/scripts")
	for i := 0; i < scriptcount; i++ {
		sn := jhq.SmartJSON.GetStringByPathWithDefault(fmt.Sprintf("/scripts/[%d]/name", i), "")
		if sn == scriptName {
			p.SetHwDeviceId(int(jhq.SmartJSON.GetFloat64ByPathWithDefault(fmt.Sprintf("/scripts/[%d]/id", i), -1.0)))
			run, _ := jhq.SmartJSON.GetBoolByPath(fmt.Sprintf("/scripts/[%d]/running", i))
			if run {
				sr.state = 1
			} else {
				sr.state = 0
			}
			sr.ok = true
			return sr
		}
	}
	p.InvalidateInfo()
	sr.ok = false
	return sr
}
