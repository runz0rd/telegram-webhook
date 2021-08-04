package telegramwebhook

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
		{"pass", args{fmt.Sprintf("{%q:%q}", "text", "test_pass")}, false},
		{"fail_template", args{fmt.Sprintf("{%q:%q}", "asd", "test")}, true},
		{"fail_json", args{fmt.Sprintf("{{%q:%q}", "text", "test")}, true},
	}
	c, err := ReadConfig("config.yaml")
	if err != nil {
		t.Error(err)
	}
	th, err := NewTelegramHandler(c.BotToken, c.ChatId, c.MessageTemplate)
	if err != nil {
		t.Error(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.args.json)
			req, _ := http.NewRequest("POST", c.Serve.Path, reader)
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
