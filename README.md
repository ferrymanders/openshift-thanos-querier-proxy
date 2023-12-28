# Thanos Querier Proxy

This tool is made to act as a proxy between the openshift thanos querier and a prometheus/grafana-agent instance.
When you query the proxy using a prometheus/grafana-agent scrape_job the proxy will convert the query to be send to the thanos querier and present the result in a OpenTelemetry format that then can be handled by the scraper.

It is required to send a namespace with the query and a oauth token from a ServiceAccount token that has `view` rights on that namespace.
Example of a query:
```
curl -G \
     --data-urlencode 'namespace=<namespace>' \
     --data-urlencode 'query={namespace="<namespace>"}' \
     -H "Authorization: Bearer <token>" \
     http://proxy.url/query
```

## Docker image
```
docker run -d \
    -e THANOS_QUERIER_URL="https://<THANOS_QUERIER_HOST>/api/v1/query" \
    fmanders/thanos-querier-proxy:latest
```