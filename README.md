# Fetch SRE Health Check Exercise 

## Overview

This Go program performs health checks on a list of HTTP endpoints specified in a YAML configuration file. It does the following:

1. Read an input argument to a file path with a list of HTTP endpoints in YAML format.
2. Test the health of the endpoints every 15 seconds.
3. Track cumulative availability percentage for
each domain and log to console after the completion of each 15-second test cycle.
4. Keep testing the endpoints every 15 seconds until the user manually exits the program.

## Setup
* Go 1.21 or later installed on your machine. You can download it from [golang.org](https://golang.org/dl/).
* Install dependencies: `go mod tidy`

## Usage
> Create a YAML configuration file. Please see `example.yaml` that was originally provided.

For local testing and development, run `go run main.go example.yaml`.

To produce an executable file to run independently, run `go build -o health-check` and `./health-check example.yaml`.

## Assumptions
This program is developed under these assumptions:

1. Only YAML file is accepted as input. The program rejects other file input.
2. YAML content is valid (i.e. valid endpoint with valid URL). As a result, no validation is done on parsed endpoint information. We assume method is allowed, url is valid and headers/body are well-formed.


## Other Considerations
To optimize performance, this program utilizes shared HTTP client with 1 second timeout to prevent hanging requests. In addition, the program runs health check request concurrently in goroutines with a concurrency limit of 10. Here are some considerations for future scalability:

1. Concurrency Limit & Timeouts: The program hardcodes a concurrent limit of 10 and an HTTP client timeout of 1 second. Since UP is categorized to be latency of 500ms or less, 1 second seems to be a good metric for unresponsive domain. For future development, we should reconsider timeout and transport settings, as well as concurrency limit with respect to system resources. 
2. No retries against transient failures: With frequent checks of 15 seconds, transient errors are partially mitigated. However, for future development, we should reconsider the likelihood of such false positives. In addition, if a domain is known to be unresponsive, we should consider backing off.
3. Graceful shutdown: When the program receives an interrupt (Ctrl+C), it exits immediately. For future development, we should consider more graceful handling such as waiting for all goroutines to complete before exiting.

