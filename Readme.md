# Esive

<img width="300" height="300" src="https://raw.githubusercontent.com/code-cell/esive/master/_img/cubicle.svg">

*Esive* is a generic MMO game. Still not set in any theme. The focus of the project is around the technical challenges of MMO, details on the game itself will come as we go.

## Current features

- Entity-Component-System using Redis as a storage for entities and components.
- Players can join the server and move around. They have a visibility range of 15 units.
- The world coordinates are [int64, int64] (pretty big)
- Players can chat with nearby players.
- Uses [Jaeger](https://www.jaegertracing.io/)

## Running the server

At this point the easiest is to clone the repo and run `make run_deps` and `make run_server`.

## Running the client

To run a client, the easiest at this point is to run `go run ./cmd/client --name <NAME>`.

- Use the arrows to move
- press `t` to chat.
