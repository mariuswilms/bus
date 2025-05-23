# bus

[![Go Reference](https://pkg.go.dev/badge/github.com/mariuswilms/bus.svg)](https://pkg.go.dev/github.com/mariuswilms/bus)

bus provides a simple publish and subscribe mechanism for Go. It is designed to be
lightweight and easy to use, with a focus on simplicity. The package
has no dependencies outside of the standard library. 

## Usage

I usually use the package to connect external event based service APIs to my Go
services so that I can update external state from internal events, if I want to.

```go
type PresenceDetector struct {
    *bus.Broker
}

// ...

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    br := bus.NewBroker(ctx)

    ha := NewHomeAssistantClient("...")
    presence := &PresenceDetector{Broker: br}

    // Receive published messages on a channel.
    _, ch := presence.Subscribe("changed")
    for msg := range ch {
        ha.SendEvent("presence.changed", msg.Data)
    }

    // Alternatively use a callback style.
    presence.SubscribeFn(ctx, "changed", func(msg Message) {
        ha.SendEvent("presence.changed", msg.Data)
    })
}
```

It is also possible to join together multiple brokers. 

```go
func main() {
    main := bus.NewBroker(ctx)

    br1 := bus.NewBroker(ctx) 
    br2 := bus.NewBroker(ctx)

    br1.Publish("foo", "Hello from br1")
    br2.Publish("bar", "Hello from br2")

    main.Connect(ctx, br1, "br1")
    main.Connect(ctx, br2, "br2")

    main.SubscribeFn(ctx, "br1:.*", func(msg Message) {
        // When an event is published on br1 to the "foo" topic, it will be received here
        // as "br1:foo".
    })
}
```

To enable debug logging, set the `BUS_DEBUG` environment variable to `y`. This will
enable debug logging for the bus package:

```shell
BUS_DEBUG=y go run .
```
