package request

import (
	"strconv"
	"strings"

	"github.com/avct/uasurfer"
)

// Define Device type constants
const desktop = "desktop"
const mobile = "mobile"
const tablet = "tablet"

// Device struct returns the information about the device by parsing the user agent
type Device struct {
	Type, Browser string
	OS, OSName    string
	OSVersion     string
}

func getBrowser(uaInfo *uasurfer.UserAgent) string {
	var browser = uaInfo.Browser.Name.String()
	if len(browser) > 7 { // because device.Browser = "BrowserChrome" remove the prefix
		browser = browser[7:]
	}
	var version = uaInfo.Browser.Version.Major
	if version > 0 {
		browser += " " + strconv.Itoa(version)
	}
	return browser
}

func getOS(uaInfo *uasurfer.UserAgent) (string, string, string) {
	var os = uaInfo.OS.Name.String()
	if len(os) > 2 { // because device.OS = "OSMacOSX" remove the prefix "OS"
		os = os[2:]
	}
	if os == "MacOSX" {
		os = "Mac OS"
	}
	var osName = os
	var majorVersion = uaInfo.OS.Version.Major
	var minorVersion = uaInfo.OS.Version.Minor
	var osVersion = strconv.Itoa(majorVersion) + "." + strconv.Itoa(minorVersion)
	os += " " + osVersion
	return os, osName, osVersion
}

func getDeviceType(uaInfo *uasurfer.UserAgent) string {
	var deviceType = uaInfo.DeviceType.String()
	deviceType = deviceType[6:]

	if deviceType == "Computer" {
		deviceType = desktop
	} else if deviceType == "Phone" {
		deviceType = mobile
	}

	return strings.ToLower(deviceType)
}

// ParseUA function returns a device struct containing the parsed user agent info
func ParseUA(ua string) *Device {
	var uaInfo = uasurfer.Parse(ua)

	var device = &Device{}
	device.Type = getDeviceType(uaInfo)
	device.Browser = getBrowser(uaInfo)
	device.OS, device.OSName, device.OSVersion = getOS(uaInfo)

	return device
}
