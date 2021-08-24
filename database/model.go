package database

import (
	"cdr/object"
	"encoding/json"
	"fmt"
)

func AddData(data []object.CDRDataStruct) bool {
	db = GetConnection()
	for _, val := range data {
		jsondata, _ := json.Marshal(val)
		insForm, err := db.Prepare("INSERT INTO cdr (ANUM,BNUM,ServiceType,CallCategory,SubscriberType,StartDatetime,UsedAmount,RoundedUsedAmount,Charge,VoiceCharge,GprsCharge,SmsCharge) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)")
		if err != nil {
			fmt.Println(err)
			return false
		}
		_, err = insForm.Exec(jsondata, user)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}
	return true
}
