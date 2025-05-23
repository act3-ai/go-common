---
title: sample sample-config
description: Help for sample CLI configuration
---

<!--
This documentation is auto generated by a script.
Please do not edit this file directly.
-->

<!-- markdownlint-disable-next-line single-title -->
# Configuration Options

Table of contents:

- [example](#example): Example options

## example

Example options

| Option                      | Description                    |
| --------------------------- | ------------------------------ |
| [`name`](#name)             | Your name.                     |
| [`--greeting`](#--greeting) | Greeting for the user.         |
| [`--count`](#--count)       | Number of greetings to output. |
| [`--excited`](#--excited)   | Greet with excitement.         |

### `name`

Your name.

| Name      | Value             |
| --------- | ----------------- |
| type      | string            |
| json/yaml | `name`            |
| cli       | `--name`, `-n`    |
| env       | `ACE_SAMPLE_NAME` |

Name of the sample CLI's user.

### `--greeting`

Greeting for the user.

| Name    | Value                 |
| ------- | --------------------- |
| type    | string                |
| default | `"Hello"`             |
| cli     | `--greeting`, `-g`    |
| env     | `ACE_SAMPLE_GREETING` |

### `--count`

Number of greetings to output.

| Name    | Value              |
| ------- | ------------------ |
| type    | integer            |
| default | `1`                |
| cli     | `--count`, `-c`    |
| env     | `ACE_SAMPLE_COUNT` |

### `--excited`

Greet with excitement.

| Name    | Value                |
| ------- | -------------------- |
| type    | boolean              |
| default | `false`              |
| cli     | `--excited`, `-e`    |
| env     | `ACE_SAMPLE_EXCITED` |
