package config

import "fmt"

func ExampleDefaultConfigSearchPath() {
	fmt.Println(DefaultConfigSearchPath("sample", "cli", "config.yaml"))
}
