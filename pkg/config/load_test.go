package config

import "fmt"

func ExampleDefaultConfigSearchPath() {
	fmt.Println(DefaultConfigSearchPath("ace", "dt", "config.yaml"))
}
