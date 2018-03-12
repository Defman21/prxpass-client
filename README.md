PrxPass-client
===

A client for the [prxpass-server](//github.com/Defman21/prxpass-server) project.

## Usage

```
prxpass-client -server remote:8080 -proxy-to localhost:4000 -id mycustomid
```

This will run a prxpass-client instance that will connect to the prxpass-server located at
`remote:8080` and proxy-pass every request made to `http(s)://mycustomid.remote` to `localhost:4000`.

If the server does not accept custom IDs, you'll get a generated one.
