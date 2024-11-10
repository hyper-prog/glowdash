![GlowDash logo](https://raw.githubusercontent.com/hyper-prog/glowdash/master/storage/static/glowdash_s.webp)

GlowDash - The Smart Home Web Dashboard
============================================

GlowDash is web based dasboard to control (mostly) Shelly switches, shanding relays, scripts
and custom thermostat daemons. **It does not need any cloud access, it uses the local rpc interface** of
the Shelly relays to query/set states. It was originally written to allow for more complex control
over window shading slats as the original shelly application can.
Therefore, the program can define unique actions that can accept complex scripting methods.
The GlowDash written in Go language and designed to run in Docker container.

![GlowDash screenshot](https://raw.githubusercontent.com/hyper-prog/glowdash/master/docs/images/screenshot.jpg)

Use the GlowDash If:
- If you want a dashboard similar to Shelly's design.
- If you don't want to use the cloud or even you disabled cloud access on your devices, but you want to keep the functionality.
- If you want to create unique actions that perform complex tasks on multiple devices at the same time.
- If you want any local wifi device to be able to control your Shelly devices without a Shelly app or Shelly account.
- If you would like to use DHT22 and RaspberriPi based thermostat, with more sensors. (SMTherm daemon required)
- If you would like to monitor and log temperatures and humidity with DHT22 and RaspberriPi. (SMTherm daemon required)
- If you want to create scheduled tasks which runs your custom actions or set the termostat.

The SMTherm daemon is available here: https://github.com/hyper-prog/smtherm

Architecture
-----------------
The GlowDash written in Go language and designed to run in Docker container.
It has a Yaml config file which define all the panels and pages of the dasboard.
The configuration file must be passed to the program as a parameter.
The GlowDash (by design) maintains a minimal program state, so it does not need external database.
The current state of the devices is queried in every full page refresh throug RPC.
The only permanent information is the schedules tasks, which holds in a small text file.

The GlowDash can works together with Hasses (SSE daemon) to immediately show the background changing informations.
The dashboard is also functional without it, however, in this case,
the latest information will only be displayed when the page is updated.

Compile / Install
-----------------
It is recommended to use docker compose.
If you do so, the config file is the "storage/config.yml" and your user images can be put into "storage/user"
directory. After that, just edit the config file according to your needs and run the container

    cat docs/config-samples/minimal.yml > storage/config.yml
    docker compose build
    docker compose up -d


If you still want to compile it yourself, add the dependencies and compile all the *.go files

    export GO111MODULE=auto
    go install github.com/hyper-prog/smartjson
    go install github.com/hyper-prog/smartyaml

    go build -o glowdash glowdash/*.go
    ./glowdash myconfig.yml

Other devices, Future
---------------------
Although the program mainly supports Shelly devices, it has the option to control other types of devices as well.
Unfortunately I have only some types of Shelly relays too, so my testing possibilities are limited.
I would be happy if the program would support other devices in the future,
but only those capable of communicating via a local network (without cloud).
**Some Shelly devices also can only be read through cloud, they are not expected to be supported in GlowDash.**

Author
------
- Written by Péter Deák (C) hyper80@gmail.com, License GPLv2
- The author wrote this project entirely as a hobby. Any help is welcome!

------

[![paypal](https://raw.githubusercontent.com/hyper-prog/glowdash/master/docs/images/tipjar.png)](https://www.paypal.com/donate/?business=EM2E9A6BZBK64&no_recurring=0&currency_code=USD) 
