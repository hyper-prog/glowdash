/*
	GlowDash - Smart Home Web Dashboard

	(C) 2024-2026 Péter Deák (hyper80@gmail.com)
	License: GPLv2
*/

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hyper-prog/smartyaml"
)

type GlowdashConsoleDataType struct {
	lines []string
	pos   int
}

var GlowdashConsole GlowdashConsoleDataType

type PageConsole struct {
	PageBase
}

func NewPageConsole() *PageConsole {
	return &PageConsole{
		PageBase{
			idStr:      "",
			pageType:   Console,
			title:      "",
			deviceType: "",
			index:      0,
		},
	}
}

func (c *GlowdashConsoleDataType) Init() {
	if MaxLogLines <= 0 {
		return
	}
	c.lines = make([]string, MaxLogLines)
	for i := 0; i < MaxLogLines; i++ {
		c.lines[i] = ""
	}
	c.pos = 0
}

func (c *GlowdashConsoleDataType) Write(s string) {
	if MaxLogLines <= 0 {
		return
	}
	ct := time.Now()
	ll := fmt.Sprintf("&lt;%d-%02d-%02d %02d:%02d:%02d&gt; %s", ct.Year(), ct.Month(), ct.Day(), ct.Hour(), ct.Minute(), ct.Second(), s)

	c.lines[c.pos] = ll
	c.pos++
	if c.pos >= MaxLogLines {
		c.pos = 0
	}
}

func (p *PageConsole) LoadCustomConfig(sy smartyaml.SmartYAML, indexInConfig int) {

}

func (p PageConsole) PageHtml(withContainer bool, r *http.Request) string {
	html := "<div class=\"logpage-inner-show\">"

	if MaxLogLines > 0 {
		for i := MaxLogLines - 1; i >= 0; i-- {
			ll := GlowdashConsole.lines[(GlowdashConsole.pos+i)%MaxLogLines]
			if len(ll) > 0 {
				html += ll + "<br/>"
			}
		}
	}
	html += "</div>"

	if withContainer {
		return fmt.Sprintf("<div id=\"pc-%s\" class=\"fullpage-content\" tabindex=\"-1\">", p.IdStr()) +
			html + "</div>"
	}

	return html
}
