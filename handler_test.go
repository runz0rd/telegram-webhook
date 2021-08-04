package telegramwebhook

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
)

func TestTelegramHandler_Handler(t *testing.T) {
	type args struct {
		json string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"pass", args{fmt.Sprintf("{%q:%q, %q:{%q:%q}}",
			"text", "A `pod` in namespace `default` has been `created`:\n`default/feedtransmission-8b7f44dff-bk8tg`",
			"eventmeta", "reason", "created"),
		}, false},
		{"pass_no_message", args{fmt.Sprintf("{%q:%q, %q:{%q:%q}}",
			"text", "A `pod` in namespace `default` has been `updated`:\n`default/feedtransmission-8b7f44dff-bk8tg`",
			"eventmeta", "reason", "updated"),
		}, false},
		{"fail_template", args{fmt.Sprintf("{%q:%q}", "asd", "test")}, true},
		{"fail_json", args{fmt.Sprintf("{{%q:%q}", "text", "test")}, true},
	}
	c, err := ReadConfig("config.yaml")
	if err != nil {
		t.Error(err)
	}
	th, err := NewTelegramHandler(c.BotToken, c.Webhooks[0].MessageTemplate)
	if err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.args.json)
			req, _ := http.NewRequest("POST", path.Join(c.Webhooks[0].Path, "779348941"), reader)
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
