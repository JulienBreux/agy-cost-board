package server

import (
	_ "embed"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

//go:embed statusline.sh.tmpl
var statuslineTemplateStr string

var statuslineTmpl = template.Must(template.New("statusline.sh").Parse(statuslineTemplateStr))

type statuslineData struct {
	DefaultBoardURL string
	DefaultUser     string
	DefaultDays     string
	DefaultTTL      string
}

func (s *Server) handleStatuslineScript(w http.ResponseWriter, r *http.Request) {
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	if host == "" {
		host = "localhost:8080"
	}

	boardURL := proto + "://" + host

	defaultUser := r.URL.Query().Get("user")
	// Sanitize user string to prevent escape from bash double quotes
	defaultUser = strings.ReplaceAll(defaultUser, `"`, "")
	defaultUser = strings.ReplaceAll(defaultUser, `$`, "")
	defaultUser = strings.ReplaceAll(defaultUser, "`", "")
	defaultUser = strings.ReplaceAll(defaultUser, `\`, "")

	defaultDays := r.URL.Query().Get("days")
	if defaultDays != "" {
		if val, err := strconv.Atoi(defaultDays); err != nil || val <= 0 {
			defaultDays = "30"
		}
	} else {
		defaultDays = "30"
	}

	defaultTTL := r.URL.Query().Get("ttl")
	if defaultTTL != "" {
		if val, err := strconv.Atoi(defaultTTL); err != nil || val <= 0 {
			defaultTTL = "300"
		}
	} else {
		defaultTTL = "300"
	}

	data := statuslineData{
		DefaultBoardURL: boardURL,
		DefaultUser:     defaultUser,
		DefaultDays:     defaultDays,
		DefaultTTL:      defaultTTL,
	}

	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Header().Set("Content-Disposition", `inline; filename="statusline.sh"`)
	w.WriteHeader(http.StatusOK)
	_ = statuslineTmpl.Execute(w, data)
}
