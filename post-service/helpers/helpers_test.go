package helpers

import (
	"io"
	"net/http/httptest"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestResponseWithPayload(t *testing.T) {
	type testCase = struct {
		StatusCode int
		Body       []byte
	}
	tests := []testCase{
		{StatusCode: 100, Body: []byte("test1")},
		{StatusCode: 200, Body: []byte("")},
		{StatusCode: 300, Body: []byte{}},
	}
	for _, test := range tests {
		writer := httptest.NewRecorder()
		ResponseWithPayload(writer, test.StatusCode, test.Body)
		respBytes, _ := io.ReadAll(writer.Body)
		if string(respBytes) != string(test.Body) {
			t.Error("body does not match")
		}
		if writer.Code != test.StatusCode {
			t.Error("status code does not match")
		}
	}
}
