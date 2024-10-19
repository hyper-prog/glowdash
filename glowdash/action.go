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
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hyper-prog/smartyaml"
)

type PanelAction struct {
	PanelBase

	Commands      string
	RelatedPanels []string
}

func NewPanelAction() *PanelAction {
	return &PanelAction{
		PanelBase{
			idStr:       "",
			panelType:   Action,
			title:       "",
			subPage:     "",
			thumbImg:    "",
			deviceType:  "",
			hide:        false,
			hasPoweInfo: false,
			index:       0,
		},
		"", []string{},
	}
}

func (p *PanelAction) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	p.Commands = sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/Commands", indexInConfig), "")
	if sy.NodeExists(fmt.Sprintf("/GlowDash/Panels/[%d]/CommandFile", indexInConfig)) {
		commandFile := sy.GetStringByPathWithDefault(fmt.Sprintf("/GlowDash/Panels/[%d]/CommandFile", indexInConfig), "")
		commandFileProgram, commandFileErr := ioutil.ReadFile(commandFile)
		if commandFileErr != nil {
			log.Printf("Error, cannot read external program file: %s\n", commandFileErr.Error())
		} else {
			p.Commands = string(commandFileProgram)
		}
	}
}

func (p PanelAction) PanelHtml(withContainer bool) string {
	templ, _ := template.New("PcT").Parse(`
	<div class="badge badge-left" style="max-width: 100%;">
		<div class="label label-s no-radius-bottom-left-diagonal">
			<span class="mr-xs icon-grid icon-grid-xs"><i class="fas fa-program2"></i></span>
			<div class="label-value-container">
				<p class="text-600 miniature-styles text-nowrap">Action</p>
			</div>
		</div>
	</div>

	<div class="main-container">
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
			<button id="b-{{.Id}}" class="align-self-center device-button primary medium jsaction {{if eq .State 0}}inactive{{end}}">
				<span class="device-action-border">
					<span class="device-action">
						<span class="text-primary icon-grid icon-grid-s">
							<i class="fa fa-action"></i>
						</span>
					</span>
				</span>
			</button>
		</div>
	</div>`)

	pass := struct {
		Title    string
		Id       string
		ThumbImg string
		State    int
	}{
		Title:    p.title,
		Id:       p.idStr,
		ThumbImg: p.thumbImg,
		State:    0,
	}

	buffer := bytes.Buffer{}
	templ.Execute(&buffer, pass)

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"widget-card\" tabindex=\"-1\">", p.IdStr()) +
			buffer.String() + "</div>"
	}

	return buffer.String()
}

func Contains(strings []string, needle string) bool {
	sc := len(strings)
	for i := 0; i < sc; i++ {
		if strings[i] == needle {
			return true
		}
	}
	return false
}

func (p PanelAction) DoAction(actionName string, parameters map[string]string) (string, []string) {
	var updatedIds []string = []string{}

	p.RelatedPanels = []string{}
	initVariables := map[string]string{}
	initVariables["ActionPanel.RunType"] = "UserAction"
	initVariables["ActionPanel.Title"] = p.title
	initVariables["ActionPanel.Id"] = p.idStr
	initVariables["ActionPanel.DeviceType"] = p.deviceType
	ExecuteCommands(p.Commands, initVariables, &(p.RelatedPanels))
	updatedIds = append(updatedIds, p.QueryDevice()...)
	return "ok", updatedIds
}

func (p PanelAction) DoActionFromScheduler(actionName string) []string {
	if actionName == "run" {
		p.RelatedPanels = []string{}
		initVariables := map[string]string{}
		initVariables["ActionPanel.RunType"] = "ScheduledTask"
		initVariables["ActionPanel.Title"] = p.title
		initVariables["ActionPanel.Id"] = p.idStr
		initVariables["ActionPanel.DeviceType"] = p.deviceType
		ExecuteCommands(p.Commands, initVariables, &(p.RelatedPanels))
		return p.QueryDevice()
	}
	return []string{}
}

func (p *PanelAction) QueryDevice() []string {
	var updatedIds []string = []string{}

	for i := 0; i < len(p.RelatedPanels); i++ {
		rpstr := strings.TrimSpace(p.RelatedPanels[i])
		parts := strings.Split(rpstr, " ")
		if len(parts) == 3 {
			var pt PanelTypes = Unknown
			if parts[0] == "Switch" {
				pt = Switch
			}
			if parts[0] == "Shading" {
				pt = Shading
			}
			if parts[0] == "Script" {
				pt = Script
			}
			idval, err := strconv.Atoi(parts[2])
			if err == nil && pt != Unknown {
				updatedIds = append(updatedIds, UpdateFirstHwPanel(pt, parts[1], idval)...)
			}
		}
	}
	updatedIds = append(updatedIds, p.idStr)
	return updatedIds
}

func (p *PanelAction) SetHwDeviceId(id int) {

}

func (p *PanelAction) RefreshHwStateIfMatch(fromPanelType PanelTypes, fromDeviceIp string, fromInDeviceId int, fromScriptName string, State int, InputState int) string {
	return p.idStr
}

type RunContext struct {
	variables   map[string]string
	iwblocks    IntStack
	whileblocks IntStack
}

func ExecuteCommands(program string, contextVariables map[string]string, relatedPanels *[]string) string {
	returnValue := ""
	cmds := strings.Split(program, "\n")
	ip := 0
	cmdCount := len(cmds)
	var ctx RunContext = RunContext{map[string]string{}, []int{}, []int{}}
	ctx.variables = contextVariables
	AddBaseValiables(&ctx)

	for ip < cmdCount {
		cmd := strings.TrimSpace(cmds[ip])

		if strings.HasPrefix(cmd, "//") {
			ip++
			continue
		}

		if strings.HasPrefix(cmd, "If ") {
			Command_If(&ctx, cmd[3:])
			ip++
			continue
		}
		if cmd == "Else" {
			Command_Else(&ctx)
			ip++
			continue
		}
		if cmd == "EndIf" {
			Command_EndIf(&ctx)
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "While ") {
			Command_While(&ctx, cmd[6:], ip)
			ip++
			continue
		}
		if cmd == "EndWhile" {
			Command_EndWhile(&ctx, &ip)
			ip++
			continue
		}

		if ctx.iwblocks.TopDef(1) == 0 {
			ip++
			continue
		}

		if cmd == "Return" {
			returnValue = ""
			break
		}

		if strings.HasPrefix(cmd, "Return ") {
			returnValue = ResolveVariables(ctx, cmd[7:])
			break
		}

		if strings.HasPrefix(cmd, "RelatedPanel ") {
			*relatedPanels = append(*relatedPanels, ResolveVariables(ctx, cmd[13:]))
			ip++
			continue
		}

		if strings.HasPrefix(cmd, "Run ") {
			Command_Run(ctx, cmd[4:], relatedPanels)
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "RunSet ") {
			Command_RunSet(&ctx, cmd[7:], relatedPanels)
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "PrintConsole ") {
			Command_PrintConsole(ctx, cmd[13:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "AddTo ") {
			Command_AddTo(&ctx, cmd[6:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "SubFrom ") {
			Command_SubFrom(&ctx, cmd[8:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "MulWith ") {
			Command_MulWith(&ctx, cmd[8:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "DivWith ") {
			Command_DivWith(&ctx, cmd[8:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "WaitMs ") {
			Command_WaitMs(ctx, cmd[7:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "Set ") {
			Command_Set(&ctx, cmd[4:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "CallHttp ") {
			Command_CallHttp(&ctx, cmd[9:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "SetFromJsonReq ") {
			Command_SetFromJsonReq(&ctx, cmd[15:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "LoadVariablesFromPanelId ") {
			Command_LoadVariablesFromPanelId(&ctx, cmd[25:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "LoadVariablesFromPanelIdWithPrefix ") {
			Command_LoadVariablesFromPanelIdWithPrefix(&ctx, cmd[35:])
			ip++
			continue
		}
		if cmd == "PrintVariablesConsole" {
			Command_PrintVariablesConsole(ctx)
			ip++
			continue
		}

		ip++
	}
	return returnValue
}

func ResolveVariables(ctx RunContext, str string) string {
	rstr := str
	if strings.Contains(str, "{{") && strings.Contains(str, "}}") {
		for name, value := range ctx.variables {
			rstr = strings.Replace(rstr, "{{"+name+"}}", value, -1)
		}
	}
	return rstr
}

func EvalExpressionBool(ctx RunContext, cmdpart string) bool {
	parts := strings.Split(cmdpart, " ")
	if len(parts) != 3 {
		return false //not correct expression
	}
	b := false
	if parts[1] == "==" ||
		parts[1] == "!=" ||
		parts[1] == "<" ||
		parts[1] == ">" ||
		parts[1] == "<=" ||
		parts[1] == ">=" {
		v1, err1 := strconv.ParseFloat(ResolveVariables(ctx, parts[0]), 32)
		v2, err2 := strconv.ParseFloat(ResolveVariables(ctx, parts[2]), 32)
		if err1 != nil || err2 != nil {
			return false //not correct expression
		}

		if parts[1] == "==" && v1 == v2 {
			return v1 == v2
		}
		if parts[1] == "!=" && v1 != v2 {
			return v1 != v2
		}
		if parts[1] == "<" && v1 < v2 {
			return v1 < v2
		}
		if parts[1] == ">" && v1 > v2 {
			return v1 > v2
		}
		if parts[1] == "<=" && v1 <= v2 {
			return v1 <= v2
		}
		if parts[1] == ">=" && v1 >= v2 {
			return v1 >= v2
		}
		return false //error in expression
	}
	if parts[1] == "eq" ||
		parts[1] == "neq" ||
		parts[1] == "in" ||
		parts[1] == "nin" {
		v1 := ResolveVariables(ctx, parts[0])
		v2 := ResolveVariables(ctx, parts[2])

		if parts[1] == "eq" {
			return v1 == v2
		}
		if parts[1] == "neq" {
			return v1 != v2
		}

		if parts[1] == "in" ||
			parts[1] == "nin" {
			v2parts := strings.Split(v2, ",")
			if parts[1] == "in" {
				return Contains(v2parts, v1)
			}
			if parts[1] == "nin" {
				return !Contains(v2parts, v1)
			}
		}
	}

	return b
}

func Command_If(ctx *RunContext, cmdpart string) {
	if !ctx.iwblocks.IsEmpty() && ctx.iwblocks.Top() == 0 {
		ctx.iwblocks.Push(0)
		return
	}
	pushval := 0
	if EvalExpressionBool(*ctx, cmdpart) {
		pushval = 1
	}
	ctx.iwblocks.Push(pushval)
}

func Command_Else(ctx *RunContext) {
	if ctx.iwblocks.IsEmpty() {
		return
	}
	topval := ctx.iwblocks.Pop()
	if topval == 0 {
		topval = 1
	} else {
		topval = 0
	}
	ctx.iwblocks.Push(topval)
}

func Command_EndIf(ctx *RunContext) {
	ctx.iwblocks.Pop()
}

func Command_While(ctx *RunContext, cmdpart string, ip int) {
	if !ctx.iwblocks.IsEmpty() && ctx.iwblocks.Top() == 0 {
		ctx.iwblocks.Push(0)
		return
	}
	pushval := 0
	if EvalExpressionBool(*ctx, cmdpart) {
		pushval = 1
	}
	ctx.iwblocks.Push(pushval)
	ctx.whileblocks.Push(ip)
}

func Command_EndWhile(ctx *RunContext, ip *int) {
	if ctx.iwblocks.TopDef(1) == 1 {
		*ip = ctx.whileblocks.Pop() - 1
	}
	ctx.iwblocks.Pop()
}

func Command_Run(ctx RunContext, cmdpart string, relatedPanels *[]string) {
	code, ok := ProgramLibrary[cmdpart]
	if ok {
		ExecuteCommands(code, ctx.variables, relatedPanels)
	}
}

func Command_RunSet(ctx *RunContext, cmdpart string, relatedPanels *[]string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 2 {
		code, ok := ProgramLibrary[parts[1]]
		if ok {
			result := ExecuteCommands(code, ctx.variables, relatedPanels)
			ctx.variables[parts[0]] = result
		}
	}
}

func Command_PrintConsole(ctx RunContext, cmdpart string) {
	fmt.Println("ACTION-CONSOLE> " + ResolveVariables(ctx, cmdpart))
}

func Command_PrintVariablesConsole(ctx RunContext) {
	for n, v := range ctx.variables {
		fmt.Println("ACTION-CONSOLE> " + n + " = " + v)
	}
}

func Command_AddTo(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 2 {
		fvv, err1 := strconv.ParseFloat(ctx.variables[parts[0]], 32)
		fav, err2 := strconv.ParseFloat(ResolveVariables(*ctx, parts[1]), 32)

		if err1 == nil && err2 == nil {
			ctx.variables[parts[0]] = fmt.Sprintf("%f", fvv+fav)
		}
	}
}

func Command_SubFrom(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 2 {
		fvv, err1 := strconv.ParseFloat(ctx.variables[parts[0]], 32)
		fav, err2 := strconv.ParseFloat(ResolveVariables(*ctx, parts[1]), 32)

		if err1 == nil && err2 == nil {
			ctx.variables[parts[0]] = fmt.Sprintf("%f", fvv-fav)
		}
	}
}

func Command_MulWith(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 2 {
		fvv, err1 := strconv.ParseFloat(ctx.variables[parts[0]], 32)
		fav, err2 := strconv.ParseFloat(ResolveVariables(*ctx, parts[1]), 32)

		if err1 == nil && err2 == nil {
			ctx.variables[parts[0]] = fmt.Sprintf("%f", fvv*fav)
		}
	}
}

func Command_DivWith(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 2 {
		fvv, err1 := strconv.ParseFloat(ctx.variables[parts[0]], 32)
		fav, err2 := strconv.ParseFloat(ResolveVariables(*ctx, parts[1]), 32)

		if err1 == nil && err2 == nil && fav != 0 {
			ctx.variables[parts[0]] = fmt.Sprintf("%f", fvv/fav)
		}
	}
}

func Command_WaitMs(ctx RunContext, cmdpart string) {
	msval, err := strconv.Atoi(ResolveVariables(ctx, cmdpart))
	if err == nil {
		time.Sleep(time.Millisecond * time.Duration(msval))
	}
}

func Command_Set(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 2 {
		ctx.variables[parts[0]] = ResolveVariables(*ctx, parts[1])
	}
}

func Command_SetFromJsonReq(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 3 {
		jhq := execJsonHttpQuery(ResolveVariables(*ctx, parts[1]))
		if jhq.Success {
			ctx.variables["LastHttpCallSuccess"] = "true"
		} else {
			ctx.variables["LastHttpCallSuccess"] = "false"
		}

		_, typestr := jhq.SmartJSON.GetNodeByPath(parts[2])
		if typestr == "string" {
			ctx.variables[parts[0]] = jhq.SmartJSON.GetStringByPathWithDefault(parts[2], "")
		}
		if typestr == "float64" {
			ctx.variables[parts[0]] = fmt.Sprintf("%f", jhq.SmartJSON.GetFloat64ByPathWithDefault(parts[2], 0.0))
		}
		if typestr == "int" {
			ctx.variables[parts[0]] = fmt.Sprintf("%d", jhq.SmartJSON.GetIntegerByPathWithDefault(parts[2], 0))
		}
		if typestr == "bool" {
			b, _ := jhq.SmartJSON.GetBoolByPath(parts[2])
			if b {
				ctx.variables[parts[0]] = "true"
			} else {
				ctx.variables[parts[0]] = "false"
			}
		}
	}
}

func Command_CallHttp(ctx *RunContext, cmdpart string) {
	ro := execJsonHttpQuery(ResolveVariables(*ctx, cmdpart))
	if ro.Success {
		ctx.variables["LastHttpCallSuccess"] = "true"
	} else {
		ctx.variables["LastHttpCallSuccess"] = "false"
	}
}

func AddBaseValiables(ctx *RunContext) {
	now := time.Now()
	ctx.variables["Time.Hour"] = fmt.Sprintf("%02d", now.Hour())
	ctx.variables["Time.Minute"] = fmt.Sprintf("%02d", now.Minute())
	ctx.variables["Time.Second"] = fmt.Sprintf("%02d", now.Second())
	ctx.variables["Time.SecOfDay"] = fmt.Sprintf("%d", now.Hour()*3600+now.Minute()*60+now.Second())
	ctx.variables["Time.WeekDay"] = fmt.Sprintf("%d", now.Weekday())
	ctx.variables["Time.Month"] = fmt.Sprintf("%d", now.Month())
	ctx.variables["Time.Day"] = fmt.Sprintf("%d", now.Day())
	ctx.variables["Time.Year"] = fmt.Sprintf("%d", now.Year())
	ctx.variables["Time.YearDay"] = fmt.Sprintf("%d", now.YearDay())
}

func Command_LoadVariablesFromPanelId(ctx *RunContext, cmdpart string) {
	pc := len(Panels)
	for i := 0; i < pc; i++ {
		if strings.TrimSpace(cmdpart) == Panels[i].IdStr() {
			merges := Panels[i].ExposeVariables()
			for n, v := range merges {
				ctx.variables[n] = v
			}
		}
	}
}

func Command_LoadVariablesFromPanelIdWithPrefix(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 2 {
		pc := len(Panels)
		for i := 0; i < pc; i++ {
			if strings.TrimSpace(parts[1]) == Panels[i].IdStr() {
				merges := Panels[i].ExposeVariables()
				for n, v := range merges {
					ctx.variables[parts[0]+n] = v
				}
			}
		}
	}
}

func (p PanelAction) ExposeVariables() map[string]string {

	var m map[string]string = map[string]string{}

	m["Panel.Id"] = p.idStr
	m["Panel.Title"] = p.title
	m["Panel.DeviceType"] = p.deviceType
	m["Panel.SubPage"] = p.subPage
	m["Panel.Index"] = fmt.Sprintf("%d", p.index)

	m["Panel.PowerInfo"] = "false"
	return m
}
