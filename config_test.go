package main

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type nonFunctioningReader struct{}

func (e nonFunctioningReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("error from reader")
}

func newStringReader(value string) io.Reader {
	return strings.NewReader(value)
}

func newNonFunctioningReader() io.Reader {
	return nonFunctioningReader{}
}

func TestParseConfig(t *testing.T) {
	type testCase struct {
		reader         io.Reader
		expectedConfig *config
		expectedError  error
	}

	for _, test := range []testCase{
		{
			reader:         newStringReader("#123"),
			expectedConfig: &config{paths: map[string]string{}},
			expectedError:  nil,
		},
		{
			reader:         newStringReader("#123\r\n"),
			expectedConfig: &config{paths: map[string]string{}},
			expectedError:  nil,
		},
		{
			reader:         newStringReader("   \t#    123 # \r\n"),
			expectedConfig: &config{paths: map[string]string{}},
			expectedError:  nil,
		},
		{
			reader:         newStringReader("   \t#    123 # \r\n   #hh"),
			expectedConfig: &config{paths: map[string]string{}},
			expectedError:  nil,
		},
		{
			reader: newStringReader("   \t 123 # \r\n   #hh"),
			expectedConfig: &config{paths: map[string]string{
				"123 #": "123 #",
			}},
			expectedError: nil,
		},
		{
			reader: newStringReader("  alias\n  \t\ralias2\t   \r\n    \r\n"),
			expectedConfig: &config{paths: map[string]string{
				"alias":  "alias",
				"alias2": "alias2",
			}},
			expectedError: nil,
		},
		{
			reader: newStringReader("  alias  : \rfile\n  \t\ralias2\t   \r\n "),
			expectedConfig: &config{paths: map[string]string{
				"alias":  "file",
				"alias2": "alias2",
			}},
			expectedError: nil,
		},
		{
			reader:         newStringReader("  alias  : \rfile:file\n  \t\ralias2\t   \r\n "),
			expectedConfig: nil,
			expectedError:  errors.New("malformed line"),
		},
		{
			reader:         newStringReader("alias: 123\nalias:321"),
			expectedConfig: nil,
			expectedError:  errors.New("duplicated alias"),
		},
		{
			reader:         newStringReader(""),
			expectedConfig: &config{paths: map[string]string{}},
			expectedError:  nil,
		},
		{
			reader:         newNonFunctioningReader(),
			expectedConfig: nil,
			expectedError:  errors.New("error from reader"),
		},
	} {
		config, err := parseConfig(test.reader)
		assert.Equal(t, config, test.expectedConfig)
		if test.expectedError != nil {
			assert.Error(t, err, test.expectedError)
		} else {
			assert.Nil(t, err)
		}
	}
}
