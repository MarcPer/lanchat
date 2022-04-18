# Lanchat

Chat on the terminal over the local network.

> Tested only on Linux.

## Usage

Download the binary and run it:

```sh
./lanchat -u my_user
```

It automatically detects if there's already a running _Lanchat_ server in the local network and connects to it. If there isn't, the command starts one on port _6776_.

Start typing to chat, or run one of the available commands (enter `:h` to see a list).

### Build from source

Download repository and run `make`.

## Plan

- [ ] Handle broadcasts properly (host should forward messages to all clients)
- [ ] Add notifications with a cooldown
- [ ] Tests
- [ ] Configuration file
- [ ] Become host if previous host disconnects
- [ ] Add :help (and :h) command
