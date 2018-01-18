package useragent

import (
	"Service/log"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"os"
)

type wurfl struct {
	XMLName xml.Name `xml:"wurfl"`
	Version string   `xml:"version>last_updated"`
	Devices []device `xml:"devices>device"`
}

type device struct {
	Id               string  `xml:"id,attr"`
	FallBack         string  `xml:"fall_back,attr"`
	UserAgent        string  `xml:"user_agent,attr"`
	ActualDeviceRoot bool    `xml:"actual_device_root,attr"`
	Group            []group `xml:"group"`
}

type group struct {
	Id           string       `xml:"id,attr"`
	Capabilities []capability `xml:"capability"`
}

type capability struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type mobile struct {
	modelName string
	brandName string
}

var brandMap = make(map[string]string)

func GetBrand(model string) string {
	return brandMap[model]
}

func init() {
	readWurflFile()
}

func readWurflFile() {
	file, err := os.Open("wurfl-evaluation.xml")
	if err != nil {
		log.Errorf("Open Wurfl File Error: ", err)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Read Wurfl File Error: ", err)
		return
	}

	wurfl := wurfl{}
	err = xml.Unmarshal(data, &wurfl)
	if err != nil {
		log.Errorf("XML Unmarshal Error: ", err)
		return
	}

	deviceMap := make(map[string]device)
	for _, device := range wurfl.Devices {
		deviceMap[device.Id] = device
	}
	log.Infof("Get Devices: %v\n", len(deviceMap))

	// 写入文件
	fileName := "Brand.json"
	fileJson, _ := os.Create(fileName)
	defer fileJson.Close()
	for _, device := range wurfl.Devices {
		groups := device.Group
		mobile := getMobileFromGroup(groups)
		if mobile.brandName == "" {
			parentDevice := deviceMap[device.FallBack]
			groups := parentDevice.Group
			brandName := getMobileFromGroup(groups).brandName
			mobile.brandName = brandName
		}
		if mobile.brandName == "" || mobile.modelName == "" {
			continue
		}
		brandMap[mobile.modelName] = mobile.brandName
	}
	json, _ := json.Marshal(brandMap)
	fileJson.Write(json)
	log.Infof("Get Mobiles: %v\n", len(brandMap))

	/*db, err := sql.Open("mysql", "root:111111@tcp(127.0.0.1:3306)/Test")

	if err != nil {
		fmt.Printf("DB Error: %v\n", err)
	}
	sql := "insert into Mobile(mobileName, brandName) values(?, ?)"
	stm, _ := db.Prepare(sql)
	for _, mobile := range mobiles {
		//fmt.Printf("----------\nModelName: %v, BrandName: %v\n----------", mobile.ModelName, mobile.BrandName)
		if mobile.ModelName == "" && mobile.BrandName == "" {
			continue
		}
		stm.Exec(mobile.ModelName, mobile.BrandName)
	}*/
}

func getMobileFromGroup(groups []group) mobile {
	mobile := mobile{}
	for _, group := range groups {
		if group.Id != "product_info" {
			continue
		}
		capabs := group.Capabilities
		for _, capab := range capabs {
			if capab.Name == "model_name" {
				mobile.modelName = capab.Value
			}
			if capab.Name == "brand_name" {
				mobile.brandName = capab.Value
			}
		}
	}
	return mobile
}
