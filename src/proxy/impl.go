package proxy

import (
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Helper function to write an HTTP error and increment the errored request counter
func httpError(w http.ResponseWriter, endpointMetrics *EndpointMetrics) {
	endpointMetrics.FailedRequestCount.Inc()
	w.WriteHeader(http.StatusBadGateway)
}

// Registers all the HTTP endpoints based on the given config
func RegisterEndpoints(config *Config) {
	// Use config to register reverse proxy routes
	for _, endpoint := range config.Endpoints {
		log.Println("Registering endpoint")
		// Capture config values for use in the handler closure
		from := endpoint.From
		pool := endpoint.Pool

		endpointMetricsPrefix := FromResourceToMetricsPrefix(from)

		// Index used to implement round-robin load balancing
		poolIndex := 0

		// Since all threads spawned to handle this endpoint will need to use the pool index, we'll
		// need them all to share a mutex to ensure no races when incrementing the index.
		var poolMutex sync.Mutex

		// Metric counters local to this current endpoint
		endpointMetrics := NewEndpointMetrics(config, endpointMetricsPrefix)

		// Handle this endpoint!
		http.HandleFunc(from, func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			endpointMetrics.RequestCount.Inc()

			poolMutex.Lock()
			// The current value is the pool index we should use, so copy it for use outside the
			// critical section
			myPoolIndex := poolIndex

			// Shift index to move to the next endpoint for the next thread to run this handler
			poolIndex = (poolIndex + 1) % len(pool)

			poolMutex.Unlock()

			// Determine which target in pool should be used
			targetUrl := pool[myPoolIndex]

			// Forward the request onward
			httpClient := &http.Client{}

			forwardRequest, err := http.NewRequest(r.Method, targetUrl, r.Body)
			if err != nil {
				httpError(w, endpointMetrics)
				return
			}

			res, err := httpClient.Do(forwardRequest)
			if err != nil {
				log.Println(err)
				httpError(w, endpointMetrics)
				return
			}

			// Forward the response to the requester
			for key, values := range res.Header {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				httpError(w, endpointMetrics)
				return
			}

			w.WriteHeader(res.StatusCode)
			w.Write(resBody)

			// Success!
			endpointMetrics.SuccessfulRequestCount.Inc()

			// Time the handler body
			duration := time.Since(startTime)
			durationSeconds := float64(duration) / float64(time.Second)
			endpointMetrics.LastResponseTime.Set(durationSeconds)
		})
	}
}
