package autosignr

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// The OID for Puppet's pp_preshared_key in the Certificate Extensions
// https://docs.puppetlabs.com/puppet/latest/reference/ssl_attributes_extensions.html
const puppetPSKoid string = "1.3.6.1.4.1.34380.1.1.4"

// ErrNoPemData is returned when the data expected to be a PEM encoded cert is not actually a cert
var ErrNoPemData = errors.New("no PEM data found in block")

// ErrNoPSK is returned when a certificate does not contain a preshared key
var ErrNoPSK = errors.New("certificate did not contain a PSK")

// regexpPSK limits what the PSK can contain to alphanumeric chars plus '_', '-', and '.'
var regexpPSK = regexp.MustCompile("([a-zA-Z0-9_\\-\\.]+)")

// CertnameFromFilename returns the name of a cert given the path to the file
func CertnameFromFilename(file string) string {
	return strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
}

// CheckCert checks if the cert is valid
func CheckCert(conf *Config, file string) (bool, error) {
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

// ValidateCert validates the cert
func ValidateCert(conf *Config, data []byte, certname string) (bool, error) {
	log.Debugf("ValidateCert %s", certname)
	if conf.CheckPSK {
		psk, err := PuppetPSKFromCSR(data)
		if err != nil {
			log.WithFields(log.Fields{
				"certname": certname,
				"err":      err,
			}).Warning("psk-extract-error")
			return false, err
		}
		if _, ok := conf.PresharedKeys[psk]; !ok {
			log.WithFields(log.Fields{
				"certname": certname,
				"psk":      psk,
			}).Warning("invalid-psk")
			return false, errors.New("Invalid PSK")
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

// SignCert will run the puppet command to sign the cert
func SignCert(conf *Config, certname string) {
	cmd := fmt.Sprintf(conf.CmdSign, certname)
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

// ExistingCerts checks existing certs in directory
func ExistingCerts(conf *Config) error {
	matches, err := filepath.Glob(fmt.Sprintf("%s/*.pem", conf.Dir))
	if err != nil {
		return errors.Wrap(err, "globbing")
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

	return nil
}

// PuppetPSKFromCSRFile return the CSR file data
func PuppetPSKFromCSRFile(file string) (string, error) {
	var f string

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return f, err
	}

	return PuppetPSKFromCSR(data)
}

// PuppetPSKFromCSR decodes and parses the cert data
func PuppetPSKFromCSR(data []byte) (string, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return "", ErrNoPemData
	}

	parsedcsr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return "", err
	}

	for _, e := range parsedcsr.Extensions {
		if e.Id.String() == puppetPSKoid {
			match := regexpPSK.FindStringSubmatch(string(e.Value))
			if len(match) > 0 {
				return match[1], nil
			}
		}
	}

	return "", ErrNoPSK
}
