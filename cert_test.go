package autosignr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPSKExtraction(t *testing.T) {
	psk, err := PuppetPSKFromCSRFile("testdata/cert_csr_psk.pem")
	require.NoError(t, err)
	require.Equal(t, "my_preshared_key_jason", psk)
}

func TestCertnameFromFilename(t *testing.T) {
	f := "/path/to/a/filename.example.com.pem"
	require.Equal(t, "filename.example.com", CertnameFromFilename(f))
}
