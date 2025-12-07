package sshUtils

import (
	"log/slog"
	"maps"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"unicode"
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

func isForwardX11Valid(forwardX11 string) bool {
	return yesNoOptionValid(forwardX11)
}

func isForwardX11TimeoutValid(forwardX11Timeout string) bool {
	return validSSHTime(forwardX11Timeout)
}

func isHostKeyAliasValid(hostKeyAlias string) bool {
	return len(hostKeyAlias) > 0
}

func isKbdInteractiveAuthenticationValid(kbdInteractiveAuthentication string) bool {
	return yesNoOptionValid(kbdInteractiveAuthentication)
}

func isPasswordAuthenticationValid(passwordAuthentication string) bool {
	return yesNoOptionValid(passwordAuthentication)
}

// splitForwardSpec splits a forwarding spec supports ipv6 host blocks.
// Example: "[::1]:8080:[2001:db8::1]:22"
func splitForwardSpec(s string) []string {
	var parts []string
	var strBuilder strings.Builder
	inBrackets := false

	for _, r := range s {
		switch r {
		case '[':
			inBrackets = true
			strBuilder.WriteRune(r)
		case ']':
			inBrackets = false
			strBuilder.WriteRune(r)
		case ':':
			if inBrackets {
				strBuilder.WriteRune(r)
			} else {
				parts = append(parts, strBuilder.String())
				strBuilder.Reset()
			}
		default:
			strBuilder.WriteRune(r)
		}
	}
	parts = append(parts, strBuilder.String())

	return parts
}

func isValidPort(port string) bool {
	n, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return n > 0 && n <= 65535
}

func isValidHostIP(h string) bool {
	if strings.HasPrefix(h, "[") && strings.HasSuffix(h, "]") {
		ip, err := netip.ParseAddr(h[1 : len(h)-1])
		if err != nil {
			return false
		}
		return ip.Is6()
	}

	// check if ipv4
	// needs to be in form of quads so 255.255.255.255 can drop leading zeros
	ip, err := netip.ParseAddr(h)
	if err == nil && ip.Is4() {
		return true
	}
	return false
}

func isValidHostname(h string) bool {
	// total host length has to be less than 254 chars
	if len(h) == 0 || len(h) > 253 {
		return false
	}
	// edge case remove trailing/ending dot in hostname
	if h[len(h)-1] == '.' {
		h = h[:len(h)-1]
	}
	parts := strings.Split(h, ".") // get all parts of the hostname
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		if len(part) > 63 {
			return false
		}

		if !(unicode.IsDigit(rune(part[0])) || unicode.IsLetter(rune(part[0]))) ||
			!(unicode.IsDigit(rune(part[len(part)-1])) || unicode.IsLetter(rune(part[len(part)-1]))) {
			return false
		}
		for _, char := range part {
			if unicode.IsLetter(char) || unicode.IsDigit(char) {
				continue
			}
			if char == '-' {
				continue
			}
			return false
		}
	}
	return true
}

func ValidHost(h string) bool {
	return isValidHostIP(h) || isValidHostname(h)
}

func isLocalForwardValid(localForward string) bool {
	// so we should either get [bindAddr]:port:[bindAddr]:port
	// or port:bindAddr:port
	parts := splitForwardSpec(localForward)
	if len(parts) != 3 && len(parts) != 4 {
		return false
	}
	if len(parts) == 3 { //short for ie port:host:port
		return isValidPort(parts[0]) && ValidHost(parts[1]) && isValidPort(parts[2])
	} else {
		return ValidHost(parts[0]) && isValidPort(parts[1]) && ValidHost(parts[2]) && isValidPort(parts[3])
	}
}

func isRemoteForwardValid(remoteForward string) bool {
	parts := splitForwardSpec(remoteForward)
	if len(parts) != 3 && len(parts) != 4 {
		return false
	}
	if len(parts) == 3 { //short for ie port:host:port
		return isValidPort(parts[0]) && ValidHost(parts[1]) && isValidPort(parts[2])
	} else {
		return ValidHost(parts[0]) && isValidPort(parts[1]) && ValidHost(parts[2]) && isValidPort(parts[3])
	}
}

func isDynamicForwardValid(dynamicForward string) bool {
	parts := splitForwardSpec(dynamicForward)
	if len(parts) > 1 {
		return false
	}
	if len(parts) == 1 {
		return ValidHost(parts[0]) && isValidPort(parts[1])
	} else {
		return isValidPort(parts[0])
	}
}
