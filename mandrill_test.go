package mandrill

import (
  "fmt"
  "testing"
  "reflect"
  "net/http/httptest"
  "net/http"
  "net/url"
)

func expect(t *testing.T, a interface{}, b interface{}) {
  if a != b {
    t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
  }
}

func refute(t *testing.T, a interface{}, b interface{}) {
  if a == b {
    t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
  }
}

func MessagesTestTools(code int, body string) (*httptest.Server, *Client)  {

  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(code)
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintln(w, body)
  }))

  tr := &http.Transport{
    Proxy: func(req *http.Request) (*url.URL, error) {
      return url.Parse(server.URL)
    },
  }
  httpClient := &http.Client{Transport: tr}

  client := ClientWithKey("APIKEY")
  client.HTTPClient = httpClient
  client.BaseURL = "http://example.com/api/1.0/"
  return server, client
}

// MessagesSendTemplate //////////

func Test_MessagesSendTemplate_Success(t *testing.T) {
  server, m := MessagesTestTools(200, `[{"email":"bob@example.com","status":"sent","reject_reason":"hard-bounce","_id":"1"}]`)
  defer server.Close()
  mandrillResponses, mandrillError, _ := m.MessagesSendTemplate(&Message{}, "cheese", map[string]string{"name": "bob"})

  expect(t, len(mandrillResponses), 1)
  expect(t, mandrillError, (*MError)(nil))

  correctResponse := &MResponse{
    Email: "bob@example.com",
    Status: "sent",
    RejectionReason: "hard-bounce",
    Id: "1",
  }
  expect(t, reflect.DeepEqual(correctResponse, mandrillResponses[0]), true)
}

func Test_MessagesSendTemplate_Fail(t *testing.T) {
  server, m := MessagesTestTools(400, `{"status":"error","code":12,"name":"Unknown_Subaccount","message":"No subaccount exists with the id 'customer-123'"}`)
  defer server.Close()
  mandrillResponses, mandrillError, _ := m.MessagesSendTemplate(&Message{}, "cheese", map[string]string{"name": "bob"})

  expect(t, len(mandrillResponses), 0)

  correctResponse := &MError{
    Status: "error",
    Code: 12,
    Name: "Unknown_Subaccount",
    Message: "No subaccount exists with the id 'customer-123'",
  }
  expect(t, reflect.DeepEqual(correctResponse, mandrillError), true)
}

// MessagesSend //////////

func Test_MessageSend_Success(t *testing.T) {
  server, m := MessagesTestTools(200, `[{"email":"bob@example.com","status":"sent","reject_reason":"hard-bounce","_id":"1"}]`)
  defer server.Close()
  mandrillResponses, mandrillError, _ := m.MessagesSend(&Message{})

  expect(t, len(mandrillResponses), 1)
  expect(t, mandrillError, (*MError)(nil))

  correctResponse := &MResponse{
    Email: "bob@example.com",
    Status: "sent",
    RejectionReason: "hard-bounce",
    Id: "1",
  }
  expect(t, reflect.DeepEqual(correctResponse, mandrillResponses[0]), true)
}

func Test_MessageSend_Fail(t *testing.T) {
  server, m := MessagesTestTools(400, `{"status":"error","code":12,"name":"Unknown_Subaccount","message":"No subaccount exists with the id 'customer-123'"}`)
  defer server.Close()
  mandrillResponses, mandrillError, _ := m.MessagesSend(&Message{})

  expect(t, len(mandrillResponses), 0)

  correctResponse := &MError{
    Status: "error",
    Code: 12,
    Name: "Unknown_Subaccount",
    Message: "No subaccount exists with the id 'customer-123'",
  }
  expect(t, reflect.DeepEqual(correctResponse, mandrillError), true)
}

// AddRecipient //////////

func Test_AddRecipient(t *testing.T) {
  m := &Message{}
  m.AddRecipient("bob@example.com", "Bob Johnson", "to")
  tos := []*To{&To{"bob@example.com", "Bob Johnson", "to"}}
  expect(t, reflect.DeepEqual(m.To, tos), true)
}

// ConvertMapToVariables /////

func Test_ConvertMapToVariables(t *testing.T) {
  m := map[string]string{"name": "bob", "food": "cheese"}
  target := ConvertMapToVariables(m)
  hand := []*Variable{
    &Variable{"name", "bob"},
    &Variable{"food", "cheese"},
  }
  expect(t, reflect.DeepEqual(target, hand), true)
}

// ConvertMapToVariablesForRecipient ////

func Test_ConvertMapToVariablesForRecipient(t *testing.T) {
  m := map[string]string{"name": "bob", "food": "cheese"}
  target := ConvertMapToVariablesForRecipient("bob@example.com", m)
  hand := &RcptMergeVars{"bob@example.com", ConvertMapToVariables(m)}
  expect(t, reflect.DeepEqual(target, hand), true)
}
