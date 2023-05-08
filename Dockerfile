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
