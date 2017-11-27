Envoy websocket filter example.

NOTE: Require latest version of envoy which applied the merge request [#2070](https://github.com/envoyproxy/envoy/pull/2070)

## Build ##

    go get github.com/tangxinfa/envoy-socket.io-example
    cd $GOPATH/src/github.com/tangxinfa/envoy-socket.io-example
    glide install
    go build

## Run ##

### Single service node ###

    ./envoy-socket.io-example -addr 127.0.0.1:8001 -logtostderr &
    envoy --base-id 1 --config-path ./envoy.json

Open <http://localhost:9001/index.html> in your favorite browser, 
if socket.io connected to server a welcome message will arrive,
you can input something then click send, the server will echo the
same contents back.

### Multiple service nodes ###

    ./envoy-socket.io-example -addr 127.0.0.1:8002  -logtostderr &
    ./envoy-socket.io-example -addr 127.0.0.1:8003  -logtostderr &
    ./envoy-socket.io-example -addr 127.0.0.1:8004  -logtostderr &
    envoy --base-id 2 --config-path ./envoy2.json

Open <http://localhost:9002/index2.html> in your favorite browser,
open it in multiple tabs. In any tab, if socket.io connected to server
a welcome message will arrive, you can input something then click send,
the server will echo the same contents back.

The key ideas for use envoy to proxy socket.io on multiple service nodes:
  1. socket.io client must set `transports` option to `['websocket', 'polling']`.
  2. socket.io client must put the user identifier on url.
  3. envoy upstream cluster use `ring_hash` load balance type.
  4. envoy websocket route add `hash_policy` with header_name `referrer`.
See envoy2.json and index2.html for details.

If you try to open <http://localhost:9002/index.html>, you will see many
websocket reconnect events, because socket.io's transports option default
is `['polling', 'websocket']`, it need two http round-trip to establish
a websocket connection, in current envoy's implementation, the first http
polling request can schedule to the same upstream host, but the second is
websocket upgrade request, which use TcpProxy to connect upstream, it only
support random hash policy, which fails the socket.io handshake process.
