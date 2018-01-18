// Copyright (C) 2012-2017 Miquel Sabaté Solà <mikisabate@gmail.com>
// This file is licensed under the MIT license.
// See the LICENSE file.

package useragent

import (
	//"fmt"
	"regexp"
	"strings"
)

// Represents full information on the operating system extracted from the user agent.
type OSInfo struct {
	// Full name of the operating system. This is identical to the output of ua.OS()
	FullName string

	// Name of the operating system. This is sometimes a shorter version of the
	// operating system name, e.g. "Mac OS X" instead of "Intel Mac OS X"
	Name string

	// Operating system version, e.g. 7 for Windows 7 or 10.8 for Max OS X Mountain Lion
	Version string
}

var reg *regexp.Regexp = regexp.MustCompile(`([0-9]*)_`)

// Normalize the name of the operating system. By now, this just
// affects to Windows NT.
//
// Returns a string containing the normalized name for the Operating System.
func normalizeOS(name string) string {
	sp := strings.SplitN(name, " ", 3)
	if len(sp) != 3 || sp[1] != "NT" {
		if sp[0] == "CPU" {
			var osName string
			if len(sp) == 3 {
				idx := strings.Index(sp[2], "like")
				if idx > 0 {
					os := strings.Join([]string{sp[1], sp[2][:idx]}, " ")
					osName = strings.TrimSpace(os)
				}
			} else {
				osName = strings.Join(sp[1:], " ")
			}

			osName = reg.ReplaceAllString(osName, "${1}.")
			osName = strings.Replace(osName, "iPhone OS", "IOS", 1)
			return osName
		}
		return name
	}

	switch sp[2] {
	case "5.0":
		return "Windows 2000"
	case "5.01":
		return "Windows 2000, Service Pack 1 (SP1)"
	case "5.1":
		return "Windows XP"
	case "5.2":
		return "Windows XP x64 Edition"
	case "6.0":
		return "Windows Vista"
	case "6.1":
		return "Windows 7"
	case "6.2":
		return "Windows 8"
	case "6.3":
		return "Windows 8.1"
	case "10.0":
		return "Windows 10"
	}
	return name
}

// Guess the OS, the localization and if this is a mobile device for a
// Webkit-powered browser.
//
// The first argument p is a reference to the current UserAgent and the second
// argument is a slice of strings containing the comment.
func webkit(p *UserAgent, comment []string) {
	p.devicetype = DeviceDesktop
	switch p.platform {
	case "iPod":
		p.mobile = true
		p.devicetype = DeviceMobile
		p.mobilebrand = "Apple"
		p.mobilemodel = "iPod"
	case "iPhone":
		p.mobile = true
		p.devicetype = DeviceMobile
		p.mobilebrand = "Apple"
		p.mobilemodel = "iPhone"
	case "iPad":
		p.mobile = true
		p.devicetype = DeviceTablet
		p.mobilebrand = "Apple"
		p.mobilemodel = "iPad"
	}

	if p.platform == "webOS" {
		p.browser.Name = p.platform
		p.os = "Palm"
		if len(comment) > 2 {
			p.localization = comment[2]
		}
		p.mobile = true
		p.devicetype = DeviceMobile
	} else if p.platform == "Symbian" {
		p.mobile = true
		p.devicetype = DeviceMobile
		p.browser.Name = p.platform
		p.os = comment[0]
	} else if p.platform == "Linux" {
		p.mobile = true
		p.devicetype = DeviceMobile
		if p.browser.Name == "Safari" {
			p.browser.Name = "Android"
		}
		if len(comment) > 1 {
			if comment[1] == "U" {
				if len(comment) > 2 {
					p.os = comment[2]
				} else {
					p.mobile = false
					p.devicetype = DeviceDesktop
					p.os = comment[0]
				}
			} else {
				p.os = comment[1]
			}
		}
		if len(comment) > 3 {
			p.localization = comment[3]
		}
		//		if len(comment) > 4 {
		//			model := strings.Split(comment[4], " ")
		//			if len(model) > 1 {
		//				p.mobilebrand = model[0]
		//				p.mobilemodel = model[1]
		//			}
		//		}
	} else if len(comment) > 0 {
		if len(comment) > 3 {
			p.localization = comment[3]
		}
		if strings.HasPrefix(comment[0], "Windows NT") {
			p.os = normalizeOS(comment[0])
		} else if len(comment) < 2 {
			p.localization = comment[0]
			p.devicetype = DeviceUnknown
		} else if len(comment) < 3 {
			if !p.googleBot() {
				p.os = normalizeOS(comment[1])
			}
		} else {
			p.os = normalizeOS(comment[2])
		}
		if p.platform == "BlackBerry" {
			p.browser.Name = p.platform
			p.devicetype = DeviceMobile
			p.mobilebrand = "BlackBerry"
			if p.os == "Touch" {
				p.os = p.platform
			}
		}
	}
}

// Guess the OS, the localization and if this is a mobile device
// for a Gecko-powered browser.
//
// The first argument p is a reference to the current UserAgent and the second
// argument is a slice of strings containing the comment.
func gecko(p *UserAgent, comment []string) {
	p.devicetype = DeviceDesktop
	if len(comment) > 1 {
		if comment[1] == "U" {
			if len(comment) > 2 {
				p.os = normalizeOS(comment[2])
			} else {
				p.os = normalizeOS(comment[1])
			}
		} else {
			if p.platform == "Android" {
				p.mobile = true
				p.devicetype = DeviceMobile
				if comment[0] == "Tablet" || comment[1] == "Tablet" {
					p.devicetype = DeviceTablet
				}
				p.platform, p.os = normalizeOS(comment[1]), p.platform
			} else if comment[0] == "Mobile" {
				p.mobile = true
				p.devicetype = DeviceMobile
				p.os = "FirefoxOS"
			} else if comment[0] == "Tablet" {
				p.mobile = true
				p.devicetype = DeviceTablet
				p.os = "FirefoxOS"
			} else {
				if p.os == "" {
					p.os = normalizeOS(comment[1])
				}
			}
		}
		// Only parse 4th comment as localization if it doesn't start with rv:.
		// For example Firefox on Ubuntu contains "rv:XX.X" in this field.
		if len(comment) > 3 && !strings.HasPrefix(comment[3], "rv:") {
			p.localization = comment[3]
		}
	}
}

// Guess the OS, the localization and if this is a mobile device
// for Internet Explorer.
//
// The first argument p is a reference to the current UserAgent and the second
// argument is a slice of strings containing the comment.
func trident(p *UserAgent, comment []string) {
	// Internet Explorer only runs on Windows.
	p.platform = "Windows"
	p.devicetype = DeviceDesktop

	// The OS can be set before to handle a new case in IE11.
	if p.os == "" {
		if len(comment) > 2 {
			p.os = normalizeOS(comment[2])
		} else {
			p.os = "Windows NT 4.0"
		}
	}

	// Last but not least, let's detect if it comes from a mobile device.
	hasModel := false
	for _, v := range comment {
		if hasModel {
			p.mobilemodel = v
			hasModel = false
			continue
		}
		if strings.HasPrefix(v, "IEMobile") || strings.HasPrefix(v, "Windows Phone") {
			//TODO Windows Phone是否包含平板，还不清楚
			p.mobile = true
			p.devicetype = DeviceMobile
		}
		if strings.HasPrefix(v, "Tablet") || strings.HasPrefix(v, "Touch") {
			p.mobile = true
			p.devicetype = DeviceTablet
		}
		switch v {
		case "SAMSUNG":
			p.mobilebrand = "SAMSUNG"
			hasModel = true
		}
	}
}

// Guess the OS, the localization and if this is a mobile device
// for Opera.
//
// The first argument p is a reference to the current UserAgent and the second
// argument is a slice of strings containing the comment.
func opera(p *UserAgent, comment []string) {
	p.devicetype = DeviceDesktop
	slen := len(comment)

	if strings.HasPrefix(comment[0], "Windows") {
		p.platform = "Windows"
		p.os = normalizeOS(comment[0])
		if slen > 2 {
			if slen > 3 && strings.HasPrefix(comment[2], "MRA") {
				p.localization = comment[3]
			} else {
				p.localization = comment[2]
			}
		}
	} else {
		if strings.HasPrefix(comment[0], "Android") {
			p.mobile = true
			p.devicetype = DeviceMobile
		}
		p.platform = comment[0]
		if slen > 1 {
			p.os = comment[1]
			if slen > 3 {
				p.localization = comment[3]
			}
		} else {
			p.os = comment[0]
		}
	}
}

// Guess the OS. Android browsers send Dalvik as the user agent in the
// request header.
//
// The first argument p is a reference to the current UserAgent and the second
// argument is a slice of strings containing the comment.
func dalvik(p *UserAgent, comment []string) {
	slen := len(comment)

	if strings.HasPrefix(comment[0], "Linux") {
		p.platform = comment[0]
		if slen > 2 {
			p.os = comment[2]
		}
		p.mobile = true
	}
}

// Given the comment of the first section of the UserAgent string,
// get the platform.
func getPlatform(comment []string) string {
	if len(comment) > 0 {
		if comment[0] != "compatible" {
			if strings.HasPrefix(comment[0], "Windows") {
				return "Windows"
			} else if strings.HasPrefix(comment[0], "Symbian") {
				return "Symbian"
			} else if strings.HasPrefix(comment[0], "webOS") {
				return "webOS"
			} else if comment[0] == "BB10" {
				return "BlackBerry"
			}
			return comment[0]
		}
	}
	return ""
}

// Detect some properties of the OS from the given section.
func (p *UserAgent) detectOS(s section) {
	for _, v := range s.comment {
		if !strings.Contains(v, "Build") {
			continue
		}
		svs := strings.Split(v, " ")
		for i, sv := range svs {
			if strings.HasPrefix(sv, "Build") {
				if i > 1 {
					p.mobilebrand = svs[0]
					p.mobilemodel = svs[1]
				} else if i == 1 {
					p.mobilemodel = svs[0]
				}
				break
			}
		}
	}

	if s.name == "Mozilla" {
		// Get the platform here. Be aware that IE11 provides a new format
		// that is not backwards-compatible with previous versions of IE.
		p.platform = getPlatform(s.comment)
		if p.platform == "Windows" && len(s.comment) > 0 {
			p.os = normalizeOS(s.comment[0])
		}
		//fmt.Printf("s:%#+v\n", s)
		//fmt.Printf("p:%+v\n", *p)

		// And finally get the OS depending on the engine.
		switch p.browser.Engine {
		case "":
			fallthrough
		case "SkypeUriPreview":
			p.undecided = true
		case "Gecko": // Firefox,...
			gecko(p, s.comment)
		case "AppleWebKit": // Safari,Chrome,...
			webkit(p, s.comment)
		case "EdgeHTML": // Edge,...
			fallthrough
		case "Trident": // IE,...
			trident(p, s.comment)
		}
	} else if s.name == "Opera" {
		// Opera
		//fmt.Printf("Opera s:%#+v\n", s)
		//fmt.Printf("Opera p:%+v\n", *p)
		if len(s.comment) > 0 {
			opera(p, s.comment)
		}
	} else if s.name == "Dalvik" {
		// Android VM
		//fmt.Printf("s:%#+v\n", s)
		//fmt.Printf("p:%+v\n", *p)
		p.devicetype = DeviceMobile
		if len(s.comment) > 0 {
			dalvik(p, s.comment)
		}
	} else {
		// Check whether this is a bot or just a weird browser.
		p.undecided = true
	}
}

// Returns a string containing the platform..
func (p *UserAgent) Platform() string {
	return p.platform
}

// Returns a string containing the name of the Operating System.
func (p *UserAgent) OS() string {
	return p.os
}

// Returns a string containing the localization.
func (p *UserAgent) Localization() string {
	return p.localization
}

// Return OS name and version from a slice of strings created from the full name of the OS.
func osName(osSplit []string) (name, version string) {
	if len(osSplit) == 1 {
		name = osSplit[0]
		version = ""
	} else {
		// Assume version is stored in the last part of the array.
		nameSplit := osSplit[:len(osSplit)-1]
		version = osSplit[len(osSplit)-1]

		// Nicer looking Mac OS X
		if len(nameSplit) >= 2 && nameSplit[0] == "Intel" && nameSplit[1] == "Mac" {
			nameSplit = nameSplit[1:]
		}
		name = strings.Join(nameSplit, " ")

		if strings.Contains(version, "x86") || strings.Contains(version, "i686") {
			// x86_64 and i868 are not Linux versions but architectures
			version = ""
		} else if version == "X" && name == "Mac OS" {
			// X is not a version for Mac OS.
			name = name + " " + version
			version = ""
		}
	}
	return name, version
}

// Returns combined information for the operating system.
func (p *UserAgent) OSInfo() OSInfo {
	// Special case for iPhone weirdness
	os := strings.Replace(p.os, "like Mac OS X", "", 1)
	os = strings.Replace(os, "CPU", "", 1)
	os = strings.Trim(os, " ")

	osSplit := strings.Split(os, " ")

	// Special case for x64 edition of Windows
	if os == "Windows XP x64 Edition" {
		osSplit = osSplit[:len(osSplit)-2]
	}

	name, version := osName(osSplit)

	// Special case for names that contain a forward slash version separator.
	if strings.Contains(name, "/") {
		s := strings.Split(name, "/")
		name = s[0]
		version = s[1]
	}

	// Special case for versions that use underscores
	version = strings.Replace(version, "_", ".", -1)

	return OSInfo{
		FullName: p.os,
		Name:     name,
		Version:  version,
	}
}

func (p *UserAgent) MobileBrand() string {
	return p.mobilebrand
}

func (p *UserAgent) MobileModel() string {
	return p.mobilemodel
}

func (p *UserAgent) DeviceType() string {
	return p.devicetype
}
