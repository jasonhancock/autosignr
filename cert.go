package autosignr

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func CheckCert(conf Config, file string) {
	base := filepath.Base(file)
	extension := filepath.Ext(base)
	if extension == ".pem" {
		var name = base[0 : len(base)-len(extension)]
		result := false
		for _, acct := range conf.Accounts {
			result = acct.Check(name)
			if result {
				SignCert(conf, name)
				break
			}
		}
		if !result {
			log.Printf("Unable to validate instance %s", name)
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
