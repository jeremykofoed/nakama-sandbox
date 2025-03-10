# Nakama Sandbox

This repository serves as a project testbed and sandbox environment for experimenting with [Nakama](https://heroiclabs.com/), an open-source scalable game server. It includes setup instructions for running Nakama locally, assumptions about registries for game components using `RWMutex` for live-ops overwrites, and guidance on testing with the provided `client.go`.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Setup Instructions](#setup-instructions)
- [Assumptions](#assumptions)
- [Tasks](#tasks)
- [Testing with client.go](#testing-with-clientgo)
- [References](#references)

## Prerequisites

Before setting up the Nakama server locally, ensure you have the following installed:

- [Docker](https://docs.docker.com/get-docker/): To containerize and run the Nakama server.
- [Go](https://golang.org/dl/): (version 1.23 or later) To run the client application.

## Setup Instructions

1. **Clone the Repository**

   ```bash
   git clone https://github.com/jeremykofoed/nakama-sandbox.git
   cd nakama-sandbox
   ```

2. **Build the Go Dependencies**

   Ensure Go is installed and run:

   ```bash
   go mod tidy
   go mod vendor
   ```

3. **Start Nakama Server with Docker Compose**

   Ensure Docker is installed and run:

   ```bash
   docker-compose up --build
   ```

   This command will set up and start the Nakama server along with its dependencies, such as the database. For more detailed instructions, refer to the [Nakama Docker Compose Setup Guide](https://heroiclabs.com/docs/nakama/getting-started/install/docker/).

4. **Verify Server is Running**

   The terminal window will begin pulling the required docker images.  Once those are completed it will then start building Nakama using the `Dockerfile` commands.  If it fails you might see messages like:

   ```bash
   => ERROR [nakama builder 4/4] RUN go build --trimpath --buildmode=plugin -o ./backend.so                                                            
   > [nakama builder 4/4] RUN go build --trimpath --buildmode=plugin -o ./backend.so:
   12.94 # github.com/nakama-sandbox
   12.94 ./main.go:66:49: undefined: LoadGameRPCs
   ```

   If successful then Namaka engine will begin outputting information about what it is loading.  The final entry before it is listening for remote calls to be made is: 

   ```bash
   nakamasandbox-nakama-1    | {"level":"info","ts":"2025-03-10T04:42:27.501Z","caller":"main.go:240","msg":"Startup done"}
   ```

   Another way is to access the Nakama developer console by navigating to `http://127.0.0.1:7351` in your web browser. The default credentials are:

   - **Username:** `admin`
   - **Password:** `password`

   To gracefully stop the Nakama server you can:

   ```bash
   Ctrl + C
   ```

## Assumptions

The project makes the following assumptions regarding game component registries:

1. **Registries for Attack Types, Enemy Types, and Status Effect Types:**

  These registries are implemented using Go's `sync.RWMutex` to allow safe concurrent access and modifications. This design supports live-ops overwrites, enabling dynamic updates to game components without requiring server restarts. For implementation details, refer to the respective Go files:

  - [attack.go](attack.go)
  - [enemy.go](enemy.go)
  - [status_effects.go](status_effects.go)

2. **Battle / Enemies**

   It was assumed that once a battle was finished another would begin and be created pairing an enemy.

3. **Enemy Attack Action**

   It was assumed, based on task and requirement interpretion, that an enemy did NOT have to perform any actions.  Time didn't allow for implementation of this at present writing. (Mar. 10 2025)

4. **Client Example**

   There are many frameworks that can be employed to faciliate the client logic.  I chose to avoid them and do it without a nakama framework to show understanding of what was taking place on a lower level.  Sometimes working with 3rd parties there are no frameworks and one must know how to interact with them.

## Tasks

The project had specific tasks, requirements, and bonuses (optional) that were to be met.

1. **RPC Call to Attack**

   This task was to accept a client request to make an attack action against a target requiring use of `player id`, `enemy id`, and an `attack action`.  This task was fulfilled on `client`/`client.go` making the request, `main.go` accepting,  and `rpc.go` processing the request.

2. **Enemy Data Stored in Nakama**

   This task was to ensure that the targeted enemy had their data stored in Nakama as a way to show knowledge of using the storage engine for reading and writing.  This task was fulfilled on `enemy.go` reading and writing enemy information to the storage enging and has an assumption. (See Assumptions) It was futher met with a specific enemy target associated with the player's data that would reference and instance of an enemy.

3. **Use RNG**

   This task was to be used primarily for determing if an attack successfully lands.  This task was fulfilled and futher extended to be used in other aspects of the game logic, like status effect application chance.

4. **Status Effects**

   This task was to implement status effects that would add some effect type to an entity (player or enemy) like Poison (reducing health pool over time) or Dazed (reducing the likely hood of landing an attack).  This task was fulfilled with application of status effects on the enemy as a player made an attack.

5. **Player Data Stored in Nakama**

   This task was to ensure that the player had their data stored in Nakama as a way to show knowledge of using the storage engine for reading and writing.  This task was fulfilled on `player.go` implementing the logic and being called on `rpc.go` for the RPC handler function.

6. **Health Pools Reach 0**

   This task was to inform the client making the request that further RPC calls to perform an attact action wasn't necessary and should be met with specific error handling at such time an entity's (player or enemy) health pool reach 0.  This task was fulfilled on `attack.go` whereby health was being check before and after an action was performed, including status effects like Poison or Bleed.

7. **RPC Call to Retrieve**

   This task was to provide another RPC call that would retrieve the player's current health, active status effects, and number of enemy types killed.  This task was fulfilled on `main.go` accepting and `rpc.go` processing the request.

8. **Use Nakama's Runtime**

   This task was to use Nakama runtime environment for data storage and retrieval.  This task was fulfilled on several accounts starting with the enemy registry initialized on `main.go` and processed on `enemy.go`.  Further fulfilled by storing the player data on `player.go` and retrieved by calling RPC `load_game` found on `main.go` processed on `rpc.go`.

9. **Logging and Error Handling**

   This task is to use the nakama runtime for logging and error handling.  This task was fulfilled on many locations throughout using calls such as `logger.Error("Effect interval 0, can't divide by 0: %+v", effect)` and `runtime.NewError("Enemy is deceased.", 5)` as examples.

10. **Use Nakama in Golang**

   This requirement was to ensure the use of Nakama's server framework.  This task was fulfilled using `Docker` to run Nakama Server as indicated by the `docker-compose.yml` file as well as the `Dockerfile`.

11. **Golang Best Practices**

   This requirement was to use proper structing of code, error handling, logging, and code comments.  This task was fulfilled to the best of my knowledge but is subjective.

12. **Client Accessible**

   This requirement was to ensure that a client could access the Nakama server and the custom RPCs.  This task was fulfilled on `client`/`client.go`.

13. **Include a README.md**

   This requirement was to make sure there were steps for being able to setup and test what was implemented.  This task was fulfilled with this `README.md`.

14. **Bonus: Period Status Effects**

   This bonus task was to implement a periodic status effect update to automatically apply effects at timed intervals.  This task was fulfilled by `attack.go` calling at specific moments logic on `status_effects.go` based on the status effect.  As of this writing (Mar. 10, 2025) the only period status effect implmeneted is `Bleed`.

15. **Bonus: Unit Tests**

   This bonus task was to implement unit tests to test critical function.  This task was NOT fulfilled.

16. **Bonus: Battle History**

   This bonus task was to implement a colllection of events into a history storing previous attacks and results.  This task was NOT fulfilled.

17. **Bonus: Store Player Data in Namaka Instead of In-Memory**

   This bonus task was to implement the use of storing the player's data in Nakama.  This task was fulfilled by storing the player data on `player.go` and retrieved by calling RPC `load_game` found on `main.go` processed on `rpc.go`.

18. **Bonus: Special Attack Types**

   This bonus task was to implment different attack types with various attributes.  This task was fulfilled on `attack.go` utilizing a registry that is initalized on `main.go` and has an assumption. (See Assumptions)

## Testing with client.go

The `client.go` file located in the `client` directory serves as a basic client to interact with the Nakama server. To test the server using this client open a separate terminal window or tab withing the window that the Nakama server is running on:

1. **Navigate to the Client Directory**

   ```bash
   cd client
   ```

2. **Build the Go Dependencies**

   Ensure Go is installed and run:

   ```bash
   go mod tidy
   go mod vendor
   ```

3. **Run the Client**

   Execute the client using the Go command:

   ```bash
   go run client.go
   ```

   The client will attempt to connect to the Nakama server at `127.0.0.1:7350` using the default server key `defaultkey`. Ensure these settings match your server configuration. You can adjust the server address and key in the `client.go` file if necessary.

4. **Observe Client Behavior**

   The client will perform predefined actions, such as authenticating, loading game state, and sending data to custom RPCs like attacking and getting information about health and stats, depending on the implementation within `client.go`. Monitor the console output for any errors or confirmations of successful interactions.

## References

- [Nakama Documentation](https://heroiclabs.com/docs/)
- [Setting Up Nakama with Docker Compose](https://heroiclabs.com/docs/nakama/getting-started/install/docker/)
- [Go Programming Language](https://golang.org/)

---

By following this guide, you should be able to set up a local Nakama server, understand the project's assumptions regarding registries and concurrency, and test interactions using the provided Go client. 
