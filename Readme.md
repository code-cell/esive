# Esive

<img width="400" height="400" src="https://raw.githubusercontent.com/code-cell/esive/master/_img/cubicle.png">

*Esive* is a generic MMO game. Still not set in any theme. The focus of the project is around the technical challenges of MMO, details on the game itself will come as we go.

### Try it
To try it using the test server just download the binary and run the `client` on your terminal.

## Current features

- Entity-Component-System using Redis as a storage for entities and components.
- Players can join the server and move around. They have a visibility range of 15 units.
- The world coordinates are [int64, int64] (pretty big)
- Players can chat with nearby players.
- Client side commands. Type `/help` to see them.
- Uses [Jaeger](https://www.jaegertracing.io/)

## Running a server

Before running the server you will have to run a redis instance. The Jaeger instance is optional.

### Using the binary

Visit the [Releases](https://github.com/code-cell/esive/releases), download the latest, unpack it and run `./server -h` to find out your options.

### Using docker

Docker images are hosted in GitHub. Follow [this guide](https://docs.github.com/en/packages/guides/configuring-docker-for-use-with-github-packages)
to make your docker able to pull images from it.

Once it's configured, run the following image: `ghcr.io/code-cell/esive_server:<VERSION>` using a proper version, find the latest one [here](https://github.com/orgs/code-cell/packages/container/package/esive_server). It's not advised to use `latest`.


## Running the client

There're no automated releases for the client and it has to be built at the moment.

To build it, clone this repo and run
```
make build_client
```

It generates the binary `esive_client`, so just run it with `./esive_client --name <YOUR NAME>`.
