package main

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
)

var (
	GIT_DIR = "../.git"
)

func init() {
	log.SetFlags(log.Lshortfile)
	dir := os.Getenv("GIT_DIR")
	if dir != "" {
		GIT_DIR = dir
	}

	if err := setDir(GIT_DIR); err != nil {
		log.Fatalf("Error setting git-dir: %s", err)
		return
	}
	os.Setenv("GIT_DIR", GIT_DIR)
}
func main() {
	app := cli.App{
		Name:  "mini-git",
		Usage: "A mini git client",
		Action: func(c *cli.Context) error {
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "init",
				Usage:   "Initialize a git repository",
				Aliases: []string{"i"},
				Action: func(c *cli.Context) error {
					// // reponame
					// repoName := c.Args().Get(0)
					// if err := validateRepoName(repoName); err != nil {
					// 	log.Fatalf("Invalid repository name: %s", err)
					// }

					for _, dir := range []string{GIT_DIR, GIT_DIR + "/objects", GIT_DIR + "/refs"} {
						if err := os.MkdirAll(dir, 0755); err != nil {
							fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
						}
					}

					headFileContents := []byte("ref: refs/heads/master\n")
					if err := os.WriteFile(GIT_DIR+"/HEAD", headFileContents, 0644); err != nil {
						fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
					}

					log.Printf("Initialized git directory GIT_DIR=[ %s ]", os.Getenv("GIT_DIR"))
					return nil
				},
			},
			{
				Name:    "cat-file",
				Usage:   "Display the contents of a file",
				Aliases: []string{"c"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "type",
						Aliases: []string{"t"},
						Usage:   "Show the object type",
					},
					&cli.BoolFlag{
						Name:    "pretty-print",
						Aliases: []string{"p"},
						Usage:   "Pretty-print the contents of the object",
					},
				},
				Action: func(c *cli.Context) error {
					hashes := c.Args().Slice()
					// println(hashes)
					for _, hash := range hashes {
						if err := catHash(hash); err != nil {
							log.Fatalf("Error reading hash %s: %s", hash, err)
						}
					}

					return nil
				},
			},
			{
				Name:    "hash-object",
				Usage:   "Compute object ID and optionally creates a blob from a file",
				Aliases: []string{"h"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "write",
						Aliases: []string{"w"},
						Usage:   "Actually write the object into the database",
					},
				},
				Action: func(c *cli.Context) error {
					println("hash-object")
					return nil
				},
			},
			{
				Name:    "log",
				Usage:   "Show commit logs",
				Aliases: []string{"l"},
				Action: func(c *cli.Context) error {
					println("not implemented yet")
					cmd := exec.Command("git", "log")
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Stdin = os.Stdin

					if err := cmd.Run(); err != nil {
						log.Fatalf("Error running git log: %s", err)
					}
					return nil
				},
			},
			{
				Name:    "ls-tree",
				Usage:   "Pretty-print a tree object",
				Aliases: []string{"ls"},
				Action: func(c *cli.Context) error {
					// read .git/object/ directories
					var dirs []fs.DirEntry
					dirs, err := os.ReadDir(GIT_DIR + "/objects")
					if err != nil {
						log.Fatalf("Error reading objects directory: %s", err)
					}

					// read the content of each directory
					log.Printf("Hash \t\t\t\t\t\t Type \t Name \t")
					for _, dir := range dirs {
						if len(dir.Name()) != 2 {
							continue
						}
						dircontents, err := os.ReadDir(GIT_DIR + "/objects/" + dir.Name())
						if err != nil {
							log.Fatalf("Error reading directory %s: %s", dir.Name(), err)
						}
						for _, file := range dircontents {
							if !file.IsDir() {
								log.Printf("%s \t %s \t %s \t", dir.Name()+file.Name(), "blob", "file")
							}

							filepath := fmt.Sprintf("%s/objects/%s/%s", GIT_DIR, dir.Name(), file.Name())
							content, err := os.ReadFile(filepath)
							if err != nil {
								log.Fatalf("Failed to read the object file %s\n", err)
								break
							}
							f, err := zlib.NewReader(bytes.NewReader(content))
							if err != nil {
								log.Fatalf("Failed to create zlib reader: %s\n", err)
							}
							defer f.Close()

							out := bytes.Buffer{}

							_, err = io.Copy(&out, f)

							if err != nil {
								log.Fatalf("Failed to read the content of the object: %s\n", err)
								break
							}

							split := strings.Split(out.String(), "\000")

							body := split[1]
							fmt.Printf("%s", body)
						}
					}
					// read the content of each file
					// print the content
					// println("not implemented yet")
					// cmd := exec.Command("git", "ls-tree", "-r", "HEAD")
					// cmd.Stdout = os.Stdout
					// cmd.Stderr = os.Stderr
					// cmd.Stdin = os.Stdin

					// if err := cmd.Run(); err != nil {
					// 	log.Fatalf("Error running git ls-tree: %s", err)
					// }
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "git-dir",
				Aliases: []string{"g"},
				Usage:   "Set the path to the repository",
				Value:   ".git",
				Action: func(_ *cli.Context, val string) error {
					if err := setDir(val); err != nil {
						log.Fatalf("Error setting git-dir: %s", err)
					}

					return nil
				},
			},
			&cli.BoolFlag{
				// verbose logging
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Verbose logging",
				Value:   false,
				Action: func(_ *cli.Context, v bool) error {
					if v {
						log.SetFlags(log.Lshortfile | log.LstdFlags)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func validateRepoName(repoName string) error {

	if repoName == "" {
		return errors.New("repository name cannot be empty")
	}

	if !strings.Contains(repoName, "github.com") {
		return errors.New("not a valid github repository name")
	}

	return nil
}

func setDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("git-dir cannot be empty")
	}

	// check if dir exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("git-dir [ %s ] does not exist. ", dir)
		return nil
	}

	GIT_DIR = dir
	os.Setenv("GIT_DIR", GIT_DIR)
	return nil
}

func catHash(hash string) error {
	if len(hash) < 3 {
		return fmt.Errorf("invalid hash: %s", hash)
	}

	filepath := fmt.Sprintf("%s/objects/%s/%s", GIT_DIR, hash[:2], hash[2:])
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read the object file %s", err)
	}
	in := bytes.NewReader(content)
	r, err := zlib.NewReader(in)
	if err != nil {
		return fmt.Errorf("failed to create zlib reader: %s", err)
	}

	defer r.Close()
	out := bytes.Buffer{}
	_, err = io.Copy(&out, r)
	if err != nil {
		return fmt.Errorf("failed to read the content of the object: %s", err)
	}
	split := strings.Split(out.String(), "\000")
	body := split[1]
	fmt.Println(body)
	return nil
}
