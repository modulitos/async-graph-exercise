package solver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
)

// Custom type that allows setting the func that our Mock Do func will run
// instead. Borrowed from here:
// https://levelup.gitconnected.com/mocking-outbound-http-calls-in-golang-9e5a044c2555
type MockDoType func(req *http.Request) (*http.Response, error)

type MockClient struct {
	MockDo MockDoType
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

func setupMock(state map[string]NodeJSON) {
	// Override the client that gets set in our main function:
	Client = &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			value, mapContainsKey := state[req.URL.String()]
			if !mapContainsKey {
				return nil, errors.New(fmt.Sprintf("mock is not configured with nodeId: %s", req.URL))
			}
			jsonResponse, err := json.Marshal(value)
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
		"one node, 1 child": {
			args: args{nodeId: 'c'},
			want: want{reward: 400},
		},
		"one node, 2 children": {
			args: args{nodeId: 'd'},
			want: want{reward: 700},
		},
		"one node, with grandchildren": {
			args: args{nodeId: 'e'},
			want: want{reward: 900},
		},
	}

	setupMock(map[string]NodeJSON{
		GetUriForNode('a'): {
			Children: []string{},
			Reward:   100,
		},
		GetUriForNode('b'): {
			Children: []string{},
			Reward:   200,
		},
		GetUriForNode('c'): {
			Children: []string{GetUriForNode('a')},
			Reward:   300,
		},
		GetUriForNode('d'): {
			Children: []string{GetUriForNode('a'), GetUriForNode('b')},
			Reward:   400,
		},
		GetUriForNode('e'): {
			Children: []string{GetUriForNode('c')},
			Reward:   500,
		},
	})

	// iterate over all our test cases:
	g := NewGomegaWithT(t)
	for name, testCase := range cases {

		t.Run(name, func(t *testing.T) {
			// Given the testCase:

			// When:
			score, err := CalculateReward(testCase.args.nodeId)

			// Then
			g.Expect(err).To(BeNil(), "Unexpected error in test.")
			g.Expect(score).To(Equal(testCase.want.reward), "Unexpected reward value.")
		})
	}
}

func TestGetUriForNode(t *testing.T) {
	g := NewGomegaWithT(t)

	// Given:
	var nodeId byte = 'a'

	// When:
	res := GetUriForNode(nodeId)

	// Then:
	g.Expect(res).To(Equal("https://graph.modulitos.com/node/a"))
}
