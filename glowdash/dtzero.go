/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2026 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

type DeviceManipulatorInterface interface {
	SwitchTo(p DeviceHardwareInterface, toState bool, from string) SwitchSetResult
	PerformThis(p DeviceHardwareInterface, fnc string, from string) PerformThisResult
	ScriptTo(p DeviceHardwareInterface, scriptName string, fnc string, from string) PerformThisResult
	QuerySwitch(p DeviceHardwareInterface, from string) SwitchQueryResult
	QueryShader(p DeviceHardwareInterface, queryExtInfo bool, from string) ShaderQueryResult
	QueryScript(p DeviceHardwareInterface, scriptName string, from string) ScriptQueryResult
}

type SwitchSetResult struct {
	ok     bool
	state  int
	updIds []string
}

type PerformThisResult struct {
	ok     bool
	state  int
	updIds []string
}

type SwitchQueryResult struct {
	ok            bool
	state         int
	inputstate    int
	powerMeasured bool
	apower        float64
	voltage       float64
}

type ShaderQueryResult struct {
	ok            bool
	position      float64
	namedState    string
	powerMeasured bool
	apower        float64
	voltage       float64
}

type ScriptQueryResult struct {
	ok    bool
	state int
}

type DeviceTypeUnspecified struct {
}

func newUnspecifiedDevice() DeviceTypeUnspecified {
	return DeviceTypeUnspecified{}
}

func (d DeviceTypeUnspecified) SwitchTo(p DeviceHardwareInterface, toState bool, from string) SwitchSetResult {
	p.InvalidateInfo()
	return SwitchSetResult{
		ok:     false,
		state:  0,
		updIds: []string{},
	}
}

func (d DeviceTypeUnspecified) PerformThis(p DeviceHardwareInterface, fnc string, from string) PerformThisResult {
	p.InvalidateInfo()
	return PerformThisResult{
		ok:     false,
		state:  0,
		updIds: []string{},
	}
}

func (d DeviceTypeUnspecified) ScriptTo(p DeviceHardwareInterface, scriptName string, fnc string, from string) PerformThisResult {
	p.InvalidateInfo()
	return PerformThisResult{
		ok:     false,
		state:  0,
		updIds: []string{},
	}
}

func (d DeviceTypeUnspecified) QuerySwitch(p DeviceHardwareInterface, from string) SwitchQueryResult {
	p.InvalidateInfo()
	return SwitchQueryResult{
		ok:            false,
		state:         0,
		inputstate:    0,
		powerMeasured: false,
		apower:        0.0,
		voltage:       0.0,
	}
}

func (d DeviceTypeUnspecified) QueryShader(p DeviceHardwareInterface, queryExtInfo bool, from string) ShaderQueryResult {
	p.InvalidateInfo()
	return ShaderQueryResult{
		ok:            false,
		position:      0.0,
		namedState:    "unknown",
		powerMeasured: false,
		apower:        0.0,
		voltage:       0.0,
	}
}

func (d DeviceTypeUnspecified) QueryScript(p DeviceHardwareInterface, scriptName string, from string) ScriptQueryResult {
	p.InvalidateInfo()
	return ScriptQueryResult{
		ok:    false,
		state: 0,
	}
}
