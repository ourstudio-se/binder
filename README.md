# Binder for Go

Binder is a configuration reader that parses different types of configurations and adds the possibility to bind them to one or many typed instances.

It can read configuration values from files, environment variables, remote URLs, Kubernetes volumes, and is flexible enough to enable custom configuration parsers. Binder is also able to listen for file changes/volume changes, and re-bind configurations when a backing file or backing volume has been updated.

Example:
```go
package main

import (
    "fmt"
    "github.com/ourstudio-se/binder"
)

type MyFirst struct {
    KeyOne string `config:"external_key_one"`
    KeyTwo int `config:"my_int"`
}

type MySecond struct {
    AnotherKey bool `config:"boolean_value"`
}

func main() {
    bnd := binder.New(
        binder.WithFile("../values.conf"),
        binder.WithEnv("Prefix_"),
        binder.WithWatch("../values.conf"))
    defer bnd.Close()
    
    var fst MyFirst
    var snd MySecond
    bnd.Bind(&fst, &snd);

    fmt.Printf("KeyOne: %s\n", fst.KeyOne)
    fmt.Printf("KeyTwo: %d\n", fst.KeyTwo)
    fmt.Printf("AnotherKey: %t\n", snd.AnotherKey)
}
```

It's also possible to use binder without binding to any instances, however there's no way to re-bind with a watch when using this pattern:
```go
package main

import (
    "fmt"
    "github.com/ourstudio-se/binder"
)

func main() {
    bnd := binder.New(
        binder.WithFile("../values.conf"),
        binder.WithEnv("Prefix_"),
        binder.WithWatch("../values.conf"))
    defer bnd.Close()
    
    values := bnd.Values() // WithWatch has no effect on the configuration values here

    keyOne, ok := values.Get("external_key_one")
    if !ok {
        panic("no such key")
    }

    fmt.Println("KeyOne: %s\n", keyOne)
}
```

To listen for any errors, which might come from any parser, or when binding, or from the file watcher, there's a chan available:
```go
package main

import (
    "fmt"
    "log"
    "github.com/ourstudio-se/binder"
)

func main() {
    bnd := binder.New(
        binder.WithFile("../values.conf"),
        binder.WithFile("non-existent-file.conf")
        binder.WithWatch("../values.conf"))
    defer bnd.Close()
    
    go func() {
        for {
            select {
                case err := <- bnd.Errors()
                    if err != nil {
                        log.Errorf("error occurred: %v", err)
                    }
                default:
                    continue
            }
        }
    }()

    var cfg MyConfig
    bnd.Bind(&cfg)
}
```

If a rebind happens, the implemented bound instance will get a notification if it implements a `Notify()` method:
```go
package main

import (
    "fmt"
    "github.com/ourstudio-se/binder"
)

type MyConfig struct {
    Property string `config:"property"`
}

func (cfg *MyConfig) Notify() {
    fmt.Println("I get called when my `Property` changes!")
}

func main() {
    bnd := binder.New(
        binder.WithFile("../values.conf"),
        binder.WithWatch("../values.conf"))
    defer bnd.Close()

    var cfg MyConfig
    bnd.Bind(&cfg)

}
```

One can specify a `BindMode` when matching a configuration key to a struct tag. Default is case insensitivity, meaning a struct tag `config:"mykey"` will match a configuration key `MyKey`. Pass the value `ModeStrict` to disable this behavior. Example:

```go
package main

import "github.com/ourstudio-se/binder"

type MyConfig struct {
    Property string `config:"PROPERTY"`
}

func main() {
    bnd := binder.New(
        binder.WithFile("../values.conf"),
        binder.WithBindMode(binder.ModeStrict))
    defer bnd.Close()

    var cfg MyConfig
    bnd.Bind(&cfg)

}
```
