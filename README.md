# How to glue different applications inside a docker container or implement a reverse proxy in Go

Multiple web applications in the docker container may be implemented in different languages or frameworks. They all may need to be served from the container on the same external port so that it looks like a single application from outside the container.

You may have a third-party application to run in the container, which you cannot change, and you have to run it as is. At the same time, you would like to have your custom endpoints alongside that application.

For example, the main application runs from the root `/` and uses all required URLs below the root. You may want to have one URL, for example, `/health`, which you want to serve yourself.

The solution for this configuration is the reverse proxy, which routes requests inside the container to different ports, and even redirects the requests to external URLs.

## Enter the reverse proxy

In the article, we will "glue" three web applications implemented in Python, Node and Go. All three will be deployed in the same container, and each application will be invoked depending on the URL.

The default application is in Python. This application serves from the root (`/`) and all other URLs not handed by other applications. This application runs on port 9000 inside the container. This port is not exposed outside the container.

The application in Node runs on port 9100 and receives the requests with the `/node` prefix. This port also is not exposed outside the container.

The application in Go runs on port 8000 and receives requests with the "/go" prefix. This port is exposed as the container port.

Additionally, one more special prefix, `/google`, will redirect to Google Search.

Recap:

- `/node*` goes to the Node application
- `/go*` goes to the Go application
- `/google` goes to "httts://google.com".
- `/*` - everything else goes to the default Python application

The application in Go also implements the reverse proxy mechanism, the essence of the example.

Note:

The code in this article is for the demonstration purpose only. The configuration may be hard coded. In the actual production deployment, there should be a better abstraction on configuring Dockerfile and the applications.

The code of the applications is also the bare minimum. Each endpoint will print some text and the received URL.

## Python application

```python
from fastapi import FastAPI, Request
from fastapi.responses import PlainTextResponse

app = FastAPI()

@app.get("/{path:path}", response_class=PlainTextResponse)
async def root(request: Request, path: str):
    return f"I'm Python!\r\n[{path}]\r\n"
```

## Node application

```javascript
const { createServer } = require("http");

createServer((req, res) => {
    res.setHeader("content-type", "text/plain");
    res.end(`I'm Node!\r\n[${req.url}]\r\n`);
}).listen(process.env.PORT || 9000);
```

## Go application with a built-in proxy

This is the meat and potatoes of the article.

It uses the "ReverseProxy" object and the "NewSingleHostReverseProxy" utility from the standard Go library. The application has no third-party external dependencies.

The code below starts the HTTP listener on port 8000, which will be exposed outside the container, and redirects the traffic to other ports according to the logic above.

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "os"
    "strings"
)

func main() {
    defaultURL, err := url.Parse("http://localhost:9000")
    if err != nil {
        log.Fatal(fmt.Errorf("error parsing default URL: %v", err))
    }

    nodeURL, err := url.Parse("http://localhost:9100")
    if err != nil {
        log.Fatal(fmt.Errorf("error parsing node URL: %v", err))
    }

    googleURL, err := url.Parse("https://google.com")
    if err != nil {
        log.Fatal(fmt.Errorf("error parsing google's URL: %v", err))
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if q, found := strings.CutPrefix(r.URL.Path, "/google"); found {
            proxy := &httputil.ReverseProxy{
                Rewrite: func(r *httputil.ProxyRequest) {
                    r.SetURL(googleURL)
                    r.Out.URL.Path = "/" + q
                },
            }
            proxy.ServeHTTP(w, r)
            return
        }
        if strings.HasPrefix(r.URL.Path, "/go") {
            w.Write([]byte(fmt.Sprintf("I'm Go!\r\n[%v]\n", r.URL.Path)))
            return
        }
        if strings.HasPrefix(r.URL.Path, "/node") {
            httputil.NewSingleHostReverseProxy(nodeURL).ServeHTTP(w, r)
            return
        }
        httputil.NewSingleHostReverseProxy(defaultURL).ServeHTTP(w, r)
    })
    port := os.Getenv("PORT")
    if port == "" {
        port = "8000"
    }
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal(err)
    }
}
```

## Container startup script

This script starts all three applications on different ports and provides the required environment variables. This script is the container entry point.

```sh
#!/usr/bin/env bash

(source .venv/bin/activate && uvicorn main:app --port 9000) &
(PORT=9100 ./node ./main.js) &
(PORT=8000 ./proxy) &

wait
exit $?
```

## Dockerfile

Dockerfile is multistaged because it needs to collect artefacts from the different applications in one container.

```dockerfile
ARG GO_VERSION=1.20

FROM golang:${GO_VERSION}-alpine AS build-proxy
WORKDIR /app

COPY main.go .
RUN go build -o proxy main.go

# ---
FROM python:3-slim AS build-python

WORKDIR /app

COPY main.py .
RUN python -m venv .venv
RUN . .venv/bin/activate && pip install uvicorn fastapi

# ---

FROM node:20-bullseye-slim AS build-node

WORKDIR /app

COPY main.js .
RUN cp `which node` .

# ---
FROM python:3-slim
WORKDIR /app

COPY --from=build-python /app .
COPY --from=build-node /app .
COPY --from=build-proxy /app/proxy .
COPY ./run.sh .

CMD ["./run.sh"]
```

## Build the docker image

The image's name is "zoo".

```sh
docker build -t zoo .
```

## Running the container

This command runs the container in interactive mode and with `--rm` to be deleted when it stops.

```sh
docker run -p 8000:8000 --rm -it zoo
```

## Testing

The container is up the running. It listens on port 8000, so we can test it.

Hitting the Python application:

```sh
curl http://localhost:8000

I'm Python!
[]

curl http://localhost:8000/a/b/c

I'm Python!
[a/b/c]
```

Hitting the Node application:

```sh
curl http://localhost:8000/node

I'm Node!
[/node]

curl http://localhost:8000/node/a/b/c
I'm Node!
[/node/a/b/c]
```

Hitting the Go application:

```sh
curl http://localhost:8000/go

I'm Go!
[/go]

curl http://localhost:8000/go/a/b/c
I'm Go!
[/go/a/b/c]
```

Finally, you can put the <http://localhost:8000/google/?q=abc> URL into the browser, which will forward you to Google Search.

This is it! We have implemented the configurable reverse proxy in Go.

It glues together three applications written in different languages, and they run inside one container as a single deployable unit.

## Makefile

For convenience, a Makefile that can build the image and run the container:

```sh
make
```

Run tests above:

```sh
make test
```

## Performance

Extra data transfer may be necessary when the traffic goes through the proxy. We say "maybe" because, in reality, the proxy can propagate actual data copying between sockets to the kernel. The approach is called [Zero-Copy networking](https://lwn.net/Articles/726917/). It eliminates the overhead because all network listeners run on the same machine and are controlled by the same kernel.

The Go standard library on Linux does precisely that.

Extra copying is needed when the proxy redirects to the external URL, but this is not the intended use case for this application.

## Links

The sources are available at <https://github.com/begoon/go-reverse-proxy>.
