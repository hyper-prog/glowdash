/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2026 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type RunContext struct {
	variables    map[string]string
	jqrvariables map[string]JsonHttpQuery
	iwblocks     IntStack
	whileblocks  IntStack
}

func ExecuteCommands(program string, contextVariables map[string]string, relatedPanels *[]string) map[string]string {
	returnValues := map[string]string{}
	returnValues["Return"] = ""
	cmds := strings.Split(program, "\n")
	ip := 0
	cmdCount := len(cmds)
	var ctx RunContext = RunContext{map[string]string{}, map[string]JsonHttpQuery{}, []int{}, []int{}}
	ctx.variables = contextVariables
	AddBaseValiables(&ctx)

	for ip < cmdCount {
		cmd := strings.TrimSpace(cmds[ip])

		if cmd == "" {
			ip++
			continue
		}

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

		if ctx.iwblocks.HasZeroElement() {
			ip++
			continue
		}

		if cmd == "Return" {
			returnValues["Return"] = ""
			break
		}

		if strings.HasPrefix(cmd, "Return ") {
			returnValues["Return"] = ResolveVariables(ctx, cmd[7:])
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
		if strings.HasPrefix(cmd, "ModWith ") {
			Command_ModWith(&ctx, cmd[8:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "RoundDown ") {
			Command_Round(&ctx, cmd[10:], 0)
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "RoundUp ") {
			Command_Round(&ctx, cmd[8:], 1)
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "RoundMath ") {
			Command_Round(&ctx, cmd[10:], 2)
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "AddMinutesToTime ") {
			Command_AddMinutesToTime(&ctx, cmd[17:])
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
		if strings.HasPrefix(cmd, "CallHttpStoreJson ") {
			Command_CallHttpStoreJson(&ctx, cmd[18:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "SetFromJsonReq ") {
			Command_SetFromJsonReq(&ctx, cmd[15:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "SetFromStoredJson ") {
			Command_SetFromStoredJson(&ctx, cmd[18:])
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
		if strings.HasPrefix(cmd, "SetSchedule ") {
			Command_SetSchedule(&ctx, cmd[12:])
			ip++
			continue
		}
		if strings.HasPrefix(cmd, "AddOneshotSchedule ") {
			Command_AddOneshotSchedule(&ctx, cmd[19:])
			ip++
			continue
		}
		if cmd == "PrintVariablesConsole" {
			Command_PrintVariablesConsole(ctx)
			ip++
			continue
		}

		fmt.Println("--------- Script error---------\nUnknown command: ", cmd)

		ip++
	}

	for n, v := range ctx.variables {
		if strings.HasPrefix(n, "Return.") {
			returnValues[n] = v
		}
	}

	return returnValues
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

	if parts[1] == "booleq" {
		v1 := strings.TrimSpace(ResolveVariables(ctx, parts[0]))
		v2 := strings.TrimSpace(ResolveVariables(ctx, parts[2]))
		if v1 == "1" || v1 == "t" || v1 == "yes" || v1 == "on" || v1 == "y" {
			v1 = "true"
		}
		if v1 == "0" || v1 == "" || v1 == "no" || v1 == "off" || v1 == "n" {
			v1 = "false"
		}

		if v2 == "1" || v2 == "t" || v2 == "yes" || v2 == "on" || v2 == "y" {
			v2 = "true"
		}
		if v2 == "0" || v2 == "" || v2 == "no" || v2 == "off" || v2 == "n" {
			v2 = "false"
		}

		if v1 == v2 {
			return true
		}
		return false
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
			results := ExecuteCommands(code, ctx.variables, relatedPanels)
			ctx.variables[parts[0]] = results["Return"]
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

func Command_ModWith(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 2 {
		fvv, err1 := strconv.ParseFloat(ctx.variables[parts[0]], 32)
		fav, err2 := strconv.ParseFloat(ResolveVariables(*ctx, parts[1]), 32)

		if err1 == nil && err2 == nil {
			ctx.variables[parts[0]] = fmt.Sprintf("%d", int(fvv)%int(fav))
		}
	}
}

func Command_Round(ctx *RunContext, cmdpart string, mode int) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 1 {
		fvv, err1 := strconv.ParseFloat(ctx.variables[parts[0]], 32)

		if err1 == nil {
			if mode == 0 {
				ctx.variables[parts[0]] = fmt.Sprintf("%d", int(fvv))
			} else if mode == 1 {
				ctx.variables[parts[0]] = fmt.Sprintf("%d", int(fvv+0.9999))
			} else {
				ctx.variables[parts[0]] = fmt.Sprintf("%d", int(fvv+0.5))
			}
		}
	}
}

func Command_AddMinutesToTime(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) != 2 {
		ctx.variables[parts[0]] = "error"
		return
	}

	timestring, errnf := ctx.variables[parts[0]]
	if !errnf {
		ctx.variables[parts[0]] = "error"
		return
	}
	timeparts := strings.Split(timestring, ":")

	if len(timeparts) == 2 || len(timeparts) == 3 {
		hour, err1 := strconv.Atoi(timeparts[0])
		min, err2 := strconv.Atoi(timeparts[1])

		minutesToAdd, err3 := strconv.Atoi(ResolveVariables(*ctx, parts[1]))

		if err1 == nil && err2 == nil && err3 == nil {
			t := time.Date(2000, 1, 1, hour, min, 0, 0, time.UTC)
			t = t.Add(time.Duration(minutesToAdd) * time.Minute)
			ctx.variables[parts[0]] = fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
			return
		}
	}
	ctx.variables[parts[0]] = "error"
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

func Command_SetFromStoredJson(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 3 {
		jhq, ok := ctx.jqrvariables[parts[1]]
		if !ok {
			fmt.Println("--------- Script error---------\nUnknown json result variable: ", parts[1])
			return
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

func Command_SetSchedule(ctx *RunContext, cmdpart string) {
	rc := ResolveVariables(*ctx, cmdpart)
	if strings.HasPrefix(rc, "on ") {
		SetScheduleOnOffByName(rc[3:], true)
	}
	if strings.HasPrefix(rc, "off ") {
		SetScheduleOnOffByName(rc[4:], false)
	}
}

func Command_AddOneshotSchedule(ctx *RunContext, cmdpart string) {
	rc := ResolveVariables(*ctx, cmdpart)
	s := Schedule{}
	s.name = "Generated schedule on " + time.Now().Format("2006-01-02 15:04:05")
	s.enabled = true
	s.oneshot = true

	s.dayMon = true
	s.dayTue = true
	s.dayWed = true
	s.dayThu = true
	s.dayFri = true
	s.daySat = true
	s.daySun = true

	var panelId string
	var actionParam string
	var reqhour float64
	var reqmin float64

	n, err := fmt.Sscanf(rc, "%s %s %f:%f", &panelId, &actionParam, &reqhour, &reqmin)
	if err != nil || n != 4 {
		fmt.Println("--------- Script parse error---------\nError in AddOneshotSchedule parameters: ", rc)
		return
	}

	s.hour = int(reqhour)
	s.min = int(reqmin)

	s.actionType = getScheduleActionTypeByPanelId(panelId)
	if s.actionType != "" {
		s.actionId = panelId
		s.actionParam = actionParam
		addSchedule(s)
	} else {
		fmt.Println("--------- Script error---------\nUnknown panel id in AddOneshotSchedule: ", panelId)
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

func Command_CallHttpStoreJson(ctx *RunContext, cmdpart string) {
	parts := strings.Split(cmdpart, " ")
	if len(parts) == 2 {
		ctx.jqrvariables[parts[0]] = execJsonHttpQuery(ResolveVariables(*ctx, parts[1]))
		if ctx.jqrvariables[parts[0]].Success {
			ctx.variables["LastHttpCallSuccess"] = "true"
		} else {
			ctx.variables["LastHttpCallSuccess"] = "false"
		}
	}
}

func AddBaseValiables(ctx *RunContext) {
	now := time.Now()
	ctx.variables["Time.Hour"] = fmt.Sprintf("%02d", now.Hour())
	ctx.variables["Time.Minute"] = fmt.Sprintf("%02d", now.Minute())
	ctx.variables["Time.TimeHM"] = fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	ctx.variables["Time.TimeHMS"] = fmt.Sprintf("%02d:%02d:%02d", now.Hour(), now.Minute(), now.Second())
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

func getUpdatedIdsFromRelatedPanels(relatedPanels []string) []string {
	var updatedIds []string = []string{}

	for i := 0; i < len(relatedPanels); i++ {
		rpstr := strings.TrimSpace(relatedPanels[i])
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
	return updatedIds
}
