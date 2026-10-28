package main

import (
	"energyshares-job-scheduler/cmd"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	cmd.Execute()
}
