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

// "table-driven tests" inspired by:
// https://nathanleclaire.com/blog/2015/10/10/interfaces-and-composition-for-effective-unit-testing-in-golang/
// and https://dave.cheney.net/2013/06/09/writing-table-driven-tests-in-go
// and https://dave.cheney.net/2019/05/07/prefer-table-driven-tests
func TestCalculateReward(t *testing.T) {
	formatter := NewNodeFormatter()
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
		"one_node_100": {
			args: args{nodeId: 'a'},
			want: want{reward: 100},
		},
		"one_node_200": {
			args: args{nodeId: 'b'},
			want: want{reward: 200},
		},
		"one_node_1_child": {
			args: args{nodeId: 'c'},
			want: want{reward: 400},
		},
		"one_node_2_children": {
			args: args{nodeId: 'd'},
			want: want{reward: 700},
		},
		"with_grandchildren": {
			args: args{nodeId: 'e'},
			want: want{reward: 900},
		},
		"caching_an_extra_call_to_a": {
			args: args{nodeId: 'f'},
			want: want{reward: 1400},
		},
		"caching_many_calls": {
			args: args{nodeId: 'g'},
			want: want{reward: 3300},
		},
		// // TODO: test error handling
		// "one node, which doesn't exist": {
		// 	args: args{nodeId: 'z'},
		// 	want: want{err: "mock is not configured with nodeId: z"},
		// },
	}

	setupMock(map[string]NodeJSON{
		formatter.GetUriForNode('a'): {
			Children: []string{},
			Reward:   100,
		},
		formatter.GetUriForNode('b'): {
			Children: []string{},
			Reward:   200,
		},
		formatter.GetUriForNode('c'): {
			Children: []string{"a"},
			Reward:   300,
		},
		formatter.GetUriForNode('d'): {
			Children: []string{"a", "b"},
			Reward:   400,
		},
		formatter.GetUriForNode('e'): {
			Children: []string{"c"},
			Reward:   500,
		},
		formatter.GetUriForNode('f'): {
			Children: []string{"d", "a"},
			Reward:   600,
		},
		formatter.GetUriForNode('g'): {
			Children: []string{"f", "a", "b", "b", "d"},
			Reward:   700,
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
	formatter := NewNodeFormatter()
	res := formatter.GetUriForNode(nodeId)

	// Then:
	g.Expect(res).To(Equal("https://graph.modulitos.com/node/a"))
}
