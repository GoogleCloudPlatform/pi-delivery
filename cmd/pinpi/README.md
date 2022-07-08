# pi in pi

This program finds pi in pi.

With the default configuration, it looks for digits "314159265358..." that is 10 decimals or longer in pi (directly fetched using the index file).

It displays results to stdout and logs to stderr, so use redirects to save the results to a file.

## Run

```bash
go run cmd/pinpi/main.go | tee pinpi.csv
```
