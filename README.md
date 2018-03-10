PrxPass-client
===

A client for the [prxpass-server](//github.com/Defman21/prxpass-server) project.

## Usage

```
prxpass-client -server remote:8080 -proxy localhost:4000
```

This will run a prxpass-client instance that will connect to the prxpass-server located at
`remote:8080` and proxy-pass every request made to `http(s)://<YOUR_ID>.remote` to `localhost:4000`.

