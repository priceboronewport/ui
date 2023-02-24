package ui

import (
	".."
	"../../element"
	"../../logger"
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

type Page struct {
	Head    string
	header  string
	Content string
	Menu    string
	W       http.ResponseWriter
	R       *http.Request
	P       webapp.HandlerParams
}

type uiParams struct {
	Head    template.HTML
	Header  template.HTML
	Content template.HTML
	Menu    template.HTML
}

func New(w http.ResponseWriter, r *http.Request, p webapp.HandlerParams, title, icon string) *Page {
	head := "<title>" + html.EscapeString(title) + "</title>"
	if icon != "" {
		head += webapp.Icon(icon)
	}
	head += webapp.Stylesheet("/res/css/ui.css")
	head += webapp.Script("/res/js/ui.js")
	pg := Page{Head: head, W: w, R: r, P: p}
	return &pg
}

func (pg *Page) AddHeaderGlyph(src string) {
	pg.header += "<td><div class='icon' style='background-image: url(\"" + src + "\")'></div></td>"
}

func (pg *Page) AddHeaderIcon(src string, url string) {
	if url == "" {
		pg.header += "<td class='active'><div class='icon' style='background-image: url(\"" + src + "\")'></div></td>"
	} else {
		pg.header += "<td><div class='icon' onClick='window.location.href=\"" + url + "\"' style='cursor: pointer; background-image: url(\"" + src + "\")'></div></td>"
	}
}

func (pg *Page) AddHeaderLabel(label string, url string) {
	if url == "" {
		pg.header += "<td class='active'>" + html.EscapeString(label) + "</td>"
	} else {
		pg.header += "<td><a href='" + url + "'>" + html.EscapeString(label) + "</a></td>"
	}
}

func (pg *Page) AddScript(url string) (err error) {
	elements, err := element.Parse(pg.Head)
	if err == nil {
		for _, e := range elements {
			if e.Tag == "script" && e.Attributes["src"] == url {
				return
			}
		}
		pg.Head += webapp.Script(url)
	}
	return
}

func (pg *Page) AddStylesheet(url string) (err error) {
	elements, err := element.Parse(pg.Head)
	if err == nil {
		for _, e := range elements {
			if e.Tag == "link" && e.Attributes["rel"] == "stylesheet" && e.Attributes["href"] == url {
				return
			}
		}
		pg.Head += webapp.Stylesheet(url)
	}
	return
}

func (pg *Page) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return pg.R.FormFile(key)
}

func (pg *Page) HasPermission(permission string) bool {
	return webapp.HasPermission(pg.P.Username, permission)
}

func (pg *Page) Method() string {
	return pg.R.Method
}

func (pg *Page) Output(content string) {
	fmt.Fprintf(pg.W, "%s", content)
}

func (pg *Page) Param(name string) string {
	return pg.R.URL.Query().Get(name)
}

func (pg *Page) ParamExists(name string) (result bool) {
    keys, result := pg.R.URL.Query()[name]
    if len(keys[0]) < 1 {
      result = false
    }
    return
}

func (pg *Page) ParamInt(name string) (value int) {
	value, _ = strconv.Atoi(pg.R.URL.Query().Get(name))
	return
}

func (pg *Page) ParamNum(name string) (value float64) {
	value, _ = strconv.ParseFloat(strings.ReplaceAll(pg.R.URL.Query().Get(name), ",", ""), 64)
	return
}

func (pg *Page) ParseMultipartForm(size int64) {
	pg.R.ParseMultipartForm(32 << 20)
}

func (pg *Page) ParseForm() {
	pg.R.ParseForm()
}

func (pg *Page) PostParam(name string) string {
	return pg.R.Form.Get(name)
}

func (pg *Page) Render() {
	header := "<table id='ui_header'><tr>"
	if pg.P.Username != "" {
		header += "<td id='ui_menu'><button class='ui_menu_button' onClick='UIMenuShow(\"ui_menu_content\")'>&nbsp;</button></td>"
		pg.Menu += "<hr/><a href='/logout'>Log Out</a>"
	}
	header += pg.header
	if pg.P.Username != "" {
		user := webapp.User(pg.P.Username)
		display_name := user["first_name"] + " " + user["last_name"]
		if display_name == " " {
			display_name = pg.P.Username
		}
		header += "<td id='ui_user'><button class='ui_menu_button' onClick='UIMenuShow(\"ui_usermenu_content\")'>" + display_name + "</button></td>"
	}
	header += "</tr></table>"
	webapp.Render(pg.W, "ui.html", uiParams{Head: template.HTML(pg.Head),
		Header: template.HTML(header), Content: template.HTML(pg.Content),
		Menu: template.HTML(pg.Menu)})
}

func (pg *Page) RenderError(message string, url string) {
	logger.Error("ui.RenderError()", message)
	pg.RenderInfo(message, url)
}

func (pg *Page) RenderInfo(message string, url string) {
	pg.Content = "<div class='ui_modal'><div class='ui_modal_content'>"
	if url != "" {
		pg.Content += "<span class='ui_modal_close'><a href='" + url + "'>&nbsp;</a></span>"
	}
	pg.Content += "<p>" + message + "</p></div></div>"
	pg.Render()
}

func (pg *Page) RenderFile(filename string) {
	mime_type := webapp.ContentType(filename)
	fb, err := ioutil.ReadFile(filename)
	if (err != nil) || (mime_type == "") {
		pg.W.WriteHeader(http.StatusNotFound)
		fmt.Fprint(pg.W, "Not Found")
		return
	}
	b := bytes.NewBuffer(fb)
	pg.W.Header().Set("Content-type", mime_type)
	b.WriteTo(pg.W)
}

func (pg *Page) RenderQuestion(message string, url_yes string, url_no string) {
	pg.RenderInfo(message+"<hr/><div style='text-align: right'><a href='"+url_yes+"'>Yes</a>&nbsp;<a href='"+url_no+"'>No</a></div>", "")
}

func (pg *Page) Redirect(url string) {
	webapp.Redirect(pg.W, pg.R, url)
}

func (pg *Page) SessionValuesRead(keys ...string) string {
	if len(keys) > 0 {
		if len(keys) > 1 {
			return webapp.SessionValuesRead(pg.P, keys[0], keys[1])
		}
		return webapp.SessionValuesRead(pg.P, keys[0])
	}
	return ""
}

func (pg *Page) SessionValuesWrite(key, value string) error {
	return webapp.SessionValuesWrite(pg.P, key, value)
}

func (pg *Page) SubPath(index int) string {
	return webapp.UrlPath(pg.R, index)
}

func (pg *Page) Username() string {
	return pg.P.Username
}
