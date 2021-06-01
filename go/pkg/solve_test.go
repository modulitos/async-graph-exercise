package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
)

// Custom type that allows setting the func that our Mock Do func will run
// instead
type MockDoType func(req *http.Request) (*http.Response, error)

type MockClient struct {
	MockDo MockDoType
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

func TestCallSuccess(t *testing.T) {
	fmt.Println("testing")
	// Given:
	jsonResponse := `{
		"children": [],
		"reward": 100
	}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
	Client = &MockClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			fmt.Println("inside mock!")
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}

	// When:
	score, err := CalculateReward("asdf")

	// Then
	g := NewGomegaWithT(t)
	g.Expect(err).To(BeNil(), "Got an error back!")
	g.Expect(score).To(Equal(100), "Reward doesn't match up!")
}
