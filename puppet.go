package autosignr

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

// Runs one or more commands to take when de-provisioning a node for Puppet
func CleanNode(commands []string, certname string) {

	for _, cmd := range commands {

		command := fmt.Sprintf(cmd, certname)
		pieces := strings.Split(command, " ")

		cmdOut, err := exec.Command(pieces[0], pieces[1:]...).CombinedOutput()
		if err != nil {
			log.WithFields(log.Fields{
				"certname": certname,
				"err":      err,
				"output":   string(cmdOut),
				"command":  cmd,
			}).Error("clean-node-command-failure")
			return
		}
	}

	log.WithFields(log.Fields{
		"certname": certname,
	}).Info("cleanup-success")
}
