# CamJam

[![Build Status](https://travis-ci.org/abrightwell/camjam.svg?branch=master)](https://travis-ci.org/abrightwell/camjam)

CamJam is a webcam multiplexer designed for use with the FIRST Robotics
Competition (FRC).

## Overview

CamJam is a streaming camera server at its core. It publishes an HTTP endpoint
for clients to consume. Currently, it only supports streaming in MJPEG format.

## Installation

CamJam can be run on any \*nix platform that is supported by the go build tool.
However, it was intended to be run on a Raspberry Pi 3 that utilizes the
[Pi64](https://github.com) image. It is recommended that this hardware and
image be the target platform for installation as it is the reference platform
that CamJam is developed and tested against. The following instructions are
based on this recommendation.

**CamJam**

Copy the `camjam` binary to `/usr/local/bin`:

```
$> cp camjam /usr/local/bin
```

**Configuration File**

CamJam will automatically look for its configuration file in the `/etc/camjam`
directory.

If this directory does not exist, then create it:

```
$> mkdir /etc/camjam
```

Create/copy a the configuration file to `/etc/camjam/config.yaml`:

```
$> cp config.yaml /etc/camjam
```

**Systemd**

In order to start CamJam when the system it needs to be configured to run as a
service. This is accomplished by defining a systemd unit file.

Copy the [camjam.service]() file to `/etc/systemd/system`:

```
$> cp camjam.service /etc/systemd/system
```

Then enable and start the service:

```
$> sudo systemctl enable camjam
$> sudo systemctl start camjam
```


### `udev` configuration

In order to guarantee that each camera is always available at the same device
path, it is recommended to use `udev` rules to assign custom symlink for each
device.

Example:

```
KERNEL=="video*", KERNELS=="2-1:1.0", SYSMLINK+="camjam_0"
```

In the above example, the rule uses the kernel device name and the kernel
parent device name to assign the specified device path/location. The result in
this case would be that the video device found at USB Bus 2 on Port 1 would be
assigned the device `/dev/camjam_0`. 

Note: An example rules file can be found in the `examples` directory.

## Configuration

CamJam is configured using a YAML based configuration file. There are two main
sections of this file, `server` and `cameras`. The `server` section configures
all options related to running the server. The `cameras` section is a list of
camera configurations. The options for each section are detailed below. 

`cameras` is a list cameras 

**Server Options**

| Option           | Description                                                  |
| :--------------- | :----------------------------------------------------------- |
| address          | Configures the ip address and port that the server listens   |

**Camera Options**

| Option   | Description                                                  |
| :------- | :----------------------------------------------------------- |
| name     | Name of the camera. This can be any value that is relevant   |
|          | to the user.                                                 |
| device   | Path to the device. e.g. /dev/video0                         |
| format   | Format of the image captured by the camera. e.g. MJPG        |
| width    | Width of the image captured by the camera.                   |
| height   | Height of the image captured by the camera.                  |

Example:

```
server:
  address: 127.0.0.1:8000
cameras:
  - name: Front Camera
    device: /dev/video0
    format: MJPG
    width: 320
    height: 240
  - name: Rear Camera
    device: /dev/video1
    format: MJPG
    width: 320
    height: 240
```

## Build

CamJam can be built to run locally for development and testing purposes or it
can be cross compiled to run on a Raspberry Pi 3. 

**Requirements**

* [Golang 1.8+](https://golang.org)
* GNU make
* [dep](https://github.com/golang/dep)

To compile for local machine:

```
$> make
```

To compile for Raspberry Pi 3:

```
$> GOARCH=arm64 make
```

The resulting binary can be found in the `build` directory.
