// Copyright (C) 2012-2017 Miquel Sabaté Solà <mikisabate@gmail.com>
// This file is licensed under the MIT license.
// See the LICENSE file.

package useragent

import (
	"fmt"
	"reflect"
	"testing"
)

// Slice that contains all the tests. Each test is contained in a struct
// that groups the title of the test, the User-Agent string to be tested and the expected value.
var uastrings = []struct {
	title      string
	ua         string
	expected   string
	expectedOS *OSInfo
}{
	// Bots
	{
		title:    "SkypeUriPreview",
		ua:       "Mozilla/5.0 (Windows NT 6.1; WOW64) SkypeUriPreview Preview/0.5",
		expected: "Mozilla:5.0 Platform:Windows OS:Unknown Browser:SkypeUriPreview Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "GoogleBot",
		ua:       "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		expected: "Mozilla:5.0 OS:Unknown Browser:Googlebot-2.1 Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "GoogleBotSmartphone",
		ua:       "Mozilla/5.0 (iPhone; CPU iPhone OS 6_0 like Mac OS X) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Mobile/10A5376e Safari/8536.25 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		expected: "Mozilla:5.0 OS:Unknown Browser:Googlebot-2.1 Engine:Unknown Bot:true Mobile:true DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "BingBot",
		ua:       "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
		expected: "Mozilla:5.0 OS:Unknown Browser:bingbot-2.0 Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "BaiduBot",
		ua:       "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
		expected: "Mozilla:5.0 OS:Unknown Browser:Baiduspider-2.0 Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "Twitterbot",
		ua:       "Twitterbot",
		expected: "Mozilla:Unknown OS:Unknown Browser:Twitterbot Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "YahooBot",
		ua:       "Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)",
		expected: "Mozilla:5.0 OS:Unknown Browser:Yahoo! Slurp Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "FacebookExternalHit",
		ua:       "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
		expected: "Mozilla:Unknown OS:Unknown Browser:facebookexternalhit-1.1 Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "FacebookPlatform",
		ua:       "facebookplatform/1.0 (+http://developers.facebook.com)",
		expected: "Mozilla:Unknown OS:Unknown Browser:facebookplatform-1.0 Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "FaceBot",
		ua:       "Facebot",
		expected: "Mozilla:Unknown OS:Unknown Browser:Facebot Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "NutchCVS",
		ua:       "NutchCVS/0.8-dev (Nutch; http://lucene.apache.org/nutch/bot.html; nutch-agent@lucene.apache.org)",
		expected: "Mozilla:Unknown OS:Unknown Browser:NutchCVS Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "MJ12bot",
		ua:       "Mozilla/5.0 (compatible; MJ12bot/v1.2.4; http://www.majestic12.co.uk/bot.php?+)",
		expected: "Mozilla:5.0 OS:Unknown Browser:MJ12bot-v1.2.4 Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "MJ12bot",
		ua:       "MJ12bot/v1.0.8 (http://majestic12.co.uk/bot.php?+)",
		expected: "Mozilla:Unknown OS:Unknown Browser:MJ12bot Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "AhrefsBot",
		ua:       "Mozilla/5.0 (compatible; AhrefsBot/4.0; +http://ahrefs.com/robot/)",
		expected: "Mozilla:5.0 OS:Unknown Browser:AhrefsBot-4.0 Engine:Unknown Bot:true Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},

	// Internet Explorer
	{
		title:      "IE10",
		ua:         "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; Trident/6.0)",
		expected:   "Mozilla:5.0 Platform:Windows OS:Windows 8 Browser:Internet Explorer-10.0 Engine:Trident Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Windows 8", "Windows", "8"},
	},
	{
		title:    "Tablet",
		ua:       "Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.2; ARM; Trident/6.0; Touch; .NET4.0E; .NET4.0C; Tablet PC 2.0)",
		expected: "Mozilla:4.0 Platform:Windows OS:Windows 8 Browser:Internet Explorer-10.0 Engine:Trident Bot:false Mobile:true DeviceType:Tablet Brand:Unknown Model:Unknown",
	},
	{
		title:    "Touch",
		ua:       "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; ARM; Trident/6.0; Touch)",
		expected: "Mozilla:5.0 Platform:Windows OS:Windows 8 Browser:Internet Explorer-10.0 Engine:Trident Bot:false Mobile:true DeviceType:Tablet Brand:Unknown Model:Unknown",
	},
	{
		title:      "Phone",
		ua:         "Mozilla/4.0 (compatible; MSIE 7.0; Windows Phone OS 7.0; Trident/3.1; IEMobile/7.0; SAMSUNG; SGH-i917)",
		expected:   "Mozilla:4.0 Platform:Windows OS:Windows Phone OS 7.0 Browser:Internet Explorer-7.0 Engine:Trident Bot:false Mobile:true DeviceType:Mobile Brand:SAMSUNG Model:SGH-i917",
		expectedOS: &OSInfo{"Windows Phone OS 7.0", "Windows Phone OS", "7.0"},
	},
	{
		title:      "IE6",
		ua:         "Mozilla/4.0 (compatible; MSIE6.0; Windows NT 5.0; .NET CLR 1.1.4322)",
		expected:   "Mozilla:4.0 Platform:Windows OS:Windows 2000 Browser:Internet Explorer-6.0 Engine:Trident Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Windows 2000", "Windows", "2000"},
	},
	{
		title:      "IE8Compatibility",
		ua:         "Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E; InfoPath.3; MS-RTC LM 8)",
		expected:   "Mozilla:4.0 Platform:Windows OS:Windows 7 Browser:Internet Explorer-8.0 Engine:Trident Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Windows 7", "Windows", "7"},
	},
	{
		title:    "IE10Compatibility",
		ua:       "Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.1; WOW64; Trident/6.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E; InfoPath.3; MS-RTC LM 8)",
		expected: "Mozilla:4.0 Platform:Windows OS:Windows 7 Browser:Internet Explorer-10.0 Engine:Trident Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:      "IE11Win81",
		ua:         "Mozilla/5.0 (Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko",
		expected:   "Mozilla:5.0 Platform:Windows OS:Windows 8.1 Browser:Internet Explorer-11.0 Engine:Trident Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Windows 8.1", "Windows", "8.1"},
	},
	{
		title:    "IE11Win7",
		ua:       "Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko",
		expected: "Mozilla:5.0 Platform:Windows OS:Windows 7 Browser:Internet Explorer-11.0 Engine:Trident Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "IE11b32Win7b64",
		ua:       "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko",
		expected: "Mozilla:5.0 Platform:Windows OS:Windows 7 Browser:Internet Explorer-11.0 Engine:Trident Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "IE11b32Win7b64MDDRJS",
		ua:       "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; MDDRJS; rv:11.0) like Gecko",
		expected: "Mozilla:5.0 Platform:Windows OS:Windows 7 Browser:Internet Explorer-11.0 Engine:Trident Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "IE11Compatibility",
		ua:       "Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.3; Trident/7.0)",
		expected: "Mozilla:4.0 Platform:Windows OS:Windows 8.1 Browser:Internet Explorer-7.0 Engine:Trident Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},

	// Microsoft Edge
	{
		title:      "EdgeDesktop",
		ua:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.10240",
		expected:   "Mozilla:5.0 Platform:Windows OS:Windows 10 Browser:Edge-12.10240 Engine:EdgeHTML Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Windows 10", "Windows", "10"},
	},
	{
		title:    "EdgeMobile",
		ua:       "Mozilla/5.0 (Windows Phone 10.0; Android 4.2.1; DEVICE INFO) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Mobile Safari/537.36 Edge/12.10240",
		expected: "Mozilla:5.0 Platform:Windows OS:Windows Phone 10.0 Browser:Edge-12.10240 Engine:EdgeHTML Bot:false Mobile:true DeviceType:Mobile Brand:Unknown Model:Unknown",
	},

	// Gecko
	{
		title:      "FirefoxMac",
		ua:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.6; rv:2.0b8) Gecko/20100101 Firefox/4.0b8",
		expected:   "Mozilla:5.0 Platform:Macintosh OS:Intel Mac OS X 10.6 Browser:Firefox-4.0b8 Engine:Gecko-20100101 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Intel Mac OS X 10.6", "Mac OS X", "10.6"},
	},
	{
		title:      "FirefoxMacLoc",
		ua:         "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.6; en-US; rv:1.9.2.13) Gecko/20101203 Firefox/3.6.13",
		expected:   "Mozilla:5.0 Platform:Macintosh OS:Intel Mac OS X 10.6 Localization:en-US Browser:Firefox-3.6.13 Engine:Gecko-20101203 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Intel Mac OS X 10.6", "Mac OS X", "10.6"},
	},
	{
		title:      "FirefoxLinux",
		ua:         "Mozilla/5.0 (X11; Linux x86_64; rv:17.0) Gecko/20100101 Firefox/17.0",
		expected:   "Mozilla:5.0 Platform:X11 OS:Linux x86_64 Browser:Firefox-17.0 Engine:Gecko-20100101 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Linux x86_64", "Linux", ""},
	},
	{
		title:      "FirefoxLinux - Ubuntu V50",
		ua:         "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:50.0) Gecko/20100101 Firefox/50.0",
		expected:   "Mozilla:5.0 Platform:X11 OS:Ubuntu Browser:Firefox-50.0 Engine:Gecko-20100101 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Ubuntu", "Ubuntu", ""},
	},
	{
		title:      "FirefoxWin",
		ua:         "Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.14) Gecko/20080404 Firefox/2.0.0.14",
		expected:   "Mozilla:5.0 Platform:Windows OS:Windows XP Localization:en-US Browser:Firefox-2.0.0.14 Engine:Gecko-20080404 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Windows XP", "Windows", "XP"},
	},
	{
		title:    "Firefox29Win7",
		ua:       "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:29.0) Gecko/20100101 Firefox/29.0",
		expected: "Mozilla:5.0 Platform:Windows OS:Windows 7 Browser:Firefox-29.0 Engine:Gecko-20100101 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:      "CaminoMac",
		ua:         "Mozilla/5.0 (Macintosh; U; Intel Mac OS X; en; rv:1.8.1.14) Gecko/20080409 Camino/1.6 (like Firefox/2.0.0.14)",
		expected:   "Mozilla:5.0 Platform:Macintosh OS:Intel Mac OS X Localization:en Browser:Camino-1.6 Engine:Gecko-20080409 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Intel Mac OS X", "Mac OS X", ""},
	},
	{
		title:      "Iceweasel",
		ua:         "Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.8.1) Gecko/20061024 Iceweasel/2.0 (Debian-2.0+dfsg-1)",
		expected:   "Mozilla:5.0 Platform:X11 OS:Linux i686 Localization:en-US Browser:Iceweasel-2.0 Engine:Gecko-20061024 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Linux i686", "Linux", ""},
	},
	{
		title:    "SeaMonkey",
		ua:       "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.6; en-US; rv:1.9.1.4) Gecko/20091017 SeaMonkey/2.0",
		expected: "Mozilla:5.0 Platform:Macintosh OS:Intel Mac OS X 10.6 Localization:en-US Browser:SeaMonkey-2.0 Engine:Gecko-20091017 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "AndroidFirefox",
		ua:       "Mozilla/5.0 (Android; Mobile; rv:17.0) Gecko/17.0 Firefox/17.0",
		expected: "Mozilla:5.0 Platform:Mobile OS:Android Browser:Firefox-17.0 Engine:Gecko-17.0 Bot:false Mobile:true DeviceType:Mobile Brand:Unknown Model:Unknown",
	},
	{
		title:      "AndroidFirefoxTablet",
		ua:         "Mozilla/5.0 (Android; Tablet; rv:26.0) Gecko/26.0 Firefox/26.0",
		expected:   "Mozilla:5.0 Platform:Tablet OS:Android Browser:Firefox-26.0 Engine:Gecko-26.0 Bot:false Mobile:true DeviceType:Tablet Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Android", "Android", ""},
	},
	{
		title:      "FirefoxOS",
		ua:         "Mozilla/5.0 (Mobile; rv:26.0) Gecko/26.0 Firefox/26.0",
		expected:   "Mozilla:5.0 Platform:Mobile OS:FirefoxOS Browser:Firefox-26.0 Engine:Gecko-26.0 Bot:false Mobile:true DeviceType:Mobile Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"FirefoxOS", "FirefoxOS", ""},
	},
	{
		title:    "FirefoxOSTablet",
		ua:       "Mozilla/5.0 (Tablet; rv:26.0) Gecko/26.0 Firefox/26.0",
		expected: "Mozilla:5.0 Platform:Tablet OS:FirefoxOS Browser:Firefox-26.0 Engine:Gecko-26.0 Bot:false Mobile:true DeviceType:Tablet Brand:Unknown Model:Unknown",
	},
	{
		title:      "FirefoxWinXP",
		ua:         "Mozilla/5.0 (Windows NT 5.2; rv:31.0) Gecko/20100101 Firefox/31.0",
		expected:   "Mozilla:5.0 Platform:Windows OS:Windows XP x64 Edition Browser:Firefox-31.0 Engine:Gecko-20100101 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Windows XP x64 Edition", "Windows", "XP"},
	},
	{
		title:    "FirefoxMRA",
		ua:       "Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:24.0) Gecko/20130405 MRA 5.5 (build 02842) Firefox/24.0 (.NET CLR 3.5.30729)",
		expected: "Mozilla:5.0 Platform:Windows OS:Windows XP Localization:en-US Browser:Firefox-24.0 Engine:Gecko-20130405 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},

	// Opera
	{
		title:      "OperaMac",
		ua:         "Opera/9.27 (Macintosh; Intel Mac OS X; U; en)",
		expected:   "Mozilla:Unknown Platform:Macintosh OS:Intel Mac OS X Localization:en Browser:Opera-9.27 Engine:Presto Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Intel Mac OS X", "Mac OS X", ""},
	},
	{
		title:    "OperaWin",
		ua:       "Opera/9.27 (Windows NT 5.1; U; en)",
		expected: "Mozilla:Unknown Platform:Windows OS:Windows XP Localization:en Browser:Opera-9.27 Engine:Presto Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "OperaWinNoLocale",
		ua:       "Opera/9.80 (Windows NT 5.1) Presto/2.12.388 Version/12.10",
		expected: "Mozilla:Unknown Platform:Windows OS:Windows XP Browser:Opera-9.80 Engine:Presto-2.12.388 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:      "OperaWin2Comment",
		ua:         "Opera/9.80 (Windows NT 6.0; WOW64) Presto/2.12.388 Version/12.15",
		expected:   "Mozilla:Unknown Platform:Windows OS:Windows Vista Browser:Opera-9.80 Engine:Presto-2.12.388 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Windows Vista", "Windows", "Vista"},
	},
	{
		title:    "OperaMinimal",
		ua:       "Opera/9.80",
		expected: "Mozilla:Unknown OS:Unknown Browser:Opera-9.80 Engine:Presto Bot:false Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "OperaFull",
		ua:       "Opera/9.80 (Windows NT 6.0; U; en) Presto/2.2.15 Version/10.10",
		expected: "Mozilla:Unknown Platform:Windows OS:Windows Vista Localization:en Browser:Opera-9.80 Engine:Presto-2.2.15 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "OperaLinux",
		ua:       "Opera/9.80 (X11; Linux x86_64) Presto/2.12.388 Version/12.10",
		expected: "Mozilla:Unknown Platform:X11 OS:Linux x86_64 Browser:Opera-9.80 Engine:Presto-2.12.388 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:      "OperaLinux - Ubuntu V41",
		ua:         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.99 Safari/537.36 OPR/41.0.2353.69",
		expected:   "Mozilla:5.0 Platform:X11 OS:Linux x86_64 Browser:Opera-41.0.2353.69 Engine:AppleWebKit-537.36 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Linux x86_64", "Linux", ""},
	},
	{
		title:      "OperaAndroid",
		ua:         "Opera/9.80 (Android 4.2.1; Linux; Opera Mobi/ADR-1212030829) Presto/2.11.355 Version/12.10",
		expected:   "Mozilla:Unknown Platform:Android 4.2.1 OS:Linux Browser:Opera-9.80 Engine:Presto-2.11.355 Bot:false Mobile:true DeviceType:Mobile Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Linux", "Linux", ""},
	},
	{
		title:    "OperaNested",
		ua:       "Opera/9.80 (Windows NT 5.1; MRA 6.0 (build 5831)) Presto/2.12.388 Version/12.10",
		expected: "Mozilla:Unknown Platform:Windows OS:Windows XP Browser:Opera-9.80 Engine:Presto-2.12.388 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "OperaMRA",
		ua:       "Opera/9.80 (Windows NT 6.1; U; MRA 5.8 (build 4139); en) Presto/2.9.168 Version/11.50",
		expected: "Mozilla:Unknown Platform:Windows OS:Windows 7 Localization:en Browser:Opera-9.80 Engine:Presto-2.9.168 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},

	// Other
	{
		title:    "Empty",
		ua:       "",
		expected: "Mozilla:Unknown OS:Unknown Browser:Unknown Engine:Unknown Bot:false Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "Nil",
		ua:       "nil",
		expected: "Mozilla:Unknown OS:Unknown Browser:nil Engine:Unknown Bot:false Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "Compatible",
		ua:       "Mozilla/4.0 (compatible)",
		expected: "Mozilla:Unknown OS:Unknown Browser:Mozilla-4.0 Engine:Unknown Bot:false Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "Mozilla",
		ua:       "Mozilla/5.0",
		expected: "Mozilla:Unknown OS:Unknown Browser:Mozilla-5.0 Engine:Unknown Bot:false Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "Amaya",
		ua:       "amaya/9.51 libwww/5.4.0",
		expected: "Mozilla:Unknown OS:Unknown Browser:amaya-9.51 Engine:libwww-5.4.0 Bot:false Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "Rails",
		ua:       "Rails Testing",
		expected: "Mozilla:Unknown OS:Unknown Browser:Rails Engine:Testing Bot:false Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "Python",
		ua:       "Python-urllib/2.7",
		expected: "Mozilla:Unknown OS:Unknown Browser:Python-urllib-2.7 Engine:Unknown Bot:false Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "Curl",
		ua:       "curl/7.28.1",
		expected: "Mozilla:Unknown OS:Unknown Browser:curl-7.28.1 Engine:Unknown Bot:false Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},

	// WebKit
	{
		title:      "ChromeLinux",
		ua:         "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.97 Safari/537.11",
		expected:   "Mozilla:5.0 Platform:X11 OS:Linux x86_64 Browser:Chrome-23.0.1271.97 Engine:AppleWebKit-537.11 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Linux x86_64", "Linux", ""},
	},
	{
		title:    "ChromeLinux - Ubuntu V55",
		ua:       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.75 Safari/537.36",
		expected: "Mozilla:5.0 Platform:X11 OS:Linux x86_64 Browser:Chrome-55.0.2883.75 Engine:AppleWebKit-537.36 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "ChromeWin7",
		ua:       "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.19 (KHTML, like Gecko) Chrome/18.0.1025.168 Safari/535.19",
		expected: "Mozilla:5.0 Platform:Windows OS:Windows 7 Browser:Chrome-18.0.1025.168 Engine:AppleWebKit-535.19 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "ChromeMinimal",
		ua:       "Mozilla/5.0 AppleWebKit/534.10 Chrome/8.0.552.215 Safari/534.10",
		expected: "Mozilla:5.0 OS:Unknown Browser:Chrome-8.0.552.215 Engine:AppleWebKit-534.10 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:      "ChromeMac",
		ua:         "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_5; en-US) AppleWebKit/534.10 (KHTML, like Gecko) Chrome/8.0.552.231 Safari/534.10",
		expected:   "Mozilla:5.0 Platform:Macintosh OS:Intel Mac OS X 10_6_5 Localization:en-US Browser:Chrome-8.0.552.231 Engine:AppleWebKit-534.10 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"Intel Mac OS X 10_6_5", "Mac OS X", "10.6.5"},
	},
	{
		title:    "SafariMac",
		ua:       "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_3; en-us) AppleWebKit/533.16 (KHTML, like Gecko) Version/5.0 Safari/533.16",
		expected: "Mozilla:5.0 Platform:Macintosh OS:Intel Mac OS X 10_6_3 Localization:en-us Browser:Safari-5.0 Engine:AppleWebKit-533.16 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "SafariWin",
		ua:       "Mozilla/5.0 (Windows; U; Windows NT 5.1; en) AppleWebKit/526.9 (KHTML, like Gecko) Version/4.0dp1 Safari/526.8",
		expected: "Mozilla:5.0 Platform:Windows OS:Windows XP Localization:en Browser:Safari-4.0dp1 Engine:AppleWebKit-526.9 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:      "iPhone7",
		ua:         "Mozilla/5.0 (iPhone; CPU iPhone OS 7_0_3 like Mac OS X) AppleWebKit/537.51.1 (KHTML, like Gecko) Version/7.0 Mobile/11B511 Safari/9537.53",
		expected:   "Mozilla:5.0 Platform:iPhone OS:IOS 7.0.3 Browser:Safari-7.0 Engine:AppleWebKit-537.51.1 Bot:false Mobile:true DeviceType:Mobile Brand:Apple Model:iPhone",
		expectedOS: &OSInfo{"IOS 7.0.3", "IOS", "7.0.3"},
	},
	{
		title:    "iPhone",
		ua:       "Mozilla/5.0 (iPhone; U; CPU like Mac OS X; en) AppleWebKit/420.1 (KHTML, like Gecko) Version/3.0 Mobile/4A102 Safari/419",
		expected: "Mozilla:5.0 Platform:iPhone OS:Unknown Localization:en Browser:Safari-3.0 Engine:AppleWebKit-420.1 Bot:false Mobile:true DeviceType:Mobile Brand:Apple Model:iPhone",
	},
	{
		title:    "iPod",
		ua:       "Mozilla/5.0 (iPod; U; CPU like Mac OS X; en) AppleWebKit/420.1 (KHTML, like Gecko) Version/3.0 Mobile/4A102 Safari/419",
		expected: "Mozilla:5.0 Platform:iPod OS:Unknown Localization:en Browser:Safari-3.0 Engine:AppleWebKit-420.1 Bot:false Mobile:true DeviceType:Mobile Brand:Apple Model:iPod",
	},
	{
		title:    "iPad",
		ua:       "Mozilla/5.0 (iPad; U; CPU OS 3_2 like Mac OS X; en-us) AppleWebKit/531.21.10 (KHTML, like Gecko) Version/4.0.4 Mobile/7B367 Safari/531.21.10",
		expected: "Mozilla:5.0 Platform:iPad OS:OS 3.2 Localization:en-us Browser:Safari-4.0.4 Engine:AppleWebKit-531.21.10 Bot:false Mobile:true DeviceType:Tablet Brand:Apple Model:iPad",
	},
	{
		title:    "webOS",
		ua:       "Mozilla/5.0 (webOS/1.4.0; U; en-US) AppleWebKit/532.2 (KHTML, like Gecko) Version/1.0 Safari/532.2 Pre/1.1",
		expected: "Mozilla:5.0 Platform:webOS OS:Palm Localization:en-US Browser:webOS-1.0 Engine:AppleWebKit-532.2 Bot:false Mobile:true DeviceType:Mobile Brand:Unknown Model:Unknown",
	},
	{
		title:    "Android",
		ua:       "Mozilla/5.0 (Linux; U; Android 1.5; de-; HTC Magic Build/PLAT-RC33) AppleWebKit/528.5+ (KHTML, like Gecko) Version/3.1.2 Mobile Safari/525.20.1",
		expected: "Mozilla:5.0 Platform:Linux OS:Android 1.5 Localization:de- Browser:Android-3.1.2 Engine:AppleWebKit-528.5+ Bot:false Mobile:true DeviceType:Mobile Brand:HTC Model:Magic",
	},
	{
		title:      "BlackBerry",
		ua:         "Mozilla/5.0 (BlackBerry; U; BlackBerry 9800; en) AppleWebKit/534.1+ (KHTML, Like Gecko) Version/6.0.0.141 Mobile Safari/534.1+",
		expected:   "Mozilla:5.0 Platform:BlackBerry OS:BlackBerry 9800 Localization:en Browser:BlackBerry-6.0.0.141 Engine:AppleWebKit-534.1+ Bot:false Mobile:true DeviceType:Mobile Brand:BlackBerry Model:Unknown",
		expectedOS: &OSInfo{"BlackBerry 9800", "BlackBerry", "9800"},
	},
	{
		title:    "BB10",
		ua:       "Mozilla/5.0 (BB10; Touch) AppleWebKit/537.3+ (KHTML, like Gecko) Version/10.0.9.388 Mobile Safari/537.3+",
		expected: "Mozilla:5.0 Platform:BlackBerry OS:BlackBerry Browser:BlackBerry-10.0.9.388 Engine:AppleWebKit-537.3+ Bot:false Mobile:true DeviceType:Mobile Brand:BlackBerry Model:Unknown",
	},
	{
		title:      "Ericsson",
		ua:         "Mozilla/5.0 (SymbianOS/9.4; U; Series60/5.0 Profile/MIDP-2.1 Configuration/CLDC-1.1) AppleWebKit/525 (KHTML, like Gecko) Version/3.0 Safari/525",
		expected:   "Mozilla:5.0 Platform:Symbian OS:SymbianOS/9.4 Browser:Symbian-3.0 Engine:AppleWebKit-525 Bot:false Mobile:true DeviceType:Mobile Brand:Unknown Model:Unknown",
		expectedOS: &OSInfo{"SymbianOS/9.4", "SymbianOS", "9.4"},
	},
	{
		title:    "ChromeAndroid",
		ua:       "Mozilla/5.0 (Linux; Android 4.2.1; Galaxy Nexus Build/JOP40D) AppleWebKit/535.19 (KHTML, like Gecko) Chrome/18.0.1025.166 Mobile Safari/535.19",
		expected: "Mozilla:5.0 Platform:Linux OS:Android 4.2.1 Browser:Chrome-18.0.1025.166 Engine:AppleWebKit-535.19 Bot:false Mobile:true DeviceType:Mobile Brand:Galaxy Model:Nexus",
	},
	{
		title:    "WebkitNoPlatform",
		ua:       "Mozilla/5.0 (en-us) AppleWebKit/525.13 (KHTML, like Gecko; Google Web Preview) Version/3.1 Safari/525.13",
		expected: "Mozilla:5.0 Platform:en-us OS:Unknown Localization:en-us Browser:Safari-3.1 Engine:AppleWebKit-525.13 Bot:false Mobile:false DeviceType:Unknown Brand:Unknown Model:Unknown",
	},
	{
		title:    "OperaWebkitMobile",
		ua:       "Mozilla/5.0 (Linux; Android 4.2.2; Galaxy Nexus Build/JDQ39) AppleWebKit/537.31 (KHTML, like Gecko) Chrome/26.0.1410.58 Mobile Safari/537.31 OPR/14.0.1074.57453",
		expected: "Mozilla:5.0 Platform:Linux OS:Android 4.2.2 Browser:Opera-14.0.1074.57453 Engine:AppleWebKit-537.31 Bot:false Mobile:true DeviceType:Mobile Brand:Galaxy Model:Nexus",
	},
	{
		title:    "OperaWebkitDesktop",
		ua:       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.31 (KHTML, like Gecko) Chrome/26.0.1410.58 Safari/537.31 OPR/14.0.1074.57453",
		expected: "Mozilla:5.0 Platform:X11 OS:Linux x86_64 Browser:Opera-14.0.1074.57453 Engine:AppleWebKit-537.31 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "ChromeNothingAfterU",
		ua:       "Mozilla/5.0 (Linux; U) AppleWebKit/537.4 (KHTML, like Gecko) Chrome/22.0.1229.79 Safari/537.4",
		expected: "Mozilla:5.0 Platform:Linux OS:Linux Browser:Chrome-22.0.1229.79 Engine:AppleWebKit-537.4 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "SafariOnSymbian",
		ua:       "Mozilla/5.0 (SymbianOS/9.1; U; [en-us]) AppleWebKit/413 (KHTML, like Gecko) Safari/413",
		expected: "Mozilla:5.0 Platform:Symbian OS:SymbianOS/9.1 Browser:Symbian-413 Engine:AppleWebKit-413 Bot:false Mobile:true DeviceType:Mobile Brand:Unknown Model:Unknown",
	},
	{
		title:    "Chromium - Ubuntu V49",
		ua:       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/49.0.2623.108 Chrome/49.0.2623.108 Safari/537.36",
		expected: "Mozilla:5.0 Platform:X11 OS:Linux x86_64 Browser:Chromium-49.0.2623.108 Engine:AppleWebKit-537.36 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "Chromium - Ubuntu V55",
		ua:       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/53.0.2785.143 Chrome/53.0.2785.143 Safari/537.36",
		expected: "Mozilla:5.0 Platform:X11 OS:Linux x86_64 Browser:Chromium-53.0.2785.143 Engine:AppleWebKit-537.36 Bot:false Mobile:false DeviceType:Desktop Brand:Unknown Model:Unknown",
	},

	// Dalvik
	{
		title:    "Dalvik - Dell:001DL",
		ua:       "Dalvik/1.2.0 (Linux; U; Android 2.2.2; 001DL Build/FRG83G)",
		expected: "Mozilla:5.0 Platform:Linux OS:Android 2.2.2 Browser:Unknown Engine:Unknown Bot:false Mobile:true DeviceType:Mobile Brand:Softbank Model:001DL",
	},
	{
		title:    "Dalvik - HTC:001HT",
		ua:       "Dalvik/1.4.0 (Linux; U; Android 2.3.3; 001HT Build/GRI40)",
		expected: "Mozilla:5.0 Platform:Linux OS:Android 2.3.3 Browser:Unknown Engine:Unknown Bot:false Mobile:true DeviceType:Mobile Brand:HTC Model:001HT",
	},
	{
		title:    "Dalvik - ZTE:009Z",
		ua:       "Dalvik/1.4.0 (Linux; U; Android 2.3.4; 009Z Build/GINGERBREAD)",
		expected: "Mozilla:5.0 Platform:Linux OS:Android 2.3.4 Browser:Unknown Engine:Unknown Bot:false Mobile:true DeviceType:Mobile Brand:ZTE Model:009Z",
	},
	{
		title:    "Dalvik - A850",
		ua:       "Dalvik/1.6.0 (Linux; U; Android 4.2.2; A850 Build/JDQ39) Configuration/CLDC-1.1; Opera Mini/att/4.2",
		expected: "Mozilla:5.0 Platform:Linux OS:Android 4.2.2 Browser:Unknown Engine:Unknown Bot:false Mobile:true DeviceType:Mobile Brand:Lenovo Model:A850",
	},
	{
		title:      "Dalvik - Asus:T00Q",
		ua:         "Dalvik/1.6.0 (Linux; U; Android 4.4.2; ASUS_T00Q Build/KVT49L)/CLDC-1.1",
		expected:   "Mozilla:5.0 Platform:Linux OS:Android 4.4.2 Browser:Unknown Engine:Unknown Bot:false Mobile:true DeviceType:Mobile Brand:Unknown Model:ASUS_T00Q",
		expectedOS: &OSInfo{"Android 4.4.2", "Android", "4.4.2"},
	},
	{
		title:    "Dalvik - W2430",
		ua:       "Dalvik/1.6.0 (Linux; U; Android 4.0.4; W2430 Build/IMM76D)014; Profile/MIDP-2.1 Configuration/CLDC-1",
		expected: "Mozilla:5.0 Platform:Linux OS:Android 4.0.4 Browser:Unknown Engine:Unknown Bot:false Mobile:true DeviceType:Mobile Brand:Unknown Model:W2430",
	},

	// QQ Browser
	{
		title:    "QQ - iPhone6",
		ua:       "Mozilla/5.0 (iPhone 6; CPU iPhone OS 10_2_1 like Mac OS X) AppleWebKit/602.4.6 (KHTML, like Gecko) Version/10.0 MQQBrowser/7.3 Mobile/14D27 Safari/8536.25 MttCustomUA/2 QBWebViewType/1",
		expected: "Mozilla:5.0 Platform:iPhone 6 OS:IOS 10.2.1 Browser:MQQBrowser-7.3 Engine:AppleWebKit-602.4.6 Bot:false Mobile:true DeviceType:Desktop Brand:Unknown Model:Unknown",
	},
	{
		title:    "QQ - iPad",
		ua:       "Mozilla/5.0 (iPad; U; CPU OS 4_2_1 like Mac OS X; zh-cn) AppleWebKit/533.17.9 (KHTML, like Gecko) /MQQBrowser/2.6 Mobile/8C148",
		expected: "Mozilla:5.0 Platform:iPad OS:OS 4.2.1 Localization:zh-cn Browser:Safari-MQQBrowser/2.6 Engine:AppleWebKit-533.17.9 Bot:false Mobile:true DeviceType:Tablet Brand:Apple Model:iPad",
	},
	{
		title:      "QQ - Android",
		ua:         "Mozilla/5.0 (Linux; U; Android 2.3.6; zh-cn; MB526 Build/4.5.1-134_DFP-1321) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile MQQBrowser/4.0 Safari/533.1",
		expected:   "Mozilla:5.0 Platform:Linux OS:Android 2.3.6 Localization:zh-cn Browser:MQQBrowser-4.0 Engine:AppleWebKit-533.1 Bot:false Mobile:true DeviceType:Mobile Brand:Motorola Model:MB526",
		expectedOS: &OSInfo{"Android 2.3.6", "Android", "2.3.6"},
	},

	// TEST
	{
		title:      "TEST",
		ua:         "Mozilla/5.0 (Linux; U; Android 4.4.2; en-us; HUAWEI Y541-U02 Build/HUAWEIY541-U02)  AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30",
		expected:   "Mozilla:5.0 Platform:Linux OS:Android 4.4.2 Localization:en-us Browser:Android-534.30 Engine:AppleWebKit-534.30 Bot:false Mobile:true DeviceType:Mobile Brand:HUAWEI Model:Y541-U02",
		expectedOS: &OSInfo{"Android 4.4.2", "Android", "4.4.2"},
	},

	// Glaxy S5
	{
		title:      "Glaxy S5",
		ua:         "Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Mobile Safari/537.36",
		expected:   "Mozilla:5.0 Platform:Linux OS:Android 5.0 Browser:Chrome-57.0.2987.133 Engine:AppleWebKit-537.36 Bot:false Mobile:true DeviceType:Mobile Brand:Samsung Model:SM-G900P",
		expectedOS: &OSInfo{"Android 5.0", "Android", "5.0"},
	},
}

// Internal: beautify the UserAgent reference into a string so it can be
// tested later on.
//
// ua - a UserAgent reference.
//
// Returns a string that contains the beautified representation.
func beautify(ua *UserAgent) (s string) {
	if len(ua.Mozilla()) > 0 {
		s += "Mozilla:" + ua.Mozilla() + " "
	}
	if len(ua.Platform()) > 0 && ua.Platform() != "Unknown" {
		s += "Platform:" + ua.Platform() + " "
	}
	if len(ua.OS()) > 0 {
		s += "OS:" + ua.OS() + " "
	}
	if len(ua.Localization()) > 0 {
		s += "Localization:" + ua.Localization() + " "
	}
	str1, str2 := ua.Browser()
	if len(str1) > 0 {
		s += "Browser:" + str1
		if len(str2) > 0 {
			s += "-" + str2 + " "
		} else {
			s += " "
		}
	}
	str1, str2 = ua.Engine()
	if len(str1) > 0 {
		s += "Engine:" + str1
		if len(str2) > 0 {
			s += "-" + str2 + " "
		} else {
			s += " "
		}
	}
	s += "Bot:" + fmt.Sprintf("%v", ua.Bot()) + " "
	s += "Mobile:" + fmt.Sprintf("%v", ua.Mobile()) + " "
	s += "DeviceType:" + ua.DeviceType() + " "
	s += "Brand:" + ua.MobileBrand() + " "
	s += "Model:" + ua.MobileModel()
	return s
}

// The test suite.
func TestUserAgent(t *testing.T) {
	for _, tt := range uastrings {
		ua := New(tt.ua)
		got := beautify(ua)
		if tt.expected != got {
			t.Errorf("\nTest:\t%v\ngot:\t%q\nexpected:\t%q\n", tt.title, got, tt.expected)
			//fmt.Printf("got:%v\n", []byte(got))
			//fmt.Printf("expected:%v\n", []byte(tt.expected))
		}

		if tt.expectedOS != nil {
			gotOSInfo := ua.OSInfo()
			if !reflect.DeepEqual(tt.expectedOS, &gotOSInfo) {
				t.Errorf("\nTest:\t%v\ngot:\t%#v\nexpected:\t%#v\n", tt.title, gotOSInfo, tt.expectedOS)
			}
		}
	}
}

// Benchmark: it parses each User-Agent string on the uastrings slice b.N times.
func BenchmarkUserAgent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for _, tt := range uastrings {
			ua := new(UserAgent)
			b.StartTimer()
			ua.Parse(tt.ua)
		}
	}
}
