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

Press `t` to chat.
In the chat, type `/help` to see the list of commands.

Visit the [Releases](https://github.com/code-cell/esive/releases), download the latest, unpack it and run `./client -h` to find out your options. The default options should just work connecting to the demo server.
