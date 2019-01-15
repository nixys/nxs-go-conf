# nxs-go-conf

This Go package provides a simple way to handle configuration parameters for Go programs.

## Features

- **Manage options in structure field tags**  
To describe configuration file structure you simply need to define the struct in the Go program code. In that struct you can use field tags to set different options and to determine config file decoding behavior. Currently, the next tags are available:
  - `conf`: defines custom name for an option
  - `conf_extraopts`: provides advanced settings for option. This tag may have the following values:
    - `required`: option with this tag is mandatory. If it is set, but corresponding option is not defined in the config file, it will cause an error.
    - `default`: determines default value for the option. Can be set only for _Int*_, _Uint*_, _Bool_ and _String_ (not within the arrays, maps or slices) types.

- **ENV variables as option values**  
  You may specify the option value as `ENV:VARIABLE_NAME`. It will use the value of the relative environment variable (i.e. _VARIABLE_NAME_) as value for that option.

- **YAML and JSON formats are available**  
  Currently, you can use config files in YAML or JSON formats. To switch the format you only need to specify the appropriate setting for config file load function.

- **Catch the unknown options**  
  You can catch options, that are contained in config file but has no matching in the result interface.

## Install

```
go get github.com/nixys/nxs-go-conf
```

## Example of usage

*You may find more examples in unit-tests in this repository*

**Parse simple config file:**

```go
package main

import (
        "fmt"

        "github.com/nixys/nxs-go-conf"
)

const (
        confPath = "./main.conf"
)

type confOpts struct {
        Name string `conf:"name" conf_extraopts:"required"`
        Age  int    `conf_extraopts:"default=19"`
        Job  struct {
                Name    string  `conf:"name" conf_extraopts:"required"`
                Address string  `conf:"address" conf_extraopts:"default=Some Address"`
                Salary  float64 `conf:"salary" conf_extraopts:"default=1.3"`
        } `conf:"job" conf_extraopts:"required"`
        FavoriteDishes []string `conf:"favorite_dishes"`
}

func main() {

        var c confOpts

        if err := conf.Load(&c, conf.Settings{
                ConfPath:    confPath,
                ConfType:    conf.ConfigTypeYAML,
                UnknownDeny: true,
        }); err != nil {
                fmt.Println(err)
                return
        }

        fmt.Printf("%+v\n", c)
}
```

Relative config file:
```yaml
name: John Doe
job:
  name: ENV:CONF_JOB_NAME
  salary: ENV:CONF_JOB_SALARY
favorite_dishes:
- apples
- ice cream
```

Run:

```
CONF_JOB_NAME="Best job" CONF_JOB_SALARY="1.0" go run main.go
```
