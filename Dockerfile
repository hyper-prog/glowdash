FROM golang:1.23 AS glowdashbuildstage

WORKDIR /glowdash
COPY glowdash/go.mod glowdash/go.sum ./
RUN  GO111MODULE=auto go mod download

COPY glowdash/*.go /glowdash/

RUN GO111MODULE=auto CGO_ENABLED=0 GOOS=linux go build -a -o glowdash \
    glowdash.go action.go group.go html.go sensors.go schedules.go scheduleedit.go schedulepanel.go \
    sensorgraph.go sensorstats.go hwdevice.go intstack.go pagebase.go panelbase.go script.go shading.go \
    switch.go toggleswitch.go tools.go thermostat.go launch.go program.go

FROM alpine AS glowdash
LABEL maintainer="hyper80@gmail.com" \
      description="GlowDash - Smart Home Web Dashboard"
RUN mkdir /glowdash \
    && mkdir /glowdash/static \
    && mkdir /glowdash/user   \
    && mkdir /glowdash/config
COPY static/* /glowdash/static/
COPY --from=glowdashbuildstage /glowdash/glowdash /usr/local/bin
COPY --from=glowdashbuildstage /usr/share/zoneinfo /usr/share/zoneinfo
VOLUME ["/glowdash/user"]
VOLUME ["/glowdash/config"]
WORKDIR /glowdash
CMD ["/usr/local/bin/glowdash","/glowdash/config/running.yml"]
