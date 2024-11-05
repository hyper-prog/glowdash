/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
	"strings"

	"github.com/hyper-prog/smartyaml"
)

func (p PanelBase) PanelType() PanelTypes {
	return p.panelType
}

func (p PanelBase) IdStr() string {
	return p.idStr
}

func (p PanelBase) IsHide() bool {
	return p.hide
}

func (p PanelBase) Title() string {
	return p.title
}

func (p PanelBase) Sub() string {
	return p.subPage
}

func (p PanelBase) SubTo() string {
	return ""
}

func (p PanelBase) LaunchTo() string {
	return ""
}

func (p PanelBase) DeviceType() string {
	return p.deviceType
}

func (p *PanelBase) SetDeviceType(dt string) {
	p.deviceType = dt
}

func (p PanelBase) Index() int {
	return p.index
}

func (p *PanelBase) SetIndex(idx int) {
	p.index = idx
}

func (p PanelBase) PanelHtml(withContainer bool) string {
	return ""
}

func (p PanelBase) DoAction(actionName string) (string, []string) {
	return "", []string{}
}

func (p PanelBase) DoActionFromScheduler(actionName string) []string {
	return []string{}
}

func (p *PanelBase) LoadBaseConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	p.title = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/Title", indexInConfig), "-")
	p.subPage = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/SubPage", indexInConfig), "")
	p.idStr = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/Id", indexInConfig), fmt.Sprintf("autogenId%d", indexInConfig+1))
	p.thumbImg = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/Thumbnail", indexInConfig), "")
	p.deviceType = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/DeviceType", indexInConfig), "Unknown")

	p.hide = false
	if sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/Hide", indexInConfig), "") == "yes" {
		p.hide = true
	}
}

func (p *PanelBase) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
}

func (p PanelBase) IsActionIdMatch(aId string) bool {
	if "b-"+p.idStr == aId {
		return true
	}
	if "b-"+p.idStr+"-update" == aId {
		return true
	}
	return false
}

func (p PanelBase) GetActionIdFromUrl(fullurl string) string {

	if "b-"+p.idStr == fullurl {
		return ""
	}
	rightpart, found := strings.CutPrefix(fullurl, "b-"+p.idStr+"-")
	if found {
		return rightpart
	}
	return "unknown"
}

func (p PanelBase) RequiredActionParameters(actionName string) []string {
	return []string{}
}

func (p PanelBase) HandleActionEvent(res *ActionResponse, actionName string, parameters map[string]string) {
	if DebugLevel > 1 {
		fmt.Println("HandleActionEvent on " + p.idStr)
	}

	if p.panelType == Group {
		res.setResultString("ok")
		res.addCommandArg1("loadpage", "/subpage/"+Panels[p.Index()].SubTo())
		return
	}

	if p.panelType == Launch {
		res.setResultString("ok")
		res.addCommandArg1("loadpage", "/page/"+Panels[p.Index()].LaunchTo())
		return
	}

	str, updIds, stateChanged := Panels[p.Index()].DoAction(actionName, parameters)
	res.setResultString(str)
	if str == "ok" {
		var sourceCardUpdated bool = false
		for ii := 0; ii < len(updIds); ii++ {
			if updIds[ii] == p.idStr {
				sourceCardUpdated = true
			}
			res.addCommandArg2("sethtml", "#pc-"+updIds[ii], GetPanelById(updIds[ii]).PanelHtml(false))
		}
		if !sourceCardUpdated {
			updIds = append(updIds, p.idStr)
			res.addCommandArg2("sethtml", "#pc-"+p.idStr, GetPanelById(p.idStr).PanelHtml(false))
		}

		if stateChanged {
			if len(parameters["otsseid"]) > 0 {
				go sendSSENotify("panelupd-" + parameters["otsseid"] + "=refreshId(" + strings.Join(updIds, ",") + ")")
			} else {
				go sendSSENotify("panelupd=refreshId(" + strings.Join(updIds, ",") + ")")
			}

		}
	}
}

func (p PanelBase) IsHwMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int) bool {
	return false
}

func (p *PanelBase) InvalidateInfo() {
}
