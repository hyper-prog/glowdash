![GlowDash logo](https://raw.githubusercontent.com/hyper-prog/glowdash/master/static/glowdash_s.webp)

GlowDash - The Smart Home Web Dashboard
============================================

GlowDash is a [web-based](https://en.wikipedia.org/wiki/World_Wide_Web) dashboard to control (mostly) [Shelly switches](https://www.shelly.com/), shading relays, scripts,
and custom thermostat daemons. **It does not need any cloud access; it uses the local RPC interface** of
the Shelly [relays](https://en.wikipedia.org/wiki/Relay) to query/set states. It was originally written to allow for more complex control
over window shading slats than the original Shelly application can.
Therefore, the program can define unique actions that accept complex scripting methods.
GlowDash is written in [Go language](https://en.wikipedia.org/wiki/Go_(programming_language)) and designed to run in a [Docker](https://en.wikipedia.org/wiki/Docker_(software)) container.

[![GlowDash youtube video](https://raw.githubusercontent.com/hyper-prog/glowdash/master/docs/images/woyt.png)](https://www.youtube.com/watch?v=y1USYtkOYOk)

![GlowDash screenshot](https://raw.githubusercontent.com/hyper-prog/glowdash/master/docs/images/screenshot.jpg)

Use GlowDash if:
- You want a dashboard similar to Shelly's design.
- You don't want to use the cloud, or you have disabled cloud access on your devices but want to keep the functionality.
- You want to create unique actions that perform complex tasks on multiple devices at the same time.
- You want any local WiFi device to be able to control your Shelly devices without a Shelly app or Shelly account.
- You would like to use DHT22 and [Raspberry Pi](https://en.wikipedia.org/wiki/Raspberry_Pi)-based thermostats with more sensors. (SMTherm daemon required)
- You would like to monitor and log temperatures and humidity with DHT22 and Raspberry Pi. (SMTherm daemon required)
- You want to create scheduled tasks that run your custom actions or set the thermostat.

The SMTherm daemon is available here: https://github.com/hyper-prog/smtherm

Architecture
-----------------
GlowDash is written in Go language and designed to run in a Docker container.
It has a YAML config file which defines all the panels and pages of the dashboard.
The configuration file must be passed to the program as a parameter.
GlowDash (by design) maintains a minimal program state, so it does not need an external database.
The current state of the devices is queried on every full page refresh through RPC.
The only permanent information is the scheduled tasks, which are stored in a small text file.

GlowDash can work together with Hasses (SSE daemon) to immediately show background changes.
The dashboard is also functional without it; however, in this case,
the latest information will only be displayed when the page is updated.

[Configuration documentation](docs/config-yaml.md)

Scripting Capabilities
-----------------------

GlowDash supports advanced scripting capabilities, allowing users to create custom actions and automation routines for their smart home devices. Scripting is integrated into the configuration file and enables complex logic, device control, and multi-step operations.

For details on configuration and scripting, see:

[GlowDash scripting language documentation](docs/glowdash-script.md)

Docker images
-------------
Available amd64 and arm64 Linux containers on Docker Hub:

- https://hub.docker.com/r/hyperprog/glowdash

 Downloadable (pullable) image name:

    hyperprog/glowdash


Compile / Install
-----------------
It is recommended to use Docker Compose.
If you do so, the config file is "config/running.yml" and your user images can be put into the "userstuff"
directory. After that, just edit the config file according to your needs and run the container

    cat config/minimal.yml > config/running.yml
    docker compose up -d


If you still want to compile it yourself, add the dependencies and compile all the *.go files:

    export GO111MODULE=auto
    go mod download

    go build -o glowdash glowdash/*.go
    ./glowdash config/running.yml

Other devices, Future
---------------------
Although the program mainly supports Shelly devices, it has the option to control other types of devices as well.
Unfortunately, I only have some types of Shelly relays, so my testing possibilities are limited.
I would be happy if the program supported other devices in the future,
but only those capable of communicating via a local network (without cloud).
**Some Shelly devices can only be read through the cloud; they are not expected to be supported in GlowDash.**

Author
------
- Written by Péter Deák (C) hyper80@gmail.com, License GPLv2
- The author wrote this project entirely as a hobby. Any help is welcome!

------

[![paypal](https://raw.githubusercontent.com/hyper-prog/glowdash/master/docs/images/tipjar.png)](https://www.paypal.com/donate/?business=EM2E9A6BZBK64&no_recurring=0&currency_code=USD) 
