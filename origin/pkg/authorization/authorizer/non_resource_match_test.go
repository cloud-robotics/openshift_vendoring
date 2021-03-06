package authorizer

import (
	"testing"

	"github.com/openshift/kubernetes/pkg/util/sets"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

type nonResourceMatchTest struct {
	url            string
	matcher        string
	expectedResult bool
}

func TestNonResourceMatchStar(t *testing.T) {
	test := &nonResourceMatchTest{
		url:            "first/second",
		matcher:        "first/*",
		expectedResult: true,
	}
	test.run(t)
}

func TestNonResourceMatchExact(t *testing.T) {
	test := &nonResourceMatchTest{
		url:            "first/second",
		matcher:        "first/second",
		expectedResult: true,
	}
	test.run(t)
}

func TestNonResourceMatchMatcherEndsShort(t *testing.T) {
	test := &nonResourceMatchTest{
		url:            "first/second",
		matcher:        "first",
		expectedResult: false,
	}
	test.run(t)
}

func TestNonResourceMatchURLEndsShort(t *testing.T) {
	test := &nonResourceMatchTest{
		url:            "first",
		matcher:        "first/second",
		expectedResult: false,
	}
	test.run(t)
}

func TestNonResourceMatchNoSimilarity(t *testing.T) {
	test := &nonResourceMatchTest{
		url:            "first/second",
		matcher:        "foo",
		expectedResult: false,
	}
	test.run(t)
}

func (test *nonResourceMatchTest) run(t *testing.T) {
	attributes := &DefaultAuthorizationAttributes{
		NonResourceURL: true,
		URL:            test.url,
	}

	rule := authorizationapi.PolicyRule{NonResourceURLs: sets.NewString(test.matcher)}

	result := attributes.nonResourceMatches(rule)

	if result != test.expectedResult {
		t.Errorf("Expected %v, got %v", test.expectedResult, result)
	}

}
