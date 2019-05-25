package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/benderpan/webanalyze"
)

var (
	update       bool
	outputMethod string
	workers      int
	apps         string
	host         string
	hosts        string
	crawlCount   int
)

func init() {
	flag.StringVar(&outputMethod, "output", "stdout", "output format (stdout|csv|json)")
	flag.BoolVar(&update, "update", false, "update apps file")
	flag.IntVar(&workers, "worker", 4, "number of worker")
	flag.StringVar(&apps, "apps", "apps.json", "app definition file.")
	flag.StringVar(&host, "host", "", "single host to test")
	flag.StringVar(&hosts, "hosts", "", "filename with hosts, one host per line.")
	flag.IntVar(&crawlCount, "crawl", 0, "links to follow from the root page (default 0)")

	if cpu := runtime.NumCPU(); cpu == 1 {
		runtime.GOMAXPROCS(2)
	} else {
		runtime.GOMAXPROCS(cpu)
	}
}

func main() {
	var file io.ReadCloser
	var err error

	flag.Parse()

	if !update && host == "" && hosts == "" {
		flag.Usage()
		return
	}

	if update {
		err = webanalyze.DownloadFile(webanalyze.WappalyzerURL, "apps.json")
		if err != nil {
			log.Fatalf("error: can not update apps file: %v", err)
		}

		log.Println("app definition file updated from ", webanalyze.WappalyzerURL)

		if host == "" && hosts == "" {
			return
		}

	}

	// check single host or hosts file
	if host != "" {
		file = ioutil.NopCloser(strings.NewReader(host))
	} else {
		file, err = os.Open(hosts)
		if err != nil {
			log.Fatalf("error: can not open host file %s: %s", hosts, err)
		}
	}
	defer file.Close()

	results, err := webanalyze.Init(workers, file, apps, crawlCount)

	if err != nil {
		log.Fatal("error initializing:", err)
	}

	log.Printf("Scanning with %v workers.", workers)

	f, err := os.OpenFile("data.json", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("data.json file create failed. err: " + err.Error())
	}
	defer f.Close()

	for result := range results {
		if result.Error != "" {
			log.Printf("[Debug] Error: Host=%s  Msg=%s", result.Host, result.Error)
		}

		jsonValue, err := json.Marshal(result)
		if err != nil {
			os.Stdout.Write([]byte("{}\n"))
			f.Write([]byte("{}\n"))
		} else {
			jsonValue = append(jsonValue, '\n')
			os.Stdout.Write(jsonValue)
			f.Write(jsonValue)
		}

	}

}
