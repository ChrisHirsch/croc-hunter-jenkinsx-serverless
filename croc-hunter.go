// The infamous "croc-hunter" game as featured at many a demo
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var release = os.Getenv("WORKFLOW_RELEASE")
var commit = os.Getenv("GIT_SHA")
var powered = os.Getenv("POWERED_BY")
var region = ""

func main() {
	httpListenAddr := flag.String("port", "8080", "HTTP Listen address.")

	flag.Parse()

	log.Println("Starting server...")

	log.Println("release: " + release)
	log.Println("commit: " + commit)
	log.Println("powered: " + powered)

	if release == "" {
		release = "unknown"
	}
	if commit == "" {
		commit = "not present"
	}
	if powered == "" {
		powered = "deis"
	}
	// get region

	for i := 0; i < 30; i++ {
		req, err := http.NewRequest("GET", "http://metadata/computeMetadata/v1/instance/attributes/cluster-location", nil)
		if err == nil {
			req.Header.Set("Metadata-Flavor", "Google")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("could not get region: %s", err)
				continue
			}
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				log.Printf("could not get region: %s", http.StatusText(resp.StatusCode))
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Printf("could not read region response: %s", err)
			} else {
				region = string(body)
			}
		} else {
			log.Printf("could not build region request: %s", err)
		}
		if region == "" {
			log.Printf("failed to get region, retrying")
			time.Sleep(1 * time.Second)
		} else {
			log.Printf("region: %s", region)
			break
		}
	}

	// point / at the handler function
	http.HandleFunc("/", handler)

	// serve static content from /static
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	log.Println("Server started. Listening on port " + *httpListenAddr)
	log.Fatal(http.ListenAndServe(":"+*httpListenAddr, nil))
}

const (
	html = `
		<html>
			<head>
				<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
				<title>Croc Hunter</title>
				<link rel='stylesheet' href='/static/game.css'/>
				<link rel="icon" type="image/png" href="/static/favicon-16x16.png" sizes="16x16" />
				<link rel="icon" type="image/png" href="/static/favicon-32x32.png" sizes="32x32" />
			</head>
			<body>
				<canvas id="canvasBg" width="800" height="490" ></canvas>
				<canvas id="canvasEnemy" width="800" height="500" ></canvas>
				<canvas id="canvasJet" width="800" height="500" ></canvas>
				<canvas id="canvasHud" width="800" height="500" ></canvas>
				<script src='/static/game.js'></script>
				<div class="details">
				<strong>Hostname: </strong><span id="hostname">%s</span><br>
				<strong>Region: </strong><span id="region">%s</span><br>
				<strong>Release: </strong><span id="release">%s</span><br>
				<strong>Commit: </strong><span id="commit">%s</span><br>
				<strong>Powered By: </strong>%s<br>
				</div>
			</body>
		</html>
		`
)

func handler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/healthz" {
		w.WriteHeader(http.StatusOK)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("could not get hostname: %s", err)
	}

	fmt.Fprintf(w, html, hostname, region, release, commit, powered)
}
