package sshUtils

import (
	"log/slog"
	"maps"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"strings"
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
	"BindInterface":                {},
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

var timeQualifierMap = map[string]struct{}{
	"s": {},
	"m": {},
	"h": {},
	"d": {},
	"w": {},
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

func isAddressFamilyValid(family string) bool {
	return family == "any" || family == "inet" || family == "inet6"
}

func yesNoOptionValid(v string) bool {
	return v == "yes" || v == "no"
}

func validSSHTime(timeString string) bool {
	// man page excerpt
	/*	where time is a positive integer value and
		qualifier is one of the following:

		⟨none⟩  seconds
		s | S   seconds
		m | M   minutes
		h | H   hours
		d | D   days
		w | W   weeks

		Each member of the sequence is added together to calculate the
		total time value.

			Time format examples:

		600     600 seconds (10 minutes)
		10m     10 minutes
		1h30m   1 hour 30 minutes (90 minutes)
	*/
	// valid formats are "" "s" "m" "h" "d"
	timeString = strings.ToLower(timeString)
	timeString = strings.TrimSpace(timeString)
	digits := timeString[0 : len(timeString)-1]
	timeFormat := string(timeString[len(timeString)-1])
	_, err := strconv.Atoi(digits)
	if err != nil {
		return false
	}
	_, ok := timeQualifierMap[timeFormat]
	return ok
}

func isBatchModeValid(mode string) bool {
	return yesNoOptionValid(mode)
}

func isBindAddressValid(bindAddress string) bool {
	if bindAddress == "localhost" {
		return true
	}
	addr, err := netip.ParseAddr(bindAddress)
	if err != nil {
		return false
	}
	machineAdders, err := net.InterfaceAddrs()
	if err != nil {
		slog.Warn("Unable to get machine interfaces, unable to verify that address is valid, may give a false positive")
		return true
	}
	for _, machineAddr := range machineAdders {
		if addr.String() == machineAddr.String() {
			return true
		}
	}
	return false
}

func isBindInterfaceValid(bindInterface string) bool {
	machineInterface, err := net.InterfaceByName(bindInterface)
	if err != nil {
		return false
	}
	if machineInterface == nil {
		return false
	}
	return true
}

func isCompressionModeValid(mode string) bool {
	return yesNoOptionValid(mode)
}

// todo finish other validation methods

func isChannelTimeoutValid(channelTimeout string) bool {
	//timeouts as quoted from manpage is seperated by whitespace and in the form of type=interval
	pairs := strings.Split(channelTimeout, " ")
	for _, pair := range pairs {
		if !strings.Contains(pair, "=") {
			return false
		}
		k, v := strings.Split(pair, "=")[0], strings.Split(pair, "=")[1]
		if len(k) == 0 || len(v) == 0 {
			return false
		}
	}
	return true
}

func isCheckHostIPValid(checkHostIP string) bool {
	return yesNoOptionValid(checkHostIP)
}

func isConnectionAttemptsValid(connectionAttempts string) bool {
	_, err := strconv.Atoi(connectionAttempts)
	if err != nil {
		return false
	}
	return true
}

func isConnectTimeoutValid(connectTimeout string) bool {
	_, err := strconv.Atoi(connectTimeout)
	if err != nil {
		return false
	}
	return true
}

func isDynamicForwardValid(dynamicForward string) bool {

	// todo validate this input
	panic("implement me")
}

func isForwardX11Valid(forwardX11 string) bool {
	return yesNoOptionValid(forwardX11)
}

func isForwardX11TimeoutValid(forwardX11Timeout string) bool {
	return validSSHTime(forwardX11Timeout)
}

// todo continue more validation after this point
