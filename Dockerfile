FROM golang:1.21 AS glowdashbuildstage
RUN mkdir /glowdash
COPY glowdash/*.go /glowdash/
RUN cd /glowdash
RUN cd /glowdash && GO111MODULE=auto go get gopkg.in/yaml.v3
RUN cd /glowdash && GO111MODULE=auto go get github.com/hyper-prog/smartjson
RUN cd /glowdash && GO111MODULE=auto go get github.com/hyper-prog/smartyaml
RUN cd /glowdash && GO111MODULE=auto CGO_ENABLED=0 GOOS=linux go build -a -o glowdash \
    glowdash.go action.go group.go html.go sensors.go schedules.go scheduleedit.go schedulepanel.go \
    sensorgraph.go sensorstats.go hwdevice.go intstack.go pagebase.go panelbase.go script.go shading.go \
    switch.go tools.go thermostat.go launch.go

FROM alpine AS glowdash
LABEL maintainer="hyper80@gmail.com" Description="GlowDash - Smart Home Web Dashboard"
COPY --from=glowdashbuildstage /glowdash/glowdash /usr/local/bin
COPY --from=glowdashbuildstage /usr/share/zoneinfo /usr/share/zoneinfo
RUN mkdir /glowdash
VOLUME ["/glowdash"]
WORKDIR /glowdash
CMD ["/usr/local/bin/glowdash","/glowdash/config.yml"]
