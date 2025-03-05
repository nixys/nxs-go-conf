# nxs-go-conf

## Introduction

Go library provides a simple way to handle configuration parameters for Go programs.

### Features

- Extrimally simple to use this library for create a config files in YAML or JSON formats
- Defining config files structure and option extra settings with Go structs and tags
- Using environment variables as an option values

---

- **Manage options in structure field tags**  
  To describe configuration file structure you simply need to define the struct in the Go program code. In that struct you can use field tags to set different options and to determine config file decoding behavior. Currently, the next tags are available:
  - `conf`: defines custom name for an option
  - `conf_extraopts`: provides advanced settings for option. This tag may have the following values:
    - `required`: option with this tag is mandatory. If it is set, but corresponding option is not defined in the config file, it will cause an error.
    - `default`: determines default value for the option.

- **ENV variables as option values**  
  You may specify the option value as `ENV:VARIABLE_NAME`. It will use the value of the relative environment variable (i.e. _VARIABLE_NAME_) as value for that option.

- **YAML and JSON formats are available**  
  Currently, you can use config files in YAML or JSON formats. To switch the format you only need to specify the appropriate setting for config file load function.

- **Catch the unknown options**  
  You can catch options, that are contained in config file but has no matching in the result interface.

### Who can use the tool

Developers who creates an applications that using config files.

### Import

```go
import "github.com/nixys/nxs-go-conf"
```

### Initialize

To initialize this library you need to do a few steps

#### Define a config structure

You need to define a Go structure with tags. Tags may contains a following special fields:
- `conf`: defines a name for option
- `conf_extraopts`: provides advanced settings for option. This tag may have the following values:
  - `required`: option with this tag is mandatory. If it is set but corresponding option is not defined in the config file it will cause an error
  - `default`: determines default value for the option

Struct you defined as config may contains a fields with various types including other structs.

See example a config struct definition below.

Let's say we need to get for our application some data:
- Logfile path 
- Settings to connect to MySQL server

A config struct for this issue may looks as follows:
```go
type confOpts struct {
	LogFile  string `conf:"logfile" conf_extraopts:"default=stderr"`
	MySQL    *mysqlConf            `conf:"mysql"`
}

type mysqlConf struct {
	Host     string `conf:"host" conf_extraopts:"required"`
	Port     int    `conf:"port" conf_extraopts:"required"`
	DB       string `conf:"db" conf_extraopts:"required"`
	User     string `conf:"user" conf_extraopts:"required"`
	Password string `conf:"password" conf_extraopts:"required"`
}
```

This config struct will impose the following requirements to the configuration file:
- If option `logfile` is specified this set a path to log file. Otherwise `stderr` will be used
- Option `mysql` is not mandaory and if it not specified in config file an appropriate variable will be `nil`. But if option `mysql` is specified then options `host`, `port`, `db`, `user` and `password` must be specified

#### Load config

Since config struct was defined you need to read a config file and load a values with function:
```go
conf.Load(conf interface{}, s conf.Settings) error
```

where:
- `conf` arg it is a config struct variable
- `s` arg it is a settings to load config with following fields:
  - `ConfPath`: 

## How to use



## Example

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

**For more examples see apps based on this library:**
- [nxs-support-bot](https://github.com/nixys/nxs-support-bot)
- [nxs-data-anonymizer](https://github.com/nixys/nxs-data-anonymizer)
- [nxs-backup](https://github.com/nixys/nxs-backup)
- And others

## Feedback

For support and feedback please contact me:
- telegram: [@borisershov](https://t.me/borisershov)
- e-mail: b.ershov@nixys.ru

## License

nxs-go-conf is released under the [MIT License](LICENSE).
