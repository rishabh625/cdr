package object

import "encoding/xml"

// Configuration ... of App
type Configuration struct {
	InputFilePath  string       `mapstructure:"inputfilepath"`
	OutputFilePath string       `mapstructure:"outputfilepath"`
	File           []FileConfig `mapstructure:"files_config"`
	ChargeTable    []string     `mapstructure:"charge_table"`
}

// Redis ... Configuration
type FileConfig struct {
	Queue []Queues `mapstructure:"queue"`
}

// Queues ... Queue name and string
type Queues struct {
	Name            string `mapstructure:"queue_name"`
	ProcessingCount int    `mapstructure:"queue_processing_size"`
	ProcessingType  string `mapstructure:"queue_processing_type"`
	FileExtension   string `mapstructure:"queue_processing_extension"`
	Delimeter       string `mapstructure:"delimeter"`
	SuffixExtension string `mapstructure:"suffix_extension"`
}

type CDRXML struct {
	XMLName xml.Name         `xml:"CDRs"`
	CDR     []*CDRDataStruct `xml:"CDR"`
}

type CDRDataStruct struct {
	XMLName           xml.Name `xml:"CDR" json:"-"`
	ANUM              string   `xml:"ANUM" json:"ANUM,omitempty"`
	BNUM              string   `xml:"BNUM" json:"BNUM"`
	ServiceType       string   `xml:"ServiceType" json:"ServiceType"`
	CallCategory      string   `xml:"CallCategory" json:"CallCategory"`
	SubscriberType    string   `xml:"SubscriberType" json:"SubscriberType"`
	StartDateTime     string   `xml:"StartDatetime" json:"StartDatetime,omitempty"`
	UsedAmount        string   `xml:"UsedAmount" json:"UsedAmount"`
	RoundedUsedAmount string   `json:"RoundedUsedAmount,omitempty"`
	Charge            float64  `json:"Charge,omitempty"`
	VoiceCharge       float64  `json:"VoiceCharge,omitempty"`
	GprsCharge        float64  `json:"GprsCharge,omitempty"`
	SmsCharge         float64  `json:"SmsCharge,omitempty"`
}
