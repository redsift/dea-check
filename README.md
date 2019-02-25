# DEA Check [![Go Report Card](https://goreportcard.com/badge/github.com/redsift/dea-check)](https://goreportcard.com/report/github.com/redsift/dea-check) [![GoDoc](https://godoc.org/github.com/redsift/dea-check?status.svg)](https://godoc.org/github.com/redsift/dea-check)

Using DEA (Disposable Email Address) is one of anti-spam techniques used to prevent unsolicited bulk email.
On other side, as a service provider solicitous about  security and privacy of our customers, we would waste resources on using those emails in our operations.
That is why we need to include simple DEA check as one of sanitizing methods into registration process to our services.

## Sources

- A lists of DEA and wildcard domains from https://github.com/ivolo/disposable-email-domains
    - [dea.json](https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/index.json)
    - [wildcard.json](https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/wildcard.json)

## Benchmarks

```text
goos: darwin
goarch: amd64
pkg: github.com/redsift/deacheck
BenchmarkIndex_Update-4      	      50	  32096713 ns/op	11520228 B/op	  265259 allocs/op
BenchmarkIndex_HasDomain-4   	10000000	       192 ns/op	      18 B/op	       1 allocs/op
PASS
```
