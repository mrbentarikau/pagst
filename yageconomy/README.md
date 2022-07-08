Economy, gambling, waifu and so on features for YAGDPDB

To use:

1. Add 	"github.com/jonas747/yageconomy" to the bottom of the imports in yags `cmd/yagpdb/main.go`
2. Add "yageconomy.RegisterPlugin()" at the bottom of all the other "RegisterPlugin()" calls in yags main.go

If using docker, you need to make a couple changes to the docker file:

1. Add `RUN go get -d -v github.com/jonas747/yageconomy` under all the other `git clone ...` near the top 
2. Uncomment the `COPY` statement for dev usage to apply the local changes we've made to the main.go file
3. Copy the webserver files to the proper dir by putting `COPY --from=builder /go/src/github.com/jonas747/yageconomy/assets/*.html templates/plugins/` under `# Handle templates for plugins automatically`

You should also make sure you're not using a prebuilt image by using the docker-compose.dev file instead of the standard one.

After those changes it should look like this:

```

FROM golang:stretch as builder

WORKDIR $GOPATH/src

RUN git clone -b yagpdb https://github.com/jonas747/discordgo github.com/jonas747/discordgo \
  && git clone -b dgofork https://github.com/jonas747/dutil github.com/jonas747/dutil \
  && git clone -b dgofork https://github.com/jonas747/dshardmanager github.com/jonas747/dshardmanager \
  && git clone -b dgofork https://github.com/jonas747/dcmd github.com/jonas747/dcmd

RUN go get -d -v github.com/jonas747/yageconomy

RUN go get -d -v \
  github.com/jonas747/yagpdb/cmd/yagpdb

# Uncomment during development
COPY . github.com/jonas747/yagpdb

# Disable CGO_ENABLED to force a totally static compile.
RUN CGO_ENABLED=0 GOOS=linux go install -v \
  github.com/jonas747/yagpdb/cmd/yagpdb


FROM alpine:latest

WORKDIR /app
VOLUME /app/soundboard \
  /app/cert
EXPOSE 80 443

# We need the X.509 certificates for client TLS to work.
RUN apk --no-cache add ca-certificates

# Add ffmpeg for soundboard support
RUN apk --no-cache add ffmpeg

# Handle templates for plugins automatically
COPY --from=builder /go/src/github.com/jonas747/yagpdb/*/assets/*.html templates/plugins/
COPY --from=builder /go/src/github.com/jonas747/yageconomy/assets/*.html templates/plugins/

COPY --from=builder /go/src/github.com/jonas747/yagpdb/cmd/yagpdb/templates templates/
COPY --from=builder /go/src/github.com/jonas747/yagpdb/cmd/yagpdb/posts posts/
COPY --from=builder /go/src/github.com/jonas747/yagpdb/cmd/yagpdb/static static/

COPY --from=builder /go/bin/yagpdb .

# add extra flags here when running YAGPDB
# Set "-exthttps=true" if using a TLS-enabled proxy such as
# jrcs/letsencrypt-nginx-proxy-companion
# Set "-https=false" do disable https
ENV extra_flags ""

# `exec` allows us to receive shutdown signals.
CMD exec /app/yagpdb -all -pa $extra_flags

```