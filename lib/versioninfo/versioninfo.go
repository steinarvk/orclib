package versioninfo

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	ProgramNameBase64         string
	ManualVersionNumberBase64 string
	BuildTimestampBase64      string
	BuildMachineBase64        string
	GitDescribeBase64         string
)

var ProgramName string
var ManualVersionNumber string
var BuildTimestamp string
var BuildMachine string
var GitDescribe string

func decode(target *string, s string) {
	if s == "" {
		return
	}
	rv, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		logrus.Errorf("Invalid versioninfo: %q", s)
		return
	}
	*target = strings.TrimSpace(string(rv))
}

func init() {
	decode(&ProgramName, ProgramNameBase64)
	decode(&ManualVersionNumber, ManualVersionNumberBase64)
	decode(&BuildTimestamp, BuildTimestampBase64)
	decode(&BuildMachine, BuildMachineBase64)
	decode(&GitDescribe, GitDescribeBase64)
}

func MakeVersion() string {
	if ManualVersionNumber != "" {
		return ManualVersionNumber
	}
	return fmt.Sprintf("0.0.%s+%s", BuildTimestamp, GitDescribe)
}

func MakeFields() logrus.Fields {
	return logrus.Fields{
		"Program":        ProgramName,
		"Version":        MakeVersion(),
		"GitInfo":        GitDescribe,
		"BuildTimestamp": BuildTimestamp,
		"BuildMachine":   BuildMachine,
	}
}

func MakeJSON() interface{} {
	return map[string]interface{}{
		"Program":        ProgramName,
		"Version":        MakeVersion(),
		"GitInfo":        GitDescribe,
		"BuildTimestamp": BuildTimestamp,
		"BuildMachine":   BuildMachine,
	}
}

func UserAgent() string {
	return fmt.Sprintf("%s %s", ProgramName, MakeVersion())
}

func VersionInfoLines() []string {
	var rv []string

	rv = append(rv, fmt.Sprintf("Version: %s", MakeVersion()))

	if GitDescribe != "" {
		rv = append(rv, fmt.Sprintf("Git: %s", GitDescribe))
	}

	n, err := strconv.Atoi(BuildTimestamp)
	if err == nil {
		rv = append(rv, fmt.Sprintf("Build time: %d [%v]", n, time.Unix(int64(n), 0)))
	}

	if BuildMachine != "" {
		rv = append(rv, fmt.Sprintf("Build machine: %s", BuildMachine))
	}

	return rv
}
