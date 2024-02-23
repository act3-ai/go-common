#import "codereport.template.typ": codereport

#let fileList = csv("codereport.files.csv")

// Map list of files to dict of files
#let files = fileList.map(file => (
  name: file.at(0),
  lang: file.at(1),
  data: read(file.at(0), encoding: none),
))

#show: doc => codereport(
	title: "go-common Source Code",
	author: "ACT3",
	description: [
		The go-common repository contains common Go packages for ACT3 projects. The packages are tailored for CLI and REST API development. The "cmd/sample" directory contains a sample CLI application demonstrating the use of go-common's packages for CLI development.
	],
	sourcefiles: files
)
