package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	
	"github.com/mberwanger/dockerfiles/tool/internal/config"
)

type cleanCmd struct {
	Cmd *cobra.Command
}

func newCleanCmd() *cleanCmd {
	root := &cleanCmd{}
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove generated Dockerfiles and directories",
		Long:  "Remove all generated Dockerfiles and version directories, leaving only source directories intact",
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			cfg, err := config.Load(configFile)
			if err != nil {
				return err
			}

			totalRemoved := 0
			for imageName, image := range cfg.Images {
				log.Debugf("cleaning image: %s", imageName)

				var imagePath string
				if filepath.IsAbs(image.Path) {
					imagePath = image.Path
				} else {
					basePath := cfg.Defaults.BasePath
					if basePath == "" {
						return fmt.Errorf("base path not set in config")
					}
					imagePath = filepath.Join(basePath, image.Path)
				}

				removedCount := 0
				for versionName := range image.Versions {
					versionDir := filepath.Join(imagePath, versionName)

					if _, err := os.Stat(versionDir); os.IsNotExist(err) {
						// Directory doesn't exist, skip
						continue
					}

					if err := os.RemoveAll(versionDir); err != nil {
						log.Warnf("failed to remove %s: %v", versionDir, err)
					} else {
						log.Debugf("Removed: %s", versionDir)
						removedCount++
					}
				}

				if removedCount > 0 {
					log.Infof("cleaned %s (%d versions)", imageName, removedCount)
				}
				totalRemoved += removedCount
			}

			if totalRemoved == 0 {
				log.Info("no generated directories found to clean")
			} else {
				log.Infof("cleaned %d directories successfully after %s", totalRemoved, time.Since(start).Truncate(time.Millisecond))
			}

			return nil
		},
	}

	root.Cmd = cmd
	return root
}
