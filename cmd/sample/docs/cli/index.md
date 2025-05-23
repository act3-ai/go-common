---
title: sample
description: 
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# sample

## Usage

```plaintext
sample [flags]
```

## Examples

```sh
# Run sample:
sample

# Run sample with name set by flag:
sample --name "Foo"

# Run sample with name set by environment variable:
ACE_SAMPLE_NAME="Foo" sample
```

## Options

```plaintext
OPTIONS:
  -h, --help                           help for sample
  -v, --verbosity stringSlice[=warn]   Logging verbosity level (also setable with environment variable ACE_SAMPLE_VERBOSITY)
                                       Aliases: error=0, warn=4, info=8, debug=12 (default [error])

EXAMPLE OPTIONS:
  -c, --count int         Number of greetings to output. (env: ACE_SAMPLE_COUNT) (default 1)
  -e, --excited           Greet with excitement. (env: ACE_SAMPLE_EXCITED)
  -g, --greeting string   Greeting for the user. (env: ACE_SAMPLE_GREETING) (default "Hello")
  -n, --name string       Your name. (env: ACE_SAMPLE_NAME)
```

## Subcommands

- [`sample completion`](completion/index.md) - Generate the autocompletion script for the specified shell
- [`sample gendocs`](gendocs/index.md) - Generate documentation for the tool in various formats
- [`sample genschema`](genschema.md) - Outputs configuration file validators
- [`sample info`](info/index.md) - View detailed documentation for the tool
- [`sample sample-config`](sample-config.md) - Help for sample CLI configuration
- [`sample testfile`](testfile.md) - Help command that displays the test file
- [`sample version`](version.md) - Print the version
