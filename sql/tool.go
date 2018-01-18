package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"Service/db"
)

var table = flag.String("table", "OSWithVersions", "OSWithVersions Table Name")
var filepath = flag.String("file", "./OSWithVersions.txt", "loading file's path")
var referPath = flag.String("refer", "./countryCodeAlpha2And3.txt", "refer file path")

func main() {
	flag.Parse()
	fmt.Println("Result:", loadDB3(*filepath, *referPath,
		fmt.Sprintf(truncate, *table),
		fmt.Sprintf(ss3, *table)))
}

const ss = `INSERT INTO AdClickServer.%s(category,name) VALUES(?,?) ON DUPLICATE KEY UPDATE name=name`
const truncate = `TRUNCATE TABLE AdClickServer.%s`

func loadDB(path string, tt, ss string) error {
	fmt.Println("loadDB start", path, ss)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	data := make(map[string][]string)
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	_, err = db.GetDB("DB").Exec(tt)
	if err != nil {
		return err
	}

	stmt, err := db.GetDB("DB").Prepare(ss)
	if err != nil {
		return err
	}
	defer stmt.Close()
	fmt.Println("start range", len(data))
	i := 1
	for k, v := range data {
		for _, name := range v {
			if name == "" {
				name = k
			}

			_, err = stmt.Exec(k, name)
			if err != nil {
				return err
			}
			//fmt.Printf("INSERT %s(%s)", k, name)
			//fmt.Printf("                                                                \r")
		}
		FmtPercentage(i * 100 / len(data))
		i++
	}
	FmtPercentage(101)

	return nil
}

func FmtPercentage(percent int) {
	if percent > 100 {
		fmt.Printf("Finished!")
		for i := 1; i <= 100; i++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\n")
		return
	}
	for i := 1; i <= percent; i++ {
		fmt.Printf("=")
	}
	fmt.Printf(">(%%%d)\r", percent)
}

const ss2 = `INSERT INTO AdClickServer.%s(code,name) VALUES(?,?) ON DUPLICATE KEY UPDATE code=code`

func loadDB2(path string, tt, ss string) error {
	fmt.Println("loadDB start", path, ss)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	type element struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}
	decoder := json.NewDecoder(f)
	data := make([]element, 0)
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	_, err = db.GetDB("DB").Exec(tt)
	if err != nil {
		return err
	}

	stmt, err := db.GetDB("DB").Prepare(ss)
	if err != nil {
		return err
	}
	defer stmt.Close()
	fmt.Println("start range", len(data))
	i := 1
	for _, v := range data {
		_, err = stmt.Exec(v.Code, v.Name)
		if err != nil {
			return err
		}

		FmtPercentage(i * 100 / len(data))
		i++
	}
	FmtPercentage(101)

	return nil
}

const ss3 = `INSERT INTO AdClickServer.%s(country,name) VALUES(?,?) ON DUPLICATE KEY UPDATE name=name`

func loadDB3(path string, refer string, tt, ss string) error {
	fmt.Println("loadDB start", path, refer, ss)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	data := make([][]string, 0)
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	rf, err := os.Open(refer)
	if err != nil {
		return err
	}
	defer rf.Close()
	rdecoder := json.NewDecoder(rf)
	rdata := make(map[string]string)
	err = rdecoder.Decode(&rdata)
	if err != nil {
		return err
	}

	_, err = db.GetDB("DB").Exec(tt)
	if err != nil {
		return err
	}

	stmt, err := db.GetDB("DB").Prepare(ss)
	if err != nil {
		return err
	}
	defer stmt.Close()
	fmt.Println("start range", len(data))
	i := 1
	for _, v := range data {
		_, err = stmt.Exec(rdata[v[0]], v[1])
		if err != nil {
			return err
		}

		FmtPercentage(i * 100 / len(data))
		i++
	}
	FmtPercentage(101)

	return nil
}
