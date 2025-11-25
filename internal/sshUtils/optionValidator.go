package sshUtils

import (
	"maps"
	"slices"
)

var validOptionsToBeSet = map[string]struct{}{
	"Hostname":                     {},
	"Port":                         {},
	"User":                         {},
	"IdentityFile":                 {},
	"IgnoreUnknown":                {},
	"Include":                      {},
	"IPQoS":                        {},
	"AddressFamily":                {},
	"BatchMode":                    {},
	"BindAddress":                  {},
	"CertificateFile":              {},
	"ChannelTimeout":               {},
	"CheckHostIP":                  {},
	"Compression":                  {},
	"ConnectionAttempts":           {},
	"ConnectTimeout":               {},
	"DynamicForward":               {},
	"ForwardX11":                   {},
	"ForwardX11Timeout":            {},
	"GlobalKnownHostsFile":         {},
	"HostKeyAlgorithms":            {},
	"HostKeyAlias":                 {},
	"KbdInteractiveAuthentication": {},
	"LocalForward":                 {},
	"PasswordAuthentication":       {},
	"RemoteForward":                {},
}

var mutiOptTypes = map[string]struct{}{
	"RemoteForward":  {},
	"LocalForward":   {},
	"IdentityFile":   {},
	"DynamicForward": {},
}

func IsAcceptableOption(opt string) bool {
	_, ok := validOptionsToBeSet[opt]
	return ok
}

func OptionIsOfMutiType(opt string) bool {
	_, ok := mutiOptTypes[opt]
	return ok
}

func GetListOfAcceptableOptions() []string {

	return slices.Sorted(maps.Keys(validOptionsToBeSet))
}

// todo add validators for common options
