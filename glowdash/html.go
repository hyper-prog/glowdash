/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
	"net/http"
	"time"
)

func htmlStart() string {
	return `<!DOCTYPE html>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=0" />
	<meta name="version" content="1.0" />
	<html>
	<head>
		<link rel="icon" type="image/x-icon" href="/static/favicon.ico">` +
		"<link rel=\"stylesheet\" href=\"/static/glowdash.css?" + AssetVer + "\">" +
		"<link rel=\"stylesheet\" href=\"/static/breadcrumb.css?" + AssetVer + "\">" +
		"<script src=\"/static/glowdash.js?" + AssetVer + "\"></script>" +
		"<title>" + DashboardTitle + "</title>" +
		"</head>" +
		"<body>" +
		"<script>" +
		"conf_use_sse=" + fmt.Sprintf("%d", WebUseSSE) + ";\n" +
		"conf_sse_port='" + fmt.Sprintf("%d", WebSSEPort) + "';\n" +
		"</script>" +
		"<div id=\"overlay-pholder\"></div>"
}

func htmlEnd() string {
	return `</body>
	</html>`
}

func htmlHeaderLine(sub string) string {
	html := "<div class=\"header\">"
	now := time.Now()
	if ReadWindInfo {
		if int64(now.Sub(LastWindInfo.RequestTime)) > (1000000000 * WindInfoPollInterval) {
			if DebugLevel > 0 {
				fmt.Printf("Read wind info\n")
			}
			LastWindInfo = GetWindInfo()
		}
		html += "<div class=\"titleline\">" +
			fmt.Sprintf("<span class=\"titleandclock\">"+
				"<span id=\"mmmpname\">%s</span> - <span id=\"tlclock\">%02d:%02d</span>"+
				"<span class=\"pluscomma\">, </span></span>  "+
				"<span class=\"windinfosect\">Wind:%.1fkm/h, Gust: %.1fkm/h (%02d:%02d)</span>",
				DashboardTitle, now.Hour(), now.Minute(),
				LastWindInfo.Windspeed, LastWindInfo.GustSpeed,
				LastWindInfo.RequestTime.Hour(), LastWindInfo.RequestTime.Minute()) +
			"</div>"
	} else {
		html += "<div class=\"titleline\">" +
			fmt.Sprintf("<span class=\"titleandclock\">"+
				"<span id=\"mmmpname\">%s</span> - <span id=\"tlclock\">%02d:%02d</span>"+
				"</span>", DashboardTitle, now.Hour(), now.Minute()) +
			"</div>"
	}

	html += `<div class="breadcrumb" style="margin-left: 0; margin-right: auto;">
				<ul>
					<li><a href="/"><i class="fa fa-home breadcrumbhomeicon"></i></a></li>`
	if sub != "" {
		pc := len(Panels)
		for i := 0; i < pc; i++ {
			if Panels[i].SubTo() == sub {
				html += "<li><a href=\"/subpage/" + Panels[i].SubTo() + "\">" + Panels[i].Title() + "</a></li>"
			}
			if Panels[i].LaunchTo() == sub {
				html += "<li><a href=\"/page/" + Panels[i].LaunchTo() + "\">" + Panels[i].Title() + "</a></li>"
			}
		}
	}

	html += "</ul><div class=\"clearboth\"></div></div>" // .clearboth .breadcrumb
	html += "</div>"                                     // .header
	return html
}

func htmlPanels(sub string) string {
	html := "<div class=\"card-grid grid-gap-m\">"

	pcount := len(Panels)
	for i := 0; i < pcount; i++ {
		if Panels[i].IsHide() {
			continue
		}
		if sub == "" && Panels[i].Sub() != "" {
			continue
		}
		if sub != "" && Panels[i].Sub() != sub {
			continue
		}
		html += Panels[i].PanelHtml(true)
	}
	html += "</div>"
	return html
}

func htmlCustomPage(page string, r *http.Request) string {
	html := "<div class=\"fullpagecontent\">"
	pcount := len(Pages)
	for i := 0; i < pcount; i++ {
		if page == Pages[i].IdStr() {
			html += Pages[i].PageHtml(true, r)
		}
	}

	html += "</div>"
	return html
}

func htmlOnOffSlider(name string, value bool, extraclasses string) string {
	checkedstr := ""
	if value {
		checkedstr = " checked"
	}
	return "<label class=\"switch\">" +
		"<input type=\"checkbox\" name=\"" + name + "\" class=\"" + extraclasses + "\"" + checkedstr + ">" +
		"<span class=\"slider round\"></span>" +
		"</label>"
}

func htmlClockPicker(mainId string, hour int, min int, updownbuttons bool, extraclasses string, panelid string) string {
	h := "" +
		"<div id=\"" + mainId + "\" class=\"clockpicker-controller-block " + extraclasses + "\"" +
		" data-mainid=\"" + mainId + "\" data-pnlid=\"" + panelid + "\">" +
		"  <table>"
	if updownbuttons {
		h +=
			"<tr>" +
				"  <td><button id=\"ts-hour-up\" class=\"ts-ud clock-spinner dir-up type-hour\" data-sid=\"t-hour\"><i class=\"fa fa-tup\"></i></button></td>" +
				"  <td></td>" +
				"  <td><button id=\"ts-min-up\" class=\"ts-ud clock-spinner dir-up type-min\" data-sid=\"t-min\"><i class=\"fa fa-tup\"></i></button></td>" +
				"</tr>"
	}
	h += "<tr>" +
		"  <td><button id=\"t-hour\" class=\"v-hour\">" + fmt.Sprintf("%02d", hour) + "</button></td>" +
		"  <td style=\"text-align: center; vertical-align: middle; font-size: 20px; font-weight: bold; color: #efefef;\">:</td>" +
		"  <td><button id=\"t-min\" class=\"v-min\">" + fmt.Sprintf("%02d", min) + "</button></td>" +
		"</tr>"

	if updownbuttons {
		h += "<tr>" +
			"  <td><button id=\"ts-hour-down\" class=\"ts-ud clock-spinner dir-down type-hour\" data-sid=\"t-hour\"><i class=\"fa fa-tdown\"></i></button></td>" +
			"  <td></td>" +
			"  <td><button id=\"ts-min-down\" class=\"ts-ud clock-spinner dir-down type-min\" data-sid=\"t-min\"><i class=\"fa fa-tdown\"></i></button></td>" +
			"</tr>"
	}
	h += "</table>" +
		"  <input type=\"hidden\" id=\"hidedhour\" name=\"hour\" " +
		"    value=\"" + fmt.Sprintf("%d", hour) + "\"/>" +
		"  <input type=\"hidden\" id=\"hidedmin\" name=\"min\" data-pnlid=\"" + panelid + "\"" +
		"    value=\"" + fmt.Sprintf("%d", min) + "\"/>" +
		"</div>"
	return h
}

func htmlScheduleDays(sdl Schedule, mode string, breakline bool) string {
	var days_map *map[int]string = &days_short

	if mode == "oneletter" {
		days_map = &days_oneletter
	}

	html := "<div class=\"schedule-item-days\">"
	if sdl.dayMon {
		html += "<span class=\"s-on-day\">" + (*days_map)[0] + "</span>"
	} else {
		html += "<span class=\"s-off-day\">" + (*days_map)[0] + "</span>"
	}
	if sdl.dayTue {
		html += "<span class=\"s-on-day\">" + (*days_map)[1] + "</span>"
	} else {
		html += "<span class=\"s-off-day\">" + (*days_map)[1] + "</span>"
	}
	if sdl.dayWed {
		html += "<span class=\"s-on-day\">" + (*days_map)[2] + "</span>"
	} else {
		html += "<span class=\"s-off-day\">" + (*days_map)[2] + "</span>"
	}
	if sdl.dayThu {
		html += "<span class=\"s-on-day\">" + (*days_map)[3] + "</span>"
	} else {
		html += "<span class=\"s-off-day\">" + (*days_map)[3] + "</span>"
	}
	if breakline {
		html += "<br/>"
	}
	if sdl.dayFri {
		html += "<span class=\"s-on-day\">" + (*days_map)[4] + "</span>"
	} else {
		html += "<span class=\"s-off-day\">" + (*days_map)[4] + "</span>"
	}
	if sdl.daySat {
		html += "<span class=\"s-on-day\">" + (*days_map)[5] + "</span>"
	} else {
		html += "<span class=\"s-off-day\">" + (*days_map)[5] + "</span>"
	}
	if sdl.daySun {
		html += "<span class=\"s-on-day\">" + (*days_map)[6] + "</span>"
	} else {
		html += "<span class=\"s-off-day\">" + (*days_map)[6] + "</span>"
	}
	html += "</div>" //.schedule-item-days
	return html
}
