# XDS - X(cross) Development System Agent

XDS-agent is a client that should run on your local / user development machine when you use XDS.

This agent takes care, among others, of starting [Syncthing](https://syncthing.net/)
tool to synchronize your project files from your local host to XDS build server
machine or container (where `xds-server` is running).

> **SEE ALSO**: [xds-server](https://github.com/iotbzh/xds-server), a web server
used to remotely cross build applications.

## How to install xds-agent

### Install package for debian distro type

```bash
export DISTRO="Debian_8.0"
wget -O - http://download.opensuse.org/repositories/isv:/LinuxAutomotive:/app-Development/${DISTRO}/Release.key | sudo apt-key add -
sudo bash -c "cat >/etc/apt/sources.list.d/AGL.list<<EOL
#AGL
deb http://download.opensuse.org/repositories/isv:/LinuxAutomotive:/app-Development/${DISTRO}/ ./
EOL
"
sudo apt-get update
sudo apt-get install agl-xds-agent
```

The value 'DISTRO' can be set to {Debian_8.0, Debian_9.0, xUbuntu_16.04, xUbuntu_16.10, xUbuntu_17.04}

Update the package
```bash
sudo apt-get update
sudo apt-get upgrade agl-xds-agent
```

The files are install here:
```bash
/opt/AGL/agl-xds-agent
```

### Install package for rpm distro type

#### openSUSE
```bash
export DISTRO="openSUSE_Leap_42.2"
sudo zypper ar http://download.opensuse.org/repositories/isv:/LinuxAutomotive:/app-Development/${DISTRO}/isv:LinuxAutomotive:app-Development.repo
sudo zypper ref
sudo zypper install agl-xds-agent
```

The value 'DISTRO' can be set to {openSUSE_Leap_42.2, openSUSE_Leap_42.3, openSUSE_Tumbleweed}

Update the package
```bash
sudo zypper ref
sudo zypper install --force agl-xds-agent
```

The files are install here:
```bash
/opt/AGL/agl-xds-agent
```

## How to install on other platform

You need to download `xds-agent` tarballs from xds dashboard by clicking
on download icon ![download icon](./resources/images/download_icon.jpg) of
configuration page.

> **NOTE** : you can also download released tarballs from github [releases page](https://github.com/iotbzh/xds-agent/releases).

Then unzip this tarball any where into your local disk.

## Configuration

xds-agent configuration is driven by a JSON config file (named `agent-config.json`).
The tarball mentioned in previous section includes this file with default settings.

Here is the logic to determine which `agent-config.json` file will be used:
1. from command line option: `--config myConfig.json`
2. `$HOME/.xds/agent-config.json` file
3. `<current dir>/agent-config.json` file
4. `<xds-agent executable dir>/agent-config.json` file

Supported fields in configuration file are (all fields are optional and listed
values are the default values):
```
{
    "httpPort": "8010",                             # http port of agent REST interface
    "logsDir": "/tmp/logs",                         # directory to store logs (eg. syncthing output)
    "syncthing": {
        "binDir": ".",                              # syncthing binaries directory (default: executable directory)
        "home": "${HOME}/.xds/syncthing-config",    # syncthing home directory (usually .../syncthing-config)
        "gui-address": "http://localhost:8384",     # syncthing gui url (default http://localhost:8384)
        "gui-apikey": "123456789",                  # syncthing api-key to use (default auto-generated)
    }
}
```

>**NOTE:** environment variables are supported by using `${MY_VAR}` syntax.

## Start-up

Simply to start `xds-agent` executable
```bash
./xds-agent &
```

>**NOTE** if need be, you can increase log level by setting option
`--log <level>`, supported *level* are: panic, fatal, error, warn, info, debug.

You can now use XDS dashboard and check that connection with `xds-agent` is up.
(see also [xds-server README](https://github.com/iotbzh/xds-server/blob/master/README.md#xds-dashboard))


## Build xds-agent from scratch

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

And to install xds-agent (by default in `/usr/local/bin`):
```bash
make install
```

>**NOTE:** Used `DESTDIR` to specify another install directory
>```bash
>make install DESTDIR=$HOME/opt/xds-agent
>```

#### Cross build
For example on a Linux machine to cross-build for Windows, just execute:
```bash
export GOOS=windows
export GOARCH=amd64
make all
make package
```
