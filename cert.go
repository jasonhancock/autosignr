package autosignr

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// The OID for Puppet's pp_preshared_key in the Certificate Extensions
// https://docs.puppetlabs.com/puppet/latest/reference/ssl_attributes_extensions.html
const puppetPSKoid string = "1.3.6.1.4.1.34380.1.1.4"

func CertnameFromFilename(file string) string {
	base := filepath.Base(file)
	extension := filepath.Ext(base)

	return base[0 : len(base)-len(extension)]
}

func CheckCert(conf Config, file string) (bool, error) {
	name := CertnameFromFilename(file)
	log.Debugf("CheckCert %s", name)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return false, err
	}

	result, err := ValidateCert(conf, data, name)

	if err != nil {
		return false, err
	}

	if !result {
		log.Warningf("Unable to validate instance %s", name)
	}

	return result, nil
}

func ValidateCert(conf Config, data []byte, certname string) (bool, error) {
	log.Debugf("ValidateCert %s", certname)
	if conf.CheckPSK {
		psk, err := PuppetPSKFromCSR(data)
		if err != nil {
			log.WithFields(log.Fields{
				"certname": certname,
				"err":      err,
			}).Warning("psk-extract-error")
			return false, err
		} else {
			if _, ok := conf.PresharedKeys[psk]; !ok {
				log.WithFields(log.Fields{
					"certname": certname,
					"psk":      psk,
				}).Warning("invalid-psk")
				return false, errors.New("Invalid PSK")
			}
		}
	}

	result := false
	for _, acct := range conf.Accounts {
		result = acct.Check(certname)
		if result {
			break
		}
	}

	return result, nil
}

func SignCert(conf Config, certname string) {
	cmd := fmt.Sprintf(conf.Cmdsign, certname)
	pieces := strings.Split(cmd, " ")

	cmdOut, err := exec.Command(pieces[0], pieces[1:]...).CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{
			"certname": certname,
			"err":      err,
			"output":   string(cmdOut),
		}).Error("signing-failure")
		return
	}
	log.WithFields(log.Fields{
		"certname": certname,
	}).Info("signing-success")
}

func ExistingCerts(conf Config) {

	matches, err := filepath.Glob(fmt.Sprintf("%s/*.pem", conf.Dir))
	if err != nil {
		log.Println("Glob error for: %s", err)
	}

	for _, cert := range matches {
		log.WithFields(log.Fields{
			"file": cert,
		}).Info("existing-csr")
		result, _ := CheckCert(conf, cert)

		if result {
			SignCert(conf, CertnameFromFilename(cert))
		}
	}
}

func PuppetPSKFromCSRFile(file string) (string, error) {
	var f string

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return f, err
	}

	f, err = PuppetPSKFromCSR(data)

	return f, err
}

func PuppetPSKFromCSR(data []byte) (string, error) {
	var f string

	block, _ := pem.Decode(data)
	if block == nil {
		return f, errors.New("No PEM data found in block")
	}

	parsedcsr, err := x509.ParseCertificateRequest(block.Bytes)

	if err != nil {
		return f, err
	}

	for _, e := range parsedcsr.Extensions {
		if e.Id.String() == puppetPSKoid {
			r, err := regexp.Compile("([a-zA-Z0-9_\\-\\.]+)")
			if err != nil {
				log.Fatalf("Unable to compile psk regex: " + err.Error())
			}

			match := r.FindStringSubmatch(string(e.Value))
			if len(match) > 0 {
				return match[1], nil
			}
		}
	}

	return f, errors.New("Certificate did not contain a PSK")
}
