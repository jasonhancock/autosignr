package autosignr_test

import (
	"path"
	"runtime"
	"testing"

	"github.com/jasonhancock/autosignr"
)

func TestPSKExtraction(t *testing.T) {

	_, filename, _, _ := runtime.Caller(0)
	f := path.Join(path.Dir(filename), "testdata", "cert_csr_psk.pem")

	psk, err := autosignr.PuppetPSKFromCSRFile(f)

	if err != nil {
		t.Errorf("Not expecting an error extracting PSK from cert_csr_psk.pem")
	}

	if psk != "my_preshared_key_jason" {
		t.Errorf("PSK did not match expected value")
	}
}

func TestCertnameFromFilename(t *testing.T) {

	f := "/path/to/a/filename.example.com.pem"

	certname := autosignr.CertnameFromFilename(f)

	if certname != "filename.example.com" {
		t.Errorf("Returned certname did not match expected")
	}
}
