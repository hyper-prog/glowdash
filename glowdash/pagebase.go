/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2026 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
	"net/http"

	"github.com/hyper-prog/smartyaml"
)

func (p PageBase) PageType() PageTypes {
	return p.pageType
}

func (p PageBase) IdStr() string {
	return p.idStr
}

func (p PageBase) Title() string {
	return p.title
}

func (p PageBase) DeviceType() string {
	return p.deviceType
}

func (p PageBase) Index() int {
	return p.index
}

func (p *PageBase) SetIndex(idx int) {
	p.index = idx
}

func (p PageBase) PageHtml(withContainer bool, r *http.Request) string {
	return ""
}

func (p *PageBase) LoadBaseConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	p.idStr = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Pages/[%d]/PageName", indexInConfig), fmt.Sprintf("autogenId%d", indexInConfig+1))
	p.title = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Pages/[%d]/Title", indexInConfig), "")
	p.deviceType = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Pages/[%d]/DeviceType", indexInConfig), "Unknown")
}

func (p *PageBase) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
}

func (p PageBase) IsActionIdMatch(aId string) bool {
	return false
}

func (p PageBase) GetActionIdFromUrl(full string) string {
	return full
}

func (p PageBase) RequiredActionParameters(actionName string) []string {
	return []string{}
}

func (p PageBase) HandleActionEvent(res *ActionResponse, actionName string, parameters map[string]string) {
}
