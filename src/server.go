package main

import (
  "fmt"
  "crypto/tls"
  "github.com/tidwall/gjson"
  "io"
  "log"
  "net/http"
  "os"
  "slices"
  "strings"	
)

func doQuery(w http.ResponseWriter, r *http.Request) {
  // Fetch querier url from env
  thanosQuerierUrl := os.Getenv("THANOS_QUERIER_URL")

  // Get Authorization token to passthrough
  incomingToken := strings.Split(r.Header.Get("Authorization"), " ")
  authToken := incomingToken[1]
  // Remove Authorization header to prevent it showing in logs
  r.Header.Del("Authorization") 

  // Get Query data to passthrough
  query := r.URL.Query().Get("query")
  namespace := r.URL.Query().Get("namespace")

  // Debugging
  fmt.Println("# New request ############################")
  fmt.Println("GET params were:", r.URL.Query())
  fmt.Println("Headers were:", r.Header)

  // Disable TLS Certificate Verify
  if os.Getenv("THANOS_QUERIER_INSECURE") == "true" {
    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
  }

  // Start creating request to Querier
  req, err := http.NewRequest("GET", thanosQuerierUrl, nil)
  req.Header.Add("Authorization", "Bearer " + authToken)

  // Add QueryString data to outgoing request
  q := req.URL.Query()
  q.Add("namespace", namespace)
  q.Add("query", query)
  req.URL.RawQuery = q.Encode()

  // Send outgoing request
  client := &http.Client{}
  resp, err := client.Do(req)

  if err != nil {
    fmt.Println(err)
    return
  }
  defer resp.Body.Close()

  // Pull json from response body
  json, err := io.ReadAll(resp.Body)
  if err != nil {
    log.Fatalln(err)
  }

  // Parse response json
  result := gjson.Parse(string(json)).Get("data.result")

  // Define which result keys to skip in output
  skipTags := []string{"__name__", "job", "prometheus"}

  // Iterate through json metric objects
  result.ForEach(func(key, value gjson.Result) bool {
    jsonData := gjson.Parse(value.String())

    objectData := jsonData.Get("metric")
    metricTime := jsonData.Get("value.0")
    metricValue := jsonData.Get("value.1")

    // Grab metric name
    name := objectData.Get("__name__")
    var tags string
    objectData.ForEach(func(key, value gjson.Result) bool {
      if !slices.Contains(skipTags, key.String()) {
        tags = tags + key.String() + "=\"" + value.String() + "\","
      }
      return true
    })
    // Remove trailing comma from tags
    tags = strings.TrimRight(tags, ",")

    // output metrics in OpenTelemetry style
    fmt.Fprintf(w, "%s{%s} %s %s\n", name, tags, metricTime.String(), metricValue.String())

    return true 
  })
}

func main() {
  // Create new webserver
  mux := http.NewServeMux()

  // Define routes
  mux.HandleFunc("/query", doQuery)

  // Start listening
  err := http.ListenAndServe(":4000", mux)
  log.Fatal(err)
}