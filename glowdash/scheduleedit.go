/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hyper-prog/smartyaml"
)

type PageScheduleEdit struct {
	PageBase
}

func NewPageScheduleEdit() *PageScheduleEdit {
	return &PageScheduleEdit{
		PageBase{
			idStr:      "",
			pageType:   ScheduleEdit,
			title:      "",
			deviceType: "",
			index:      0,
		},
	}
}

func (p *PageScheduleEdit) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {
	if p.title == "" {
		p.title = "Glowdash schedules"
	}
}

func subActionCodeToDisplay(code string) string {
	if code == "" {
		return "Nothing"
	}
	dtext, found := subActionDisplayText[code]
	if found {
		return dtext
	}
	return code + " C"
}

func ProcessScheduleForm(r *http.Request) string {

	mode := ""
	if r.Form.Get("sdlsubmit") != "Add schedule" && r.Form.Get("sdlsubmit") != "Save schedule" {
		return mode
	}

	if r.Form.Get("sdlsubmit") == "Add schedule" {
		mode = "n"
	}
	if r.Form.Get("sdlsubmit") == "Save schedule" {
		mode = "e"
	}

	index := 0
	if mode == "e" {
		if len(r.Form.Get("index")) < 1 {
			return mode
		}
		idx, erri := strconv.Atoi(r.Form.Get("index"))
		if erri != nil {
			return mode
		}
		index = idx
	}

	if len(r.Form.Get("action")) < 1 {
		return mode
	}

	s := Schedule{}

	formname := r.Form.Get("sdlname")
	if len(formname) < 1 {
		t := time.Now()
		formname = fmt.Sprintf("Unnamed schedule - %s", t.Format("15:04:05"))
	}
	s.name = formname
	if mode == "n" {
		for n := 1; getScheduleIndex(s.name) >= 0; n++ {
			s.name = fmt.Sprintf("%s (%d)", formname, n)
		}
	}

	s.enabled = false
	if r.Form.Get("sdlenabled") == "on" {
		s.enabled = true
	}

	s.oneshot = false
	if r.Form.Get("sdloneshot") == "yes" {
		s.oneshot = true
		s.enabled = true
	}

	hour, errh := strconv.Atoi(r.Form.Get("hour"))
	if errh == nil {
		s.hour = hour
	}

	min, errh := strconv.Atoi(r.Form.Get("min"))
	if errh == nil {
		s.min = min
	}

	s.dayMon = false
	if r.Form.Get("mon") == "on" {
		s.dayMon = true
	}
	s.dayTue = false
	if r.Form.Get("tue") == "on" {
		s.dayTue = true
	}
	s.dayWed = false
	if r.Form.Get("wed") == "on" {
		s.dayWed = true
	}
	s.dayThu = false
	if r.Form.Get("thu") == "on" {
		s.dayThu = true
	}
	s.dayFri = false
	if r.Form.Get("fri") == "on" {
		s.dayFri = true
	}
	s.daySat = false
	if r.Form.Get("sat") == "on" {
		s.daySat = true
	}
	s.daySun = false
	if r.Form.Get("sun") == "on" {
		s.daySun = true
	}

	msel_str := r.Form.Get("action")
	msel_parts := strings.Split(msel_str, ":")
	if len(msel_parts) == 2 {
		s.actionType = msel_parts[0]
		s.actionId = msel_parts[1]
		s.actionParam = r.Form.Get("subaction")
	} else {
		return mode
	}

	if mode == "n" {
		addSchedule(s)
	}
	if mode == "e" {
		updateSchedule(index, s)
	}
	return mode
}

func (p PageScheduleEdit) PageHtml(withContainer bool, r *http.Request) string {
	html := ""
	formmode := ProcessScheduleForm(r)
	html += "<div class=\"schedule-edit-page\">"
	html += "<h3>" + p.title + "</h3>"
	html += "<div id=\"scheduleedit-main-list-container\" class=\"schedule-list\">"

	cs := countSchedules()

	for i := 0; i < cs; i++ {
		sdl := getScheduleByIndex(i)
		html += "<div class=\"schedule-item-show\" id=\"sdl-index-" + fmt.Sprintf("%d", i) + "\">"
		html += htmlStaticScheduleBlock(i, cs, sdl)
		html += "</div>" //.schedule-item-show
	}
	html += "</div>" //.schedule-list
	html += "<div id=\"schedule-add-edit-block\">"
	html += "</div>"

	html += "<button id=\"schedule-edit-add\" class=\"jsaction scheduleedit-ctrl-button\"><i class=\"fa fa-add3\"></i></button>"
	html += "<button id=\"schedule-edit-addoneshot\" class=\"jsaction scheduleedit-ctrl-button\"><i class=\"fa fa-gun\"></i></button>"
	if formmode != "" {
		html += "<script>setTimeout(\"window.location = '/page/schedpage';\",200);</script>"
	}

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"fullpage-content\" tabindex=\"-1\">", p.IdStr()) +
			html + "</div>"
	}

	return html
}

func htmlStaticScheduleBlock(index int, cS int, s Schedule) string {
	html := "<div class=\"schedule-item-name\">" + s.name + "</div>"
	html += "<div class=\"schedule-item-datas\">"
	if s.oneshot {
		html += "<div class=\"schedule-item-ena\"><img src=\"/static/one-shot.png\"/></div>"
	} else {
		if s.enabled {
			html += "<div class=\"schedule-item-ena\"><span class=\"act-show-ena-on\">ON</span></div>"
		} else {
			html += "<div class=\"schedule-item-ena\"><span class=\"act-show-ena-off\">OFF</span></div>"
		}
	}

	html += "<div class=\"schedule-item-time\">" + fmt.Sprintf("%02d:%02d", s.hour, s.min) + "</div>"

	html += htmlScheduleDays(s, "short", true)

	title := "-"
	refPanel := GetPanelById(s.actionId)
	if refPanel != nil {
		title = refPanel.EventTitle()
	}
	html += "<div class=\"schedule-item-act\">"
	html += "<span class=\"schedule-item-act-id\">" + title + "</span>"
	html += "<span class=\"schedule-item-act-sep\"><i class=\"fa fa-rightarrow\"></i></span>"
	html += "<span class=\"schedule-item-act-param\">" + subActionCodeToDisplay(s.actionParam) + "</span>"
	html += "<br/>"
	html += "<span class=\"schedule-item-lastrun\">Last run on: " + s.lastrun + "</span>"
	html += "</div>" //.schedule-item-act

	html += "<div class=\"schedule-item-func\"><table>"

	html += "<tr>"
	html += "<td><button class=\"jsaction scheduleedit-ctrl-button\" id=\"act-schedule-up-idx-" + fmt.Sprintf("%d", index) + "\" " + IfTrue(index == 0, "disabled") + ">" +
		"<i class=\"fa fa-smallup\"></i></button></td>"
	html += "<td><button class=\"jsaction scheduleedit-ctrl-button\" id=\"act-schedule-edit-idx-" + fmt.Sprintf("%d", index) + "\">" +
		"<i class=\"fa fa-edit3\"></i></button></td>"
	html += "</tr>"
	html += "<tr>"
	html += "<td><button class=\"jsaction scheduleedit-ctrl-button\" id=\"act-schedule-down-idx-" + fmt.Sprintf("%d", index) + "\"" + IfTrue(index == cS-1, "disabled") + ">" +
		"<i class=\"fa fa-smalldown\"></i></button></td>"
	html += "<td><button class=\"jsaction scheduleedit-ctrl-button\" id=\"act-schedule-delete-idx-" + fmt.Sprintf("%d", index) + "\">" +
		"<i class=\"fa fa-del3\"></i></button></td>"
	html += "</tr>"
	html += "</table></div>" //.schedule-item-func

	html += "</div>" //.schedule-item-datas
	return html
}

func htmlDeleteConfirmation(index int, s Schedule) string {
	html := "<div class=\"schedule-item-name\">" + s.name + "</div>"
	html += "<div>Do you really want to delete this schedule?</div>"

	html += "<table><tr>"
	html += "<td><button class=\"jsaction scheduleedit-ctrl-button\" id=\"act-schedule-real-delete-idx-" + fmt.Sprintf("%d", index) + "\">Yes, delete schedule</button></td>"
	html += "<td><button class=\"jsaction scheduleedit-ctrl-button\" id=\"act-schedule-backstatic-idx-" + fmt.Sprintf("%d", index) + "\">No, cancel</button></td>"
	html += "</tr></table>"

	return html
}

func htmlScheduleEditor(new bool, oneshotIfNew bool, s Schedule) string {
	html := "<div class=\"schedule-item\"><form method=\"post\" enctype=\"application/x-www-form-urlencoded\">"

	current_time := time.Now()
	if new {
		if oneshotIfNew {
			html += "<span style=\"font-weight: strong; font-size: larger; padding: 5px;\">New one shot schedule</span>"
		} else {
			html += "<span style=\"font-weight: strong; font-size: larger; padding: 5px;\">New schedule</span>"
		}
	} else {
		html += "<span style=\"font-weight: strong; font-size: larger; padding: 5px;\">Edit schedule</span>"
	}

	name_string := ""
	if !new {
		name_string = s.name
	}

	mon_checked := "checked"
	tue_checked := "checked"
	wed_checked := "checked"
	thu_checked := "checked"
	fri_checked := "checked"
	sat_checked := "checked"
	sun_checked := "checked"
	if !new {
		if !s.dayMon {
			mon_checked = ""
		}
		if !s.dayTue {
			tue_checked = ""
		}
		if !s.dayWed {
			wed_checked = ""
		}
		if !s.dayThu {
			thu_checked = ""
		}
		if !s.dayFri {
			fri_checked = ""
		}
		if !s.daySat {
			sat_checked = ""
		}
		if !s.daySun {
			sun_checked = ""
		}
	}

	html += "<div class=\"schedule-data-area\">"

	html += "<div class=\"schedule-data-block\">"
	html += "<div class=\"schedule-data-item-desc\">Enabled / Name</div>"
	html += "<div class=\"schedule-data-item-value\">"
	if !new {
		html += "  <input type=\"hidden\" name=\"index\" value=\"" + fmt.Sprintf("%d", getScheduleIndex(s.name)) + "\" />"
	}

	if (new && oneshotIfNew) || (!new && s.oneshot) {
		html += "<input type=\"hidden\" name=\"sdlenabled\" value=\"on\" />"
		html += "<input type=\"hidden\" name=\"sdloneshot\" value=\"yes\" />"
		html += "<img src=\"/static/one-shot.png\"/><br/>"

	} else {
		html += "<input type=\"hidden\" name=\"sdloneshot\" value=\"no\" />"
		html += htmlOnOffSlider("sdlenabled", new || s.enabled, "") + "<br/>"
	}

	html += "  <input type=\"text\" name=\"sdlname\" value=\"" + name_string + "\"/><br/>"
	html += "</div>"
	html += "</div>"

	html += "<div class=\"schedule-data-block\">"
	html += "<div class=\"schedule-data-item-desc\">Running time</div>"
	html += "<div class=\"schedule-data-item-value\">"
	if new {
		html += htmlClockPicker("newsch", current_time.Hour(), current_time.Minute(), true, "", "none")
	} else {
		html += htmlClockPicker("editsch", s.hour, s.min, true, "", "none")
	}
	html += "</div>"
	html += "</div>"

	html += "<div class=\"schedule-data-block\">"
	html += "<div class=\"schedule-data-item-desc\">Running days</div>"
	html += "<div class=\"schedule-data-item-value\">"
	html += `Mon<input type="checkbox" name="mon" ` + mon_checked + `/>
			 Tue<input type="checkbox" name="tue" ` + tue_checked + `/>
			 Wed<input type="checkbox" name="wed" ` + wed_checked + `/>
			 Thu<input type="checkbox" name="thu" ` + thu_checked + `/><br/>
			 Fri<input type="checkbox" name="fri" ` + fri_checked + `/>
			 Sat<input type="checkbox" name="sat" ` + sat_checked + `/>
			 Sun<input type="checkbox" name="sun" ` + sun_checked + `/>`
	html += "</div>"
	html += "</div>"

	html += "<div class=\"schedule-data-block\">"
	html += "<div class=\"schedule-data-item-desc\">Action on Time</div>"
	html += "<div class=\"schedule-data-item-value\">"
	html += "<select name=\"action\" class=\"schedule-action-selector\" data-actionsubid=\"sch-sub-sel-one\">"
	panelcnt := len(Panels)
	subselOpts := ""
	for i := 0; i < panelcnt; i++ {
		if strings.HasPrefix(Panels[i].IdStr(), "autogenId") {
			continue
		}

		current := false
		selectedText := ""
		if (!new && s.actionId == Panels[i].IdStr()) || (new && i == 0) {
			selectedText = "selected"
			current = true
		}

		if Panels[i].PanelType() == Switch {
			html += "<option value=\"switch:" + Panels[i].IdStr() + "\" " + selectedText + ">" + Panels[i].EventTitle() + "</option>"

			if current {
				subselOpts += "<option value=\"on\" " + IfTrue(s.actionParam == "on", "selected") + ">Switch On</option>"
				subselOpts += "<option value=\"off\" " + IfTrue(s.actionParam == "off", "selected") + ">Switch Off</option>"
			}
		}
		if Panels[i].PanelType() == Shading {
			html += "<option value=\"shading:" + Panels[i].IdStr() + "\" " + selectedText + ">" + Panels[i].EventTitle() + "</option>"

			if current {
				subselOpts += "<option value=\"open\" " + IfTrue(s.actionParam == "open", "selected") + ">Open</option>"
				subselOpts += "<option value=\"close\" " + IfTrue(s.actionParam == "close", "selected") + ">Close</option>"
			}
		}
		if Panels[i].PanelType() == Action {
			html += "<option value=\"action:" + Panels[i].IdStr() + "\" " + selectedText + ">" + Panels[i].EventTitle() + "</option>"
			if current {
				subselOpts += "<option value=\"run\" " + IfTrue(s.actionParam == "run", "selected") + ">Run</option>"
			}
		}
		if Panels[i].PanelType() == Script {
			html += "<option value=\"script:" + Panels[i].IdStr() + "\" " + selectedText + ">" + Panels[i].EventTitle() + "</option>"

			if current {
				subselOpts += "<option value=\"start\" " + IfTrue(s.actionParam == "start", "selected") + ">Start</option>"
				subselOpts += "<option value=\"stop\" " + IfTrue(s.actionParam == "stop", "selected") + ">Stop</option>"
			}
		}
		if Panels[i].PanelType() == Thermostat {
			html += "<option value=\"therm:" + Panels[i].IdStr() + "\" " + selectedText + ">" + Panels[i].EventTitle() + "</option>"

			if current {
				apf, _ := strconv.ParseFloat(s.actionParam, 8)
				for t := 5.0; t <= 30; t += 0.5 {
					subselOpts += "<option value=\"" + fmt.Sprintf("%.1f", t) + "\" " + IfTrue(t == apf, "selected") + ">" + fmt.Sprintf("%.1f", t) + "</option>"
				}
			}
		}
	}
	html += "</select>"
	html += "<br/>"
	html += "&nbsp;<i class=\"fa fa-rightarrow\"></i>"
	html += "<br/>"
	html += "<select id=\"sch-sub-sel-one\" name=\"subaction\">"
	html += subselOpts
	html += "</select>"
	html += "</div>"
	html += "</div>"

	html += "<div class=\"schedule-data-block\">"
	html += "<div class=\"schedule-data-item-desc\">Operation</div>"
	html += "<div class=\"schedule-data-item-value\">"
	if new {
		html += "  <input type=\"submit\" name=\"sdlsubmit\" value=\"Add schedule\" class=\"schedule-submit-button\" />"
	} else {
		html += "  <input type=\"submit\" name=\"sdlsubmit\" value=\"Save schedule\" class=\"schedule-submit-button\" />"
	}
	html += "<br/>"
	html += "  <input type=\"submit\" name=\"sdlsubmit\" value=\"Cancel\" class=\"schedule-submit-button\" />"
	html += "</div>"
	html += "</div>"

	html += "</div>" //.schedule-data-area

	html += "</form></div>"
	return html
}

func (p PageScheduleEdit) IsActionIdMatch(aId string) bool {
	if aId == "schedule-edit-add" {
		return true
	}
	if aId == "schedule-edit-addoneshot" {
		return true
	}
	if strings.HasPrefix(aId, "act-schedule-edit-idx-") {
		return true
	}
	if strings.HasPrefix(aId, "act-schedule-delete-idx-") {
		return true
	}
	if strings.HasPrefix(aId, "act-schedule-real-delete-idx-") {
		return true
	}
	if strings.HasPrefix(aId, "act-schedule-backstatic-idx-") {
		return true
	}
	if strings.HasPrefix(aId, "act-schedule-up-idx-") {
		return true
	}
	if strings.HasPrefix(aId, "act-schedule-down-idx-") {
		return true
	}
	return false
}

func (p PageScheduleEdit) HandleActionEvent(res *ActionResponse, actionName string, parameters map[string]string) {

	if actionName == "schedule-edit-add" {
		res.addCommandArg2("sethtml", "#schedule-add-edit-block", htmlScheduleEditor(true, false, nullSchedule()))
		res.setResultString("ok")
	}
	if actionName == "schedule-edit-addoneshot" {
		res.addCommandArg2("sethtml", "#schedule-add-edit-block", htmlScheduleEditor(true, true, nullSchedule()))
		res.setResultString("ok")
	}
	if strings.HasPrefix(actionName, "act-schedule-edit-idx-") {
		idxstr := actionName[22:]
		idx, err := strconv.Atoi(idxstr)
		if err == nil {
			res.addCommandArg2("sethtml", "#sdl-index-"+idxstr, htmlScheduleEditor(false, false, getScheduleByIndex(idx)))
		}
		res.setResultString("ok")
	}
	if strings.HasPrefix(actionName, "act-schedule-delete-idx-") {
		idxstr := actionName[24:]
		idx, err := strconv.Atoi(idxstr)
		if err == nil {
			res.addCommandArg2("sethtml", "#sdl-index-"+idxstr, htmlDeleteConfirmation(idx, getScheduleByIndex(idx)))
		}
		res.setResultString("ok")
	}
	if strings.HasPrefix(actionName, "act-schedule-real-delete-idx-") {
		idxstr := actionName[29:]
		idx, err := strconv.Atoi(idxstr)
		if err == nil {
			removeSchedule(idx)
			res.addCommandArg2("sethtml", "#sdl-index-"+idxstr, "")
		}
		res.setResultString("ok")
	}
	if strings.HasPrefix(actionName, "act-schedule-backstatic-idx-") {
		idxstr := actionName[28:]
		idx, err := strconv.Atoi(idxstr)
		if err == nil {
			res.addCommandArg2("sethtml", "#sdl-index-"+idxstr, htmlStaticScheduleBlock(idx, countSchedules(), getScheduleByIndex(idx)))
		}
		res.setResultString("ok")
	}

	if strings.HasPrefix(actionName, "act-schedule-up-idx-") {
		idxstr := actionName[20:]
		idx, err := strconv.Atoi(idxstr)
		if err == nil {
			scheduleMoveUp(idx)
			res.addCommandArg0("refreshpage")
		}
		res.setResultString("ok")
	}
	if strings.HasPrefix(actionName, "act-schedule-down-idx-") {
		idxstr := actionName[22:]
		idx, err := strconv.Atoi(idxstr)
		if err == nil {
			scheduleMoveDown(idx)
			res.addCommandArg0("refreshpage")
		}
		res.setResultString("ok")
	}
}
