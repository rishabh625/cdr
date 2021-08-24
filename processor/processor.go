package processor

import (
	"cdr/object"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var Config object.Configuration

//CDRs????
func FlatfileProcessor(extension, suffixtension, path string, delimeter rune, wg *sync.WaitGroup, errChan chan error) {
	log.Println("Flatfile Reader Started")
	filestr := fmt.Sprintf("%s.%s", suffixtension, extension)
	for {
		select {
		case er := <-errChan:
			wg.Done()
			errChan <- er
		default:
			fileArr, err := WalkMatch(path, filestr)
			if err != nil {
				fmt.Println(err)
			}
			for _, v := range fileArr {
				f, err := os.Open(v)
				if err != nil {
					log.Fatal(err)
				}
				reader := csv.NewReader(f)
				reader.Comma = delimeter
				lines, err := reader.ReadAll()
				if err != nil {
					logrus.Fatal(err)
				}
				structureMap := make(map[int]int)
				line1 := lines[0]
				for k, v := range line1 {
					switch v {
					case "ANUM":
						structureMap[0] = k
					case "BNUM":
						structureMap[1] = k
					case "ServiceType":
						structureMap[2] = k
					case "CallCategory":
						structureMap[3] = k
					case "SubscriberType":
						structureMap[4] = k
					case "StartDatetime":
						structureMap[5] = k
					case "UsedAmount":
						structureMap[6] = k
					}
				}
				d := make([]*object.CDRDataStruct, 0)
				for i := 1; i < len(lines); i++ {
					cdrdata := object.CDRDataStruct{

						ANUM:           lines[i][structureMap[0]],
						BNUM:           lines[i][structureMap[1]],
						ServiceType:    lines[i][structureMap[2]],
						CallCategory:   lines[i][structureMap[3]],
						SubscriberType: lines[i][structureMap[4]],
						StartDateTime:  lines[i][structureMap[5]],
						UsedAmount:     lines[i][structureMap[6]],
					}
					d = append(d, &cdrdata)
				}
				getOutputFile(Config.OutputFilePath, d)
				f.Close()
				err = os.Rename(v, "archive/"+v)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		//wg.Done()
	}
}

func JsonProcessor(extension, suffixtension, path string, wg *sync.WaitGroup, errChan chan error) {
	log.Println("Json Reader Started")
	filestr := fmt.Sprintf("%s.%s", suffixtension, extension)
	for {
		select {
		case er := <-errChan:
			wg.Done()
			errChan <- er
		default:
			fileArr, err := WalkMatch(path, filestr)
			if err != nil {
				fmt.Println(err)
			}
			for _, v := range fileArr {
				file, err := ioutil.ReadFile(v)
				if err != nil {
					log.Fatal(err)
				}
				data := []*object.CDRDataStruct{}

				err = json.Unmarshal([]byte(file), &data)
				if err != nil {
					log.Fatal(err)
				}
				getOutputFile(Config.OutputFilePath, data)
				//file.Close()
				err = os.Rename(v, "archive/"+v)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

}

func XmlProcessor(extension, suffixtension, path string, wg *sync.WaitGroup, errChan chan error) {
	log.Println("Xml Reader Started")
	filestr := fmt.Sprintf("%s.%s", suffixtension, extension)
	for {
		select {
		case er := <-errChan:
			wg.Done()
			errChan <- er
		default:
			fileArr, err := WalkMatch(path, filestr)
			if err != nil {
				fmt.Println(err)
			}
			for _, v := range fileArr {
				xmlFile, err := os.Open(v)
				if err != nil {
					log.Fatal(err)
				}
				byteValue, err := ioutil.ReadAll(xmlFile)
				if err != nil {
					log.Fatal(err)
				}
				var data object.CDRXML
				err = xml.Unmarshal(byteValue, &data)
				if err != nil {
					log.Fatal(err)
				}
				xmlFile.Close()
				getOutputFile(Config.OutputFilePath, data.CDR)
				//file.Close()
				err = os.Rename(v, "archive/"+v)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func getRoundedUsedAmount(serviceType, Value string) string {
	intval, err := strconv.ParseFloat(Value, 32)
	if err != nil {
		intval = 0
	}
	switch serviceType {
	case "1":
		mins := int(math.Ceil(intval / 60))
		minstr := strconv.Itoa(mins)
		return minstr
	case "2":
		return Value
	case "3":
		kb := int(math.Ceil(intval / 1024))
		mbstr := strconv.Itoa(kb)
		return mbstr
	}
	return ""
}

func getNormalizeBNUM(bnum string) string {
	bnum = strings.Replace(bnum, "+", "", -1)
	num, _ := strconv.Atoi(bnum)
	bnum = strconv.Itoa(num)
	return bnum
}

func getMaxVoiceConsumption(data []object.CDRDataStruct) string {
	max := math.MinInt64
	anum := data[0].ANUM
	for _, v := range data {
		compareVal, _ := strconv.Atoi(v.RoundedUsedAmount)
		if compareVal > max {
			max = compareVal
			anum = v.ANUM
		}
	}
	return anum
}

func getOutputFile(path string, data []*object.CDRDataStruct) {
	dbins := make([]object.CDRDataStruct, 0)
	for _, v := range data {
		v.RoundedUsedAmount = getRoundedUsedAmount(v.ServiceType, v.UsedAmount)
		v.BNUM = getNormalizeBNUM(v.BNUM)
		calculateCharge(data)
		switch v.ServiceType {
		case "1":
			v.ServiceType = "Voice"
		case "2":
			v.ServiceType = "SMS"
		case "3":
			v.ServiceType = "GPRS"
		}

		switch v.CallCategory {
		case "1":
			v.CallCategory = "Local"
		case "2":
			v.CallCategory = "Roaming"
		}

		switch v.SubscriberType {
		case "1":
			v.SubscriberType = "Postpaid"
		case "2":
			v.SubscriberType = "Prepaid"
		}
		layout := "20060102150405"
		t, err := time.Parse(layout, v.StartDateTime)
		dateval := t.Format("2006-01-02 15:04:05")
		if err != nil {
			fmt.Println(err)
			dateval = ""
		}
		dbval := object.CDRDataStruct{
			ANUM:              v.ANUM,
			BNUM:              v.BNUM,
			ServiceType:       v.ServiceType,
			CallCategory:      v.CallCategory,
			SubscriberType:    v.SubscriberType,
			StartDateTime:     dateval,
			UsedAmount:        v.UsedAmount,
			RoundedUsedAmount: v.RoundedUsedAmount,
			Charge:            v.Charge,
			VoiceCharge:       v.VoiceCharge,
			GprsCharge:        v.GprsCharge,
			SmsCharge:         v.SmsCharge,
		}
		dbins = append(dbins, dbval)
		v.UsedAmount = v.RoundedUsedAmount
		v.RoundedUsedAmount = ""
		v.GprsCharge = 0
		v.SmsCharge = 0
		v.VoiceCharge = 0
		v.StartDateTime = ""
		v.ANUM = ""
		v.XMLName.Space = ""
		v.XMLName.Local = ""
	}

	sort.Slice(data, func(p, q int) bool {
		return data[p].Charge < data[q].Charge
	})
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(string(file)); err != nil {
		panic(err)
	}
}

func calculateCharge(data []*object.CDRDataStruct) {
	for _, v := range data {
		for _, chargevals := range Config.ChargeTable {
			arr := strings.Split(chargevals, "|")
			if v.ServiceType == arr[0] && v.CallCategory == arr[1] && v.SubscriberType == arr[2] {
				usedval, _ := strconv.ParseFloat(v.RoundedUsedAmount, 64)
				chargeval, _ := strconv.ParseFloat(arr[3], 64)
				v.Charge += usedval * chargeval
				if v.ServiceType == "1" {
					v.VoiceCharge += usedval * chargeval
				}
				if v.ServiceType == "2" {
					v.SmsCharge += usedval * chargeval
				}
				if v.ServiceType == "3" {
					v.GprsCharge += usedval * chargeval
				}
			}
		}
	}
}
