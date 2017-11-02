package autosignr

import (
	"testing"

	"github.com/cheekybits/is"
)

func TestPSKExtraction(t *testing.T) {
	is := is.New(t)

	psk, err := PuppetPSKFromCSRFile("testdata/cert_csr_psk.pem")
	is.NoErr(err)
	is.Equal(psk, "my_preshared_key_jason")
}

func TestCertnameFromFilename(t *testing.T) {
	is := is.New(t)

	f := "/path/to/a/filename.example.com.pem"
	is.Equal("filename.example.com", CertnameFromFilename(f))
}
