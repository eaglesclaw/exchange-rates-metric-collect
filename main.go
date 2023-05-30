package main
import (
	_ "fmt"
	"io/ioutil"
	"log"
	"net/http"
	"encoding/xml"
	"time"
	"strconv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)
type Welcome10 struct {
	TarihDate TarihDate `xml:"Tarih_Date"`
}
type TarihDate struct {
	Currency []Currency `xml:"Currency"`  
	Tarih    string     `xml:"_Tarih"`    
	Date     string     `xml:"_Date"`     
	BultenNo string     `xml:"_Bulten_No"`
}
type Currency struct {
	Unit            string `xml:"Unit"`           
	Isim            string `xml:"Isim"`           
	CurrencyName    string `xml:"CurrencyName"`   
	ForexBuying     string `xml:"ForexBuying"`    
	ForexSelling    string `xml:"ForexSelling"`   
	BanknoteBuying  string `xml:"BanknoteBuying"` 
	BanknoteSelling string `xml:"BanknoteSelling"`
	CrossRateUSD    string `xml:"CrossRateUSD"`   
	CrossRateOther  string `xml:"CrossRateOther"` 
	CrossOrder      string `xml:"_CrossOrder"`    
	Kod             string `xml:"_Kod"`           
	CurrencyCode    string `xml:"_CurrencyCode"`  
}

var (
	banknoteBuying = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "banknote_buying",
			Help: "Banknote buying metric",
		},
		[]string{"currency_name"},
	)
	banknoteSelling = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "banknote_selling",
			Help: "Banknote selling metric",
		},
		[]string{"currency_name"},
	)
)

func init() {
	prometheus.MustRegister(banknoteBuying)
	prometheus.MustRegister(banknoteSelling)
}

func updateMetrics(currencies []Currency) {
	for _, element := range currencies {
		if element.CurrencyName == "EURO" || element.CurrencyName == "US DOLLAR" || element.CurrencyName == "POUND STERLING" {
			buying, err := strconv.ParseFloat(element.BanknoteBuying, 64)
			if err != nil {
				log.Printf("Failed to convert BanknoteBuying to float64: %s", err)
				continue
			}
			selling, err := strconv.ParseFloat(element.BanknoteSelling, 64)
			if err != nil {
				log.Printf("Failed to convert BanknoteSelling to float64: %s", err)
				continue
			}

			banknoteBuying.WithLabelValues(element.CurrencyName).Set(buying)
			banknoteSelling.WithLabelValues(element.CurrencyName).Set(selling)
		}
	}
}

func main(){
	resp, err := http.Get("https://www.tcmb.gov.tr/kurlar/today.xml")
	if err != nil {log.Fatal(err)}
	defer resp.Body.Close()

	data,err := ioutil.ReadAll(resp.Body)

	//fmt.Println(string(data))
	var results TarihDate
	err = xml.Unmarshal(data, &results)
	if err != nil {log.Fatal(err)}
	myCurrencies := results.Currency
//	for _, element := range myCurrencies {
//		if (element.CurrencyName == "EURO") || (element.CurrencyName == "US DOLLAR") || (element.CurrencyName == "POUND STERLING") {
//			fmt.Println(element.CurrencyName)
//			fmt.Println(element.BanknoteBuying)
//			fmt.Println(element.BanknoteSelling)
//		}
//	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	for {
		// Fetch and parse the XML data

		// Update the Prometheus metrics
		updateMetrics(myCurrencies)

		// Sleep or perform other tasks
		time.Sleep(time.Minute*20)
	}

}


