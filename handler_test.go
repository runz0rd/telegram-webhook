package telegramwebhook

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestTelegramHandler_Handler(t *testing.T) {
	kubewatch_tmpl := fmt.Sprintf("{{if .eventmeta}}{{if eq .eventmeta.reason %q %q}}{{.text}}{{end}}{{end}}", "created", "deleted")
	bvtd_tmpl := "{{$diff := .diff_mins}}{{range $coin, $change := .coins}}{{$coin}} => {{$change}}% in {{$diff}} mins\n{{end}}"
	type args struct {
		json     string
		template string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"kubewatch_pass", args{
			json: fmt.Sprintf("{%q:%q, %q:{%q:%q}}",
				"text", "testing `some` important `stuff`", "eventmeta", "reason", "created"),
			template: kubewatch_tmpl,
		}, false},
		{"kubewatch_no_message", args{
			json: fmt.Sprintf("{%q:%q, %q:{%q:%q}}",
				"text", "testing `some` important `stuff`", "eventmeta", "reason", "created"),
			template: kubewatch_tmpl,
		}, true},
		{"kubewatch_fail_template", args{
			json:     fmt.Sprintf("{%q:%q}", "asd", "test"),
			template: kubewatch_tmpl,
		}, true},
		{"kubewatch_fail_json", args{
			json:     fmt.Sprintf("{%q:%q}", "text", "test"),
			template: kubewatch_tmpl,
		}, true},
		{"bvtb_pass", args{
			json: fmt.Sprintf("{%q:%d, %q:{%q:%q, %q:%q}}",
				"diff_mins", 60, "coins", "BSBUSD", "69", "TTBUSD", "11"),
			template: bvtd_tmpl,
		}, false},
	}
	c, err := ReadConfig("config.yaml")
	if err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th, err := NewTelegramHandler(c.BotToken, tt.args.template, 0)
			if err != nil {
				t.Error(err)
			}
			reader := strings.NewReader(tt.args.json)
			req, _ := http.NewRequest("POST", path.Join("/webhook/", "779348941"), reader)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := http.HandlerFunc(th.Handler)
			handler.ServeHTTP(w, req)

			// Check the status code is what we expect.
			if status := w.Code; status != http.StatusOK && !tt.wantErr {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		})
	}
}

func Test_executeTemplate(t *testing.T) {
	type args struct {
		templ string
		data  map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"debug", args{"{{.text}}", map[string]interface{}{"text": fmt.Sprint(`Error creating: pods "helm-install-traefik-crd-" is forbidden: error looking up service account`)}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executeTemplate(tt.args.templ, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			spew.Dump(got)
		})
	}
}
