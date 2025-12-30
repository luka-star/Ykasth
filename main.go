package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"sync"
	"time"
)

// ... (PriceRecord, EmissionRecord, MergedRecord, EnerginetResponse, Fetcher, DataStore remain the same) ...

type PriceRecord struct {
	TimeUTC          string  `json:"TimeUTC"`
	PriceArea        string  `json:"PriceArea"`
	DayAheadPriceEUR float64 `json:"DayAheadPriceEUR"`
}

type EmissionRecord struct {
	Minutes5UTC string  `json:"Minutes5UTC"`
	CO2Emission float64 `json:"CO2Emission"`
}

type MergedRecord struct {
	TimeUTC          string  `json:"TimeUTC"`
	PriceArea        string  `json:"PriceArea"`
	DayAheadPriceEUR float64 `json:"DayAheadPriceEUR"`
	CO2Emission      float64 `json:"CO2Emission"`
	CO2Ratio         float64 `json:"CO2Ratio,omitempty"`
}

type EnerginetResponse[T any] struct {
	Records []T `json:"records"`
}

type Fetcher[T any] struct {
	Client *http.Client
}

func NewFetcher[T any](timeout time.Duration) *Fetcher[T] {
	return &Fetcher[T]{Client: &http.Client{Timeout: timeout}}
}

func (f *Fetcher[T]) Request(apiURL string) (T, error) {
	var result T
	resp, err := f.Client.Get(apiURL)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("api status: %d", resp.StatusCode)
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

type DataStore struct {
	sync.RWMutex
	LatestData []MergedRecord
}

func (s *DataStore) Update(data []MergedRecord) {
	s.Lock()
	defer s.Unlock()
	s.LatestData = data
}

func AggregateAndMerge(priceData EnerginetResponse[PriceRecord], emissionData EnerginetResponse[EmissionRecord]) []MergedRecord {
	emissionMap := make(map[string]float64)
	for _, rec := range emissionData.Records {
		emissionMap[rec.Minutes5UTC] = rec.CO2Emission
	}

	var mergedResults []MergedRecord
	for _, pRec := range priceData.Records {
		if co2, exists := emissionMap[pRec.TimeUTC]; exists {
			ratio := pRec.DayAheadPriceEUR / (co2 + 0.001)
			mergedResults = append(mergedResults, MergedRecord{
				TimeUTC:          pRec.TimeUTC,
				PriceArea:        pRec.PriceArea,
				DayAheadPriceEUR: pRec.DayAheadPriceEUR,
				CO2Emission:      co2,
				CO2Ratio:         ratio,
			})
		}
	}
	for i, j := 0, len(mergedResults)-1; i < j; i, j = i+1, j-1 {
		mergedResults[i], mergedResults[j] = mergedResults[j], mergedResults[i]
	}
	return mergedResults
}

type Reporter struct {
	Area string // Store area to display in UI
}

func (r *Reporter) ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (r *Reporter) GetVerdict(m MergedRecord) string {
	if m.CO2Emission > 110 {
		return "DIRTY"
	}
	if m.DayAheadPriceEUR < 70 && m.CO2Emission < 60 {
		return "GREEN"
	}
	if m.CO2Ratio > 1.3 {
		return "GREEN"
	} else if m.CO2Ratio < 0.8 {
		return "DIRTY"
	}
	return "NEUTRAL"
}

func (r *Reporter) RefreshUI(data []MergedRecord) {
	r.ClearScreen()
	fmt.Println("==========================================================================================")
	fmt.Printf(" ENERGY DASHBOARD for %s - 6 hours ahead (Last Update: %s)\n", r.Area, time.Now().Format("15:04:05"))
	fmt.Println("==========================================================================================")
	fmt.Printf("%-18s | %-8s | %-5s | %-8s | %-15s\n", "Time", "Price", "CO2", "EUR/CO2", "VERDICT")
	fmt.Println("------------------------------------------------------------------------------------------")

	var sumP, sumC, sumR float64
	for _, m := range data {
		layout := "2006-01-02T15:04:05"
		t, err := time.Parse(layout, m.TimeUTC)
		displayTime := m.TimeUTC
		if err == nil {
			displayTime = t.Format("02 Jan 15:04")
		}

		verdict := r.GetVerdict(m)
		fmt.Printf("%-18s | %8.2f | %5.0f | %8.4f | %-15s\n",
			displayTime, m.DayAheadPriceEUR, m.CO2Emission, m.CO2Ratio, verdict)

		sumP += m.DayAheadPriceEUR
		sumC += m.CO2Emission
		sumR += m.CO2Ratio
	}

	if count := float64(len(data)); count > 0 {
		fmt.Println("------------------------------------------------------------------------------------------")
		fmt.Printf("%-18s | %8.2f | %5.0f | %8.4f | \n",
			"SUMMARY (Averages)", sumP/count, sumC/count, sumR/count)
		fmt.Println("==========================================================================================")
	}
}

func main() {
	area := "DK1"
	if len(os.Args) > 1 {
		area = os.Args[1]
	}

	pFetcher := NewFetcher[EnerginetResponse[PriceRecord]](10 * time.Second)
	eFetcher := NewFetcher[EnerginetResponse[EmissionRecord]](10 * time.Second)
	store := &DataStore{}
	reporter := &Reporter{Area: area}
	updateQueue := make(chan []MergedRecord, 1)

	go func() {
		for newData := range updateQueue {
			store.Update(newData)
			reporter.RefreshUI(newData)
		}
	}()

	var lastState []MergedRecord
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		baseURL := "https://api.energidataservice.dk/dataset/"
		loc, _ := time.LoadLocation("Europe/Copenhagen")
		now := time.Now().In(loc).Truncate(time.Hour).Add(time.Hour)
		startStr := now.Format("2006-01-02T15:04")
		endStr := now.Add(6 * time.Hour).Add(15 * time.Minute).Format("2006-01-02T15:04")
		areafilter := fmt.Sprintf(`{"PriceArea":["%s"]}`, area)
		filter := url.QueryEscape(areafilter)

		pURL := fmt.Sprintf(baseURL+"/DayAheadPrices?start=%s&end=%s&filter=%s", startStr, endStr, filter)
		eURL := fmt.Sprintf(baseURL+"CO2EmisProg?start=%s&end=%s&filter=%s", startStr, endStr, filter)

		pData, errP := pFetcher.Request(pURL)
		eData, errE := eFetcher.Request(eURL)

		if errP == nil && errE == nil {
			currentMerged := AggregateAndMerge(pData, eData)
			if !reflect.DeepEqual(currentMerged, lastState) {
				lastState = currentMerged
				updateQueue <- currentMerged
			}
		}
	}
}
