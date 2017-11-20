Envoy websocket filter example.

Require latest version of envoy which applied the merge request [#2070](https://github.com/envoyproxy/envoy/pull/2070 "websocket: http request header Connection may contains multiple values #2070")

## Build ##

    go get github.com/tangxinfa/envoy-socket.io-example
    cd $GOPATH/src/github.com/tangxinfa/envoy-socket.io-example
    glide install
    go build

## Run ##

    ./envoy-socket.io-example -logtostderr &
    envoy --config-path ./envoy.json


Open <http://localhost:9001> with browser, if socket.io connected to server a welcome message will arrive,
you can send something, the server will echo back.
