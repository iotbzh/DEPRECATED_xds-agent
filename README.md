# XDS - X(cross) Development System Agent

XDS-agent is an agent that should run on your local machine when you use XDS.

This agent takes care of starting [Syncthing](https://syncthing.net/) tool to
synchronize your projects files from your local machine to build server machine
or container.


> **SEE ALSO**: [xds-server](https://github.com/iotbzh/xds-server), a web server
used to remotely cross build applications.


## How to build

### Dependencies

- Install and setup [Go](https://golang.org/doc/install) version 1.8 or
higher to compile this tool.


### Building

Clone this repo into your `$GOPATH/src/github.com/iotbzh` and use delivered Makefile:
```bash
 mkdir -p $GOPATH/src/github.com/iotbzh
 cd $GOPATH/src/github.com/iotbzh
 git clone https://github.com/iotbzh/xds-agent.git
 cd xds-agent
 make all
```

And to install xds-agent in /usr/local/bin:
```bash
make install
```

> **NOTE**: To cross build for example for Windows, just execute:
```bash
export GOOS=windows
export GOARCH=amd64
make all
make package
```

## How to run

## Configuration

xds-agent configuration is driven by a JSON config file (`agent-config.json`).

Here is the logic to determine which `agent-config.json` file will be used:
1. from command line option: `--config myConfig.json`
2. `$HOME/.xds/agent-config.json` file
3. `<current dir>/agent-config.json` file
4. `<xds-agent executable dir>/agent-config.json` file

Supported fields in configuration file are:
```json
{
    "httpPort": "http port of agent REST interface",
    "logsDir": "directory to store logs (eg. syncthing output)",
    "syncthing": {
        "binDir": "syncthing binaries directory (use xds-agent executable dir when not set)",
        "home": "syncthing home directory (usually .../syncthing-config)",
        "gui-address": "syncthing gui url (default http://localhost:8384)",
        "gui-apikey": "syncthing api-key to use (default auto-generated)"
    }
}
```

>**NOTE:** environment variables are supported by using `${MY_VAR}` syntax.

## Start-up

```bash
./bin/xds-agent.sh

# OR if you have install agent

/usr/local/bin/xds-agent.sh
```

>**NOTE** you can define some environment variables to setup for example
config file `XDS_CONFIGFILE` or change logs level `LOG_LEVEL`.
