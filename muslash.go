package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

//AppVer define current application version
const AppVer = "0.0.0"

func initApp(c *cli.Context) error {
	log.SetFormatter(&log.TextFormatter{})
	if c.IsSet("verbose") {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	return nil
}

var logDetails = log.Fields{}

func Action(c *cli.Context) error {

	if len(c.Args()) < 2 {
		return fmt.Errorf("too few args")
	}

	todir := c.Args().First()
	for i := 1; i < len(c.Args()); i++ {
		fromdir := c.Args().Get(i)

		if fromdir == "" || todir == "" {
			return fmt.Errorf("missing argument")
		}
		logDetails["fromdir"] = fromdir
		logDetails["todir"] = todir

		fromdirInfo, err := os.Stat(fromdir)
		if err != nil {
			return err
		}
		if !fromdirInfo.IsDir() {
			return fmt.Errorf("fromdir is not a directory")
		}

		todirInfo, err := os.Stat(todir)
		if err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(todir, os.ModePerm); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		if !todirInfo.IsDir() {
			return fmt.Errorf("todir is not a directory")
		}

		err = filepath.Walk(fromdir, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				log.WithFields(log.Fields{"message": err.Error(), "dir": path}).
					Error("walk directory error")
				return err
			}

			rel, err := filepath.Rel(filepath.Dir(fromdir), path)
			if err != nil {
				log.WithFields(log.Fields{"message": err.Error(), "dir": path}).
					Error("rel error")
				return err
			}
			tofile := filepath.Join(todir, rel)
			tofileDir := filepath.Dir(tofile)
			_, err = os.Stat(tofileDir)
			if err != nil {
				if os.IsNotExist(err) {
					if err := os.MkdirAll(tofileDir, os.ModePerm); err != nil {
						log.WithFields(log.Fields{"message": err.Error(), "tofileDir": tofileDir}).
							Error("mkdirAll error")
						return err
					}
				} else {
					log.WithFields(log.Fields{"message": err.Error(), "tofileDir": tofileDir}).
						Error("Stat file error")
					return err
				}
			}

			if !info.IsDir() && (filepath.Ext(tofile) == ".flac" || filepath.Ext(tofile) == ".m4a") {
				tofile = strings.Replace(tofile, filepath.Ext(tofile), ".mp3", 1)
				out, err := exec.Command("ffmpeg", "-i", path, "-ab", "256k", "-map_metadata", "0", "-id3v2_version", "3", tofile).Output()
				if err != nil {
					log.WithFields(log.Fields{"message": err.Error(), "from": path, "to": tofile, "out": out, "cmd": fmt.Sprintf(`ffmpeg -i "%s" -ab 256k 0map_metadata 0 -id3v2_version 3 "%s"`, path, tofile)}).
						Error("ffmpeg error")
					return err
				} else {
					log.WithFields(log.Fields{"from": path, "to": tofile}).
						Debug("seccess converting")
				}
			} else if !info.IsDir() && filepath.Ext(path) == ".mp3" {
				out, err := exec.Command("rsync", "-Prh", path, tofile).Output()
				if err != nil {
					log.WithFields(log.Fields{"message": err.Error(), "from": path, "to": tofile, "out": string(out), "cmd": fmt.Sprintf(`rsync -Prh "%s" "%s"`, path, tofile)}).
						Error("ffmpeg error")
					return err
				} else {
					log.WithFields(log.Fields{"from": path, "to": tofile}).
						Debug("seccess copy")
				}
			}
			return nil
		})
		if err != nil {
			log.WithFields(log.Fields{"message": err.Error(), "dir": fromdir}).
				Error("error walking dir")
		}
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "muslash"
	app.Usage = "Music flash creator"
	app.Version = AppVer
	app.Flags = append(app.Flags, cli.BoolFlag{
		Name:  "verbose, vv",
		Usage: "Enable verbose mode",
	})
	app.Before = initApp
	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Action = Action
	if err := app.Run(os.Args); err != nil {
		log.WithFields(logDetails).Error(err.Error())
	}
}
