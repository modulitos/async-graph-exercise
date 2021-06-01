package main

import (
	"bytes"
	"encoding/json"
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

func setupMock(state map[byte]NodeJSON) {
	// Override the client that gets set in our main function:
	Client = &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			nodeId := []byte(req.URL.String())[0]
			jsonResponse, err := json.Marshal(state[nodeId])
			if err != nil {
				return nil, err
			}
			r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}
}

func TestCallSuccess(t *testing.T) {
	type args struct {
		nodeId byte
	}

	type want struct {
		reward int
	}

	cases := map[string]struct {
		args
		want
	}{
		"one node 100": {
			args: args{nodeId: 'a'},
			want: want{reward: 100},
		},
		"one node 200": {
			args: args{nodeId: 'b'},
			want: want{reward: 200},
		},
	}

	setupMock(map[byte]NodeJSON{
		'a': {
			Children: []string{},
			Reward:   100,
		},
		'b': {
			Children: []string{},
			Reward:   200,
		},
	})

	// iterate over all our test cases:
	g := NewGomegaWithT(t)
	for name, testCase := range cases {

		t.Run(name, func(t *testing.T) {
			// Given the testCase:

			// When:
			score, err := CalculateReward(string(testCase.args.nodeId))

			// Then
			g.Expect(err).To(BeNil(), "Got an error back!")
			g.Expect(score).To(Equal(testCase.want.reward), "Reward doesn't match up!")
		})
	}
}
