package fs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileReadLine(t *testing.T) {
	testcases := []struct {
		name                  string
		filecontent           string
		readMustBeSuccessfull bool
	}{
		{
			name:                  "no-ending-newline",
			filecontent:           "test",
			readMustBeSuccessfull: true,
		},

		{
			name:                  "ending-newline",
			filecontent:           "test\n",
			readMustBeSuccessfull: true,
		},

		{
			name:                  "empty file",
			filecontent:           "",
			readMustBeSuccessfull: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			file, err := ioutil.TempFile("", "")
			require.NoError(t, err)

			defer os.Remove(file.Name())

			_, err = file.WriteString(tc.filecontent)
			require.NoError(t, err)

			read, err := FileReadLine(file.Name())
			if tc.readMustBeSuccessfull {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			assert.Equal(t, tc.filecontent, read)
		})
	}

}
