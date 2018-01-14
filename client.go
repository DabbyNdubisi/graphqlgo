package graphqlgo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// httpResponse encapsulates the resulting response and encountered error
// of an http request.
type httpResponse struct {
	Resp *http.Response
	Err  error
}

type requesterFunc func(url string, values map[string][]string) httpResponse

// Client encapsulates the structure in charge of making graphql requests.
// It is initialized with the URL to the graphql server, and an HttpClient.
type Client struct {
	Url        string
	HttpClient *http.Client
}

// GraphQLRequest represents the body of the graphql request.
// The Query could either be a mutation query or a regular query.
// It uses the ResultParser to convert the data response from the server
// into the expected type.
type GraphQLRequest struct {
	Query        string
	Variables    interface{}
	ResultParser func(*[]byte) (interface{}, error)
}

// GraphQLResult represents the result of the executed GraphQLQuery
type GraphQLResult struct {
	Query  GraphQLRequest
	Result interface{}
}

// Execute executes the given GraphQLQuery.
// It returns the GraphQLResult if successful or any error encountered.
func (c Client) Execute(q GraphQLRequest) (*GraphQLResult, error) {
	encodedVars, encodingError := encodeVariables(q.Variables)
	if encodingError != nil {
		return nil, encodingError
	}

	resp := c.performRequest(c.Url, url.Values{"query": {q.Query}, "variables": {encodedVars}})
	if resp.Err != nil {
		log.Printf("Network error occured: %v", resp.Err)
		return nil, resp.Err
	}
	defer resp.Resp.Body.Close()

	result, parseError := q.parseHttpResponseBody(resp.Resp.Body)
	if parseError != nil {
		log.Printf("Parse failed with error: %v", parseError)
		return nil, parseError
	}
	return &GraphQLResult{q, result}, nil
}

func (c Client) performRequest(url string, values map[string][]string) httpResponse {
	requester := requesterFunc(func(url string, values map[string][]string) httpResponse {
		resp, err := c.HttpClient.PostForm(url, values)
		if err != nil {
			return httpResponse{nil, err}
		}
		if resp.StatusCode != http.StatusOK {
			return httpResponse{nil, &http.ProtocolError{fmt.Sprintf("Request received status code: %v", resp.StatusCode)}}
		}

		return httpResponse{resp, nil}
	})

	requesters := []requesterFunc{requester, requester, requester}
	started := time.Now()
	result := first(url, values, requesters...)
	elapsed := time.Since(started)
	log.Printf("Request took %v", elapsed)
	return result
}

func first(url string, values map[string][]string, requester ...requesterFunc) httpResponse {
	c := make(chan httpResponse)

	firstHelper := func(i int) { c <- requester[i](url, values) }
	for index := range requester {
		go firstHelper(index)
	}
	return <-c
}

func (q GraphQLRequest) parseHttpResponseBody(body io.ReadCloser) (interface{}, error) {
	val, _ := ioutil.ReadAll(body)
	result, err := q.ResultParser(&val)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func encodeVariables(variables interface{}) (string, error) {
	bytes, err := json.Marshal(variables)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
