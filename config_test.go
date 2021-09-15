package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	type testCase struct {
		input          string
		expectedConfig *config
		expectedError  error
	}

	for _, test := range []testCase{
		{
			input:          "#123",
			expectedConfig: &config{paths: map[string]string{}},
			expectedError:  nil,
		},
		{
			input:          "#123\r\n",
			expectedConfig: &config{paths: map[string]string{}},
			expectedError:  nil,
		},
		{
			input:          "   \t#    123 # \r\n",
			expectedConfig: &config{paths: map[string]string{}},
			expectedError:  nil,
		},
		{
			input:          "   \t#    123 # \r\n   #hh",
			expectedConfig: &config{paths: map[string]string{}},
			expectedError:  nil,
		},
		{
			input: "   \t 123 # \r\n   #hh",
			expectedConfig: &config{paths: map[string]string{
				"123 #": "123 #",
			}},
			expectedError: nil,
		},
		{
			input: "  alias\n  \t\ralias2\t   \r\n    \r\n",
			expectedConfig: &config{paths: map[string]string{
				"alias":  "alias",
				"alias2": "alias2",
			}},
			expectedError: nil,
		},
		{
			input: "  alias  : \rfile\n  \t\ralias2\t   \r\n ",
			expectedConfig: &config{paths: map[string]string{
				"alias":  "file",
				"alias2": "alias2",
			}},
			expectedError: nil,
		},
		{
			input:          "  alias  : \rfile:file\n  \t\ralias2\t   \r\n ",
			expectedConfig: nil,
			expectedError:  errors.New("malformed line"),
		},
		{
			input:          "alias: 123\nalias:321",
			expectedConfig: nil,
			expectedError:  errors.New("duplicated alias"),
		},
	} {
		reader := strings.NewReader(test.input)
		config, err := parseConfig(reader)
		assert.Equal(t, config, test.expectedConfig)
		if test.expectedError != nil {
			assert.Error(t, err, test.expectedError)
		} else {
			assert.Nil(t, err)
		}
	}
}
