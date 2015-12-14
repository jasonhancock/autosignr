package autosignr

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// The OID for Puppet's pp_preshared_key in the Certificate Extensions
// https://docs.puppetlabs.com/puppet/latest/reference/ssl_attributes_extensions.html
const puppetPSKoid string = "1.3.6.1.4.1.34380.1.1.4"

func CheckCert(conf Config, file string) {
	base := filepath.Base(file)
	extension := filepath.Ext(base)
	if extension == ".pem" {
		var name = base[0 : len(base)-len(extension)]

		if conf.CheckPSK {
			psk, err := PuppetPSKFromCSR(file)
			if err != nil {
				log.WithFields(log.Fields{
					"certname": name,
					"err":      err,
				}).Warning("psk-extract-error")
			} else {
				if _, ok := conf.PresharedKeys[psk]; !ok {
					log.WithFields(log.Fields{
						"certname": name,
						"psk":      psk,
					}).Warning("invalid-psk")
				}

				return
			}
		}

		result := false
		for _, acct := range conf.Accounts {
			result = acct.Check(name)
			if result {
				SignCert(conf, name)
				break
			}
		}
		if !result {
			log.Warningf("Unable to validate instance %s", name)
		}
	}
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
		CheckCert(conf, cert)
	}
}

func PuppetPSKFromCSR(file string) (string, error) {
	var f string

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return f, err
	}

	block, _ := pem.Decode(data)
	parsedcsr, err := x509.ParseCertificateRequest(block.Bytes)

	for _, e := range parsedcsr.Extensions {
		if e.Id.String() == puppetPSKoid {
			// the first char of the trimmed string is ASCII 22,
			// synchronous idle, so remove that too
			f = strings.TrimSpace(string(e.Value))[1:]
			return f, nil
		}
	}

	return f, errors.New("Certificate did not contain a PSK")
}
