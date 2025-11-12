package stacks

import (
    "fmt"
    "github.com/hyperledger/firefly-cli/pkg/types"
    "gopkg.in/yaml.v3"
)

func main() {
    cfg := &types.FireflyConfig{
        Config: &types.CoreConfig{},
    }

    cfg.Config.AutoReload = true

    out, err := yaml.Marshal(cfg)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(out))
}
