# Nakama Sandbox

This repository serves as a sandbox environment for experimenting with [Nakama](https://heroiclabs.com/), an open-source scalable game server. It includes setup instructions for running Nakama locally, assumptions about registries for game components using `RWMutex` for live-ops overwrites, and guidance on testing with the provided `client.go`.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Setup Instructions](#setup-instructions)
- [Assumptions](#assumptions)
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

- **Registries for Attack Types, Enemy Types, and Status Effect Types:**

  These registries are implemented using Go's `sync.RWMutex` to allow safe concurrent access and modifications. This design supports live-ops overwrites, enabling dynamic updates to game components without requiring server restarts. For implementation details, refer to the respective Go files:

  - [attack.go](attack.go)
  - [enemy.go](enemy.go)
  - [status_effects.go](status_effects.go)

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
