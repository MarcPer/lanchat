# Lanchat

Chat on the terminal over the local network.

> Tested only on Linux.

## Usage

Download the binary and run it:

```sh
./lanchat -u my_user
```

Run `./lanchat -h` to see allowed flags.

The program automatically detects if there's already a running _Lanchat_ server in the local network and connects to it. If there isn't, the command starts one on port _6776_.

Start typing to chat, or run one of the available commands (enter `:h` to see a list).

### Build from source

Download repository and run `make`.

## Plan

- [x] Handle broadcasts properly (host should forward messages to all clients)
- [x] Fix handling of :id command: change own label
- [x] Transmit administrative messages to all peers (e.g. "user 'bla' connected")
- [x] Add notifications with a cooldown
- [ ] Tests
- [ ] Configuration file
- [ ] Become host if previous host disconnects; ping peers periodically
- [ ] Add :help (and :h) command

## Testing

A crude integration test can be run with `make check`. It creates a chat with few users and runs some commands for inspection.

## Architecture

The app is separated into two components:
- `Client`: Handles networking, sending and receiving messages, scanning for peers. It also parses both outbound and inbound messages to process commands (messages starting with `:`)
- `UI`: Responsible for handling the user interface, both the chat window and notifications.

Client and UI communicate to each other through two channels. For example, if the client receives a regular message, it will forward it to the UI to be rendered.

