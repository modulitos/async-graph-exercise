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
	// TODO: make this into multiple tests using a map with "args" and "wanted"
	// structs.
	fmt.Println("testing")
	// Given:

	type args struct {
		// children []char
		// reward
		nodeId byte
	}

	type want struct {
		reward int
	}

	cases := map[string]struct {
		args
		want
	}{
		"happy": {
			args: args{nodeId: 'a'},
			want: want{reward: 100},
		},
	}

	jsonResponse := `{
		"children": [],
		"reward": 100
	}`
	r := ioutil.NopCloser(bytes.NewReader([]byte(jsonResponse)))

	Client = &MockClient{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}

	// iterate over all our test cases:
	g := NewGomegaWithT(t)
	for name, testCase := range cases {

		t.Run(name, func(t *testing.T) {

			// When:
			score, err := CalculateReward(GetUrlForNode(testCase.args.nodeId))

			// Then
			g.Expect(err).To(BeNil(), "Got an error back!")
			g.Expect(score).To(Equal(testCase.want.reward), "Reward doesn't match up!")
		})
	}
}
