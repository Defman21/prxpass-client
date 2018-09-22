PrxPass-client
===

A client for the [prxpass-server](//github.com/Defman21/prxpass-server) project.

## Usage

```
prxpass-client -h
```


## Example 

```
prxpass-client -server remote:8080 -id mycustomid localhost:4000
```

Assuming that the prxpass server runs the http server at `0.0.0.0:80` and
the client server at `0.0.0.0:8080`, this will run a prxpass-client instance that will connect to the prxpass-server located at
`remote:8080` and proxy-pass every request made to `http(s)://mycustomid.remote` to `localhost:4000`.

If the server does not accept custom IDs, you'll get a generated one.
