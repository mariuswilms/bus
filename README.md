# bus

[![Go Reference](https://pkg.go.dev/badge/github.com/mariuswilms/bus.svg)](https://pkg.go.dev/github.com/mariuswilms/bus)

bus provides a simple publish and subscribe mechanism for Go. It is designed to be
lightweight and easy to use, with a focus on simplicity. The package
has no dependencies outside of the standard library. Yes

## Usage

I usually use the package to connect external event based service APIs to my Go
services so that I can update external state from internal events, if I want to.

```go

type PresenceDetector struct {
    *bus.Broker
}

func main() {
    ha := NewHomeAssistantClient("...")
    presence := &PresenceDetector{br}

    presence.Subscribe("changed", func(ev Event) {
        ha.SendEvent("presence.changed", ev.Data)
    })
}
```

It is also possible to join together multiple brokers. 

```go
func main() {
    main := bus.NewBroker()

    br1 := bus.NewBroker() 
    br2 := bus.NewBroker()

    br1.Publish("foo", "Hello from br1")
    br2.Publish("bar", "Hello from br2")

    main.Connect(br1, "br1")
    main.Connect(br2, "br2")

    main.Subscribe("br1.foo", func(ev Event) {
        // When an event is published on br1 to the "foo" topic, it will be received here
        // as "br1.foo".
    })
}
```
