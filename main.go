/*
Package gx-js provides node specific hooks to be called by the gx tool.
*/
package main

import (
        "encoding/json"
	"fmt"
        "io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cli "github.com/codegangsta/cli"
)

var cwd string

var postInstallHookCommand = cli.Command{
	Name: "post-install",
	Usage: "post install hook for newly installed node packages",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name: "global",
			Usage: "specifies whether or not the install was global",
		},
	},
	Action: func (c *cli.Context) {
		if !c.Args().Present() {
			log.Fatal("must specify path to newly installed package")
		}
		path := c.Args().First()
		packageJSON := ""
		recurse := false
		err := filepath.Walk(path, func(path string, fileinfo os.FileInfo, err error) error {
			if fileinfo.Name() == "package.json" && fileinfo.IsDir() == false {
				fmt.Println(path)
				packageJSON = path
				recurse = false
			}
			if recurse == true {
				return filepath.SkipDir
			}
			return err
		})
		if err != nil {
			log.Fatal("post-install could not find package.json")
		}
		fmt.Println(packageJSON)
		fmt.Println(getNodeModulesBinFolder(path, c.Bool("global")))
		binFolder := getNodeModulesBinFolder(path, c.Bool("global"))
		packageRoot := filepath.Dir(packageJSON)
		content, err := ioutil.ReadFile(packageJSON)
		if err != nil {
			log.Fatal("post-install could not read package.json")
		}

		var dat map[string]interface{}
		if err := json.Unmarshal(content, &dat); err != nil {
			log.Fatal("post-install could not parse package.json")
		}
		fmt.Println(dat)
		name := dat["name"].(string)
		var bin map[string]interface{}
		var binString string
		binObject, ok := dat["bin"].(map[string]interface{})
		if !ok {
			binString, ok = dat["bin"].(string)
			if !ok {
			}
		}
		fmt.Println(binObject, binString)
		if len(binObject) != 0 {
			bin = binObject
		}
		if len(binString) != 0 {
			bin = make(map[string]interface{})
			bin[name] = binString
		}
		fmt.Println(bin)
		for executable, location := range bin {
			f1 := filepath.Join(packageRoot, location.(string))
			f2 := filepath.Join(binFolder, executable)
			fmt.Println(f2, " -> ", f1)
			var perm os.FileMode = 0755
			err := os.MkdirAll(binFolder, perm)
			if err != nil {
				log.Fatal("install-path could not create bin directory")
			}
			err = os.Symlink(f1, f2)
			if (err != nil) {
				log.Fatal("install-path could not create symlink:", err)
			}
			err = os.Chmod(f2, perm)
			if (err != nil) {
				log.Fatal("install-path could not change mode on the bin file", err)
			}
		}
	},
}
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
				log.Fatal("install-path failed to get node path:", err)
			}
			npath = n
		} else {
			wd, err := os.Getwd()
			if err != nil {
				log.Fatal("install-path failed to get cwd:", err)
			}
			npath = wd
		}
		npath = getNodeModulesFolderPath(npath, c.Bool("global"))
		fmt.Println(npath)
	},
}

// HookCommand provides `gx-js hook [hook] [args]`.
var HookCommand = cli.Command{
	Name: "hook",
	Usage: "node specific hooks to be called by the gx tool",
	Subcommands: []cli.Command{
		installPathHookCommand,
		postInstallHookCommand,
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

func getNodeModulesBinFolder(path string, global bool) (string) {
	dir := filepath.Dir(path)
	base := filepath.Base(dir)
	dir = filepath.Join(dir, "..")

	for base != "node_modules" && (base != "." && base != "/") {
		base = filepath.Base(dir)
		dir = filepath.Join(dir, "..")
	}
	dir = filepath.Join(dir, "node_modules")
	if global {
		bin := filepath.Join(dir, "..", "..", "bin")
		return bin
	}
	dotbin := filepath.Join(dir, ".bin")
	return dotbin
}
func getNodeModulesFolderPath(path string, global bool) (string) {
	if global {
		path = filepath.Join(path, "..", "lib")
	}

	return filepath.Join(path, "node_modules")
}

func getNodejsExecutablePath() (string, error) {
	out, err := exec.Command("node", "-p", "process.execPath").Output()
	node := strings.TrimSpace(string(out[:]))

	if err != nil {
		return node, err
	} else if node == "" {
		return node, fmt.Errorf("node binary not found")
	}

	dir, _ := filepath.Split(node)

	return dir, nil
}
