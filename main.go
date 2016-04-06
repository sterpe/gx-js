/*
Package gx-js provides node specific hooks to be called by the gx tool.
*/
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cli "github.com/codegangsta/cli"
)

var cwd string

var installPathHookCommand = cli.Command{
	Name: "install-path",
	Usage: "prints out the install path",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name: "global",
			Usage: "print the global install directory",
		},
	},
	Action: func(c *cli.Context) {
		var npath string
		if c.Bool("global") {
			n, err := getNodejsExecutablePath()
			if err != nil {
				log.Fatal("failed to get node path:", err)
			}
			npath = getGlobalNodeModulesFolderPath(n)
		} else {
			wd, err := os.Getwd()
			if err != nil {
				log.Fatal("install-path cwd:", err)
			}
			npath = wd
		}
		fmt.Println(filepath.Join(npath, "node_modules"))
	},
}

// HookCommand provides `gx-js hook [hook] [args]`.
var HookCommand = cli.Command{
	Name: "hook",
	Usage: "node specific hooks to be called by the gx tool",
	Subcommands: []cli.Command{
		installPathHookCommand,
	},
	Action: func(c *cli.Context) {},
}

func main() {
	app := cli.NewApp()
	app.Name = "gx-js"
	app.Author = "sterpe"
	app.Usage = "gx extensions for node.js"
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "verbose",
			Usage: "turn on verbose output",
		},
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("failed to get cwd:", err)
	}
	cwd = wd

	app.Commands = []cli.Command{
		HookCommand,
		{
			Name: "install-path",
			Category: "hook",
		},
	}

	app.Run(os.Args)
}

func getGlobalNodeModulesFolderPath(n string) (string) {
	return filepath.Join(n, "..", "lib")
}
func getNodejsExecutablePath() (string, error) {
	out, err := exec.Command("which", "node").Output()
	node := strings.TrimSpace(string(out[:]))

	if err != nil {
		return node, err
	} else if node == "" {
		return node, fmt.Errorf("node binary not found")
	}

	dir, _ := filepath.Split(node)

	return dir, nil
}
