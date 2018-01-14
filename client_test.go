package graphqlgo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestError struct {
	message string
}

func (e TestError) Error() string {
	return fmt.Sprintf("%v", e.message)
}

type testErrorType struct {
	Var1 string
	Var2 string
}

func (t testErrorType) MarshalJSON() ([]byte, error) {
	return []byte{}, TestError{"Failed to Marshal data"}
}

func TestExecuteReturnsErrorIfVariableMarshallingToStringFails(t *testing.T) {
	variable := testErrorType{}
	client := Client{"", &http.Client{}}
	request := GraphQLRequest{"", variable, func(b *[]byte) (interface{}, error) { return nil, nil }}
	res, err := client.Execute(request)
	if err == nil {
		t.Fatalf("expected error but received: %v", res)
	}
}

func TestExecuteReturnsErrorIfStatusCodeIsNotOkay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := Client{server.URL, server.Client()}
	request := GraphQLRequest{"query", "example", func(b *[]byte) (interface{}, error) { return "dha", nil }}
	res, err := client.Execute(request)
	if err == nil {
		t.Fatalf("expected error but received: %v", res)
	}
}

func TestExecuteReturnsGraphQLResultIfSuccessful(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(([]byte("{ \"key\": \"Test Result\" }")))
	}))
	defer server.Close()

	client := Client{server.URL, server.Client()}
	request := GraphQLRequest{"query", "example", func(b *[]byte) (interface{}, error) {
		var v map[string]string
		e := json.Unmarshal(*b, &v)
		return v, e
	}}
	res, err := client.Execute(request)
	if err != nil {
		t.Fatalf("Expected Execute to succeed but received: %v", err)
	}

	resultData := res.Result.(map[string]string)
	if val, _ := resultData["key"]; val != "Test Result" {
		t.Fatalf("Incorred Result data received")
	}
}
