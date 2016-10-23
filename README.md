# remote-trigger

Allow local commands to be remotely triggered via HTTP. Think of git push hooks.


## Install

Assuming you have a configured `GOPATH` and `GOBIN` is part of your system's `PATH`, you can install this tool like that:

    git clone https://github.com/jojomi/remote-trigger.git
    cd remote-trigger
    go get

    export BRANCH=$(git rev-parse --abbrev-ref HEAD)
    export DATE=$(date)
    export HASH=$(git describe --always)
    go install -ldflags="-X main.gitCommitHash=${HASH} -X main.gitBranch=${BRANCH} -X main.compileDate=${DATE}"
    which remote-trigger


## Configuration

Create a config file from sample:

    cp remote-trigger.conf-example remote-trigger.conf

Edit the config and and your endpoints. The url part must be unique but you can use any string you want.


## Activation

Run the server:

    remote-trigger -p 1112 -c remote-trigger.conf -l debug

The default port is 5138. All parameters are optional.


### systemd Service

If you want to call `remote-trigger` with custom parameters, add them to `remote-trigger.service`.

Copy `remote-trigger.service` to `/etc/systemd/system/` and start the service:

    systemctl start remote-trigger


If you want the service to be started on boot, enable it in systemd:

    systemctl enable blub

