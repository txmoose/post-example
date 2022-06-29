# Post Example Tool
## What This Is
This tool is intended to be run locally to demonstrate how a POST and a GET work practically. It will write a `prod.db` SQLite3 database to disk, which can be safely deleted after use, or even parsed if you're familiar enough with SQLite.

## How To Use
copy the code down to your local machine and build it with `go build .` and then run the binary. In a separate terminal, you can issue the following `curl` commands to `localhost:8080`:
- ```curl -iX POST localhost:8080/create -d '{"key": "sample", "value": "sample_value"}'```
- ```curl -iX GET localhost:8080/get/{KEY}```

Feel free to experiment with these. The JSON you send to `/create` must be of the given form, though. It is recommended that you keep the value of `key` to only alphanumeric characters without spaces. The `{KEY}` portion of `/get/{KEY}` is going to be any value you pass in for `key` to the `/create` endpoint.