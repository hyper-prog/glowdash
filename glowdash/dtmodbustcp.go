/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2026 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
)

type DeviceTypeModbusTCP struct {
	DeviceTypeUnspecified
}

func newModbusTCPDevice() DeviceTypeModbusTCP {
	return DeviceTypeModbusTCP{}
}

// -------------------------------- ModbusTCP driven device methods ----------------------------------

func (d DeviceTypeModbusTCP) SwitchTo(p DeviceHardwareInterface, toState bool, from string) SwitchSetResult {
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
		GlowdashConsole.Write(T("Set ModbusTCP switch \"{{title}}\" to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "state": T(tostr)}))
	}
	if from == "swscheduler" {
		GlowdashConsole.Write(T("Scheduled set ModbusTCP switch \"{{title}}\" to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "state": T(tostr)}))
	}
	if from == "tswaction" {
		GlowdashConsole.Write(T("Set ModbusTCP toggle switch \"{{title}}\" to &lt;{{state}}&gt;",
			map[string]any{"title": p.EventTitle(), "state": T(tostr)}))
	}
	if from == "tswscheduler" {
		GlowdashConsole.Write(T("Scheduled set ModbusTCP toggle switch \"{{title}}\" to &lt;{{sts}}&gt;",
			map[string]any{"title": p.EventTitle(), "sts": T(tostr)}))
	}

	modbulsClient, err := Dial(p.DeviceIp(), fmt.Sprintf("%d", p.TcpPort()), byte(p.UnitId()), BackgroudDevQueryNetDialerTimeout)
	if err != nil {
		GlowdashConsole.Write(T("ERROR: The last operation failed to complete"))
		p.InvalidateInfo()
		sr.ok = false
		return sr
	}
	defer modbulsClient.Close()

	err2 := modbulsClient.WriteSingleCoil(uint16(p.InDeviceId()), toState)
	if err2 != nil {
		GlowdashConsole.Write(T("ERROR: The last operation failed to complete"))
		p.InvalidateInfo()
		sr.ok = false
		return sr
	}
	sr.ok = true
	sr.updIds = []string{p.IdStr()}
	return sr
}

func (d DeviceTypeModbusTCP) QuerySwitch(p DeviceHardwareInterface, from string) SwitchQueryResult {
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

	modbulsClient, err := Dial(p.DeviceIp(), fmt.Sprintf("%d", p.TcpPort()), byte(p.UnitId()), BackgroudDevQueryNetDialerTimeout)
	if err != nil {
		p.InvalidateInfo()
		qr.ok = false
		return qr
	}
	defer modbulsClient.Close()

	coil, err2 := modbulsClient.ReadSingleCoil(uint16(p.InDeviceId()))
	if err2 != nil {
		p.InvalidateInfo()
		qr.ok = false
		return qr
	}
	if coil {
		qr.state = 1
	}
	istate, err3 := modbulsClient.ReadInputRegister(uint16(p.InDeviceId()))
	if err3 != nil {
		qr.inputstate = int(istate)
	}
	qr.ok = true
	return qr
}
