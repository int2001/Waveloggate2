package adif

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ContactInfo mirrors the FLDigi <contactinfo> XML format.
type ContactInfo struct {
	XMLName    xml.Name `xml:"contactinfo"`
	Call       string   `xml:"call"`
	Mode       string   `xml:"mode"`
	Timestamp  string   `xml:"timestamp"`
	TxFreq     string   `xml:"txfreq"`
	RxFreq     string   `xml:"rxfreq"`
	Rcv        string   `xml:"rcv"`
	Snt        string   `xml:"snt"`
	Power      string   `xml:"power"`
	Operator   string   `xml:"operator"`
	Comment    string   `xml:"comment"`
	SntNr      string   `xml:"sntnr"`
	RcvNr      string   `xml:"rcvnr"`
	MyCall     string   `xml:"mycall"`
	GridSquare string   `xml:"gridsquare"`
}

var tsFormats = []string{
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02T15:04:05.000",
}

func parseTimestamp(ts string) time.Time {
	// Force UTC by appending Z if not present.
	s := strings.TrimSpace(ts)
	for _, f := range tsFormats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}

func hzToMHz(hzStr string) string {
	hzStr = strings.TrimSpace(hzStr)
	hz, err := strconv.ParseFloat(hzStr, 64)
	if err != nil {
		return hzStr
	}
	mhz := hz / 1_000_000.0
	return fmt.Sprintf("%.6f", mhz)
}

// ParseXML decodes a FLDigi <contactinfo> XML string into a normalised ADIF field map.
func ParseXML(data string) (map[string]string, error) {
	var ci ContactInfo
	if err := xml.Unmarshal([]byte(data), &ci); err != nil {
		return nil, err
	}

	t := parseTimestamp(ci.Timestamp)
	date := t.Format("20060102")
	timeOn := t.Format("150405")

	mode := ci.Mode
	if mode == "USB" || mode == "LSB" {
		mode = "SSB"
	}

	freq := hzToMHz(ci.TxFreq)
	freqRx := hzToMHz(ci.RxFreq)

	mhz, _ := strconv.ParseFloat(freq, 64)
	band := FreqToBand(mhz)

	fields := map[string]string{
		"CALL":             ci.Call,
		"MODE":             mode,
		"QSO_DATE":         date,
		"QSO_DATE_OFF":     date,
		"TIME_ON":          timeOn,
		"TIME_OFF":         timeOn,
		"RST_RCVD":         ci.Rcv,
		"RST_SENT":         ci.Snt,
		"FREQ":             freq,
		"FREQ_RX":          freqRx,
		"OPERATOR":         ci.Operator,
		"COMMENT":          ci.Comment,
		"TX_PWR":           ci.Power,
		"STX":              ci.SntNr,
		"SRX":              ci.RcvNr,
		"MYCALL":           ci.MyCall,
		"GRIDSQUARE":       ci.GridSquare,
		"STATION_CALLSIGN": ci.MyCall,
		"BAND":             band,
	}

	// Remove empty fields.
	for k, v := range fields {
		if v == "" {
			delete(fields, k)
		}
	}

	return fields, nil
}
