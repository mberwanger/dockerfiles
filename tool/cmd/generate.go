package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/apex/log"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/mberwanger/dockerfiles/tool/internal/config"
	"github.com/mberwanger/dockerfiles/tool/internal/generator"
	"github.com/mberwanger/dockerfiles/tool/internal/workflow"
)

var (
	boldStyle = lipgloss.NewStyle().Bold(true)
)

type generatorCmd struct {
	Cmd *cobra.Command
}

func newGeneratorCmd() *generatorCmd {
	root := &generatorCmd{}
	cmd := &cobra.Command{
		Use:               "generate",
		Aliases:           []string{"gen"},
		Short:             "Generate Dockerfiles from templates",
		Long:              `Generate Dockerfiles from templates for all images, specific images, or GitHub Actions workflows`,
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	var generateAll bool
	imageSubCmd := &cobra.Command{
		Use:     "image [image-name]",
		Aliases: []string{"img"},
		Short:   "Generate Dockerfiles for a specific image or all images",
		Long:    "Generate Dockerfiles for a specific image, or all images with the --all flag",
		Example: `  # Generate a specific image
  dockerfiles generate image core

  # Generate all images
  dockerfiles generate image --all
  dockerfiles generate image -A`,
		Args: func(cmd *cobra.Command, args []string) error {
			if generateAll && len(args) > 0 {
				return fmt.Errorf("cannot specify image name with --all flag")
			}
			if !generateAll && len(args) != 1 {
				_ = cmd.Help()
				os.Exit(0)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			cfg, err := config.Load(configFile)
			if err != nil {
				return err
			}

			if generateAll {
				if err := generator.GenerateAll(cfg); err != nil {
					log.Fatalf("Failed to generate all images: %v", err)
				}

				imageCount := len(cfg.Images)
				log.Info(boldStyle.Render(fmt.Sprintf("generated %d images successfully after %s", imageCount, time.Since(start).Truncate(time.Second))))
			} else {
				imageName := args[0]
				if err := generator.GenerateImage(cfg, imageName); err != nil {
					log.Fatalf("Failed to generate image '%s': %v", imageName, err)
				}

				image := cfg.Images[imageName]
				versionCount := len(image.Versions)
				log.Info(boldStyle.Render(fmt.Sprintf("generated image '%s' (%d versions) successfully after %s", imageName, versionCount, time.Since(start).Truncate(time.Second))))
			}
			return nil
		},
	}
	imageSubCmd.Flags().BoolVarP(&generateAll, "all", "A", false, "Generate all images")

	var outputFile string
	workflowSubCmd := &cobra.Command{
		Use:     "workflow",
		Aliases: []string{"wf"},
		Short:   "Generate GitHub Actions workflow (outputs to stdout by default)",
		Long:    "Generate a GitHub Actions workflow file with dependency-ordered build jobs. Outputs to stdout by default, or to a file if specified with --output/-o",
		Example: `  # Output to stdout
  dockerfiles generate workflow

  # Output to file
  dockerfiles generate workflow -o .github/workflows/dockerfiles.yaml`,
		Args: cobra.NoArgs,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Disable logging when writing to stdout
			if outputFile == "" {
				log.SetLevel(log.FatalLevel)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configFile)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if outputFile != "" {
				if err := workflow.Generate(cfg, outputFile); err != nil {
					return fmt.Errorf("generating workflow: %w", err)
				}
				log.Infof("Generated workflow file: %s", outputFile)
			} else {
				if err := workflow.GenerateToWriter(cfg, os.Stdout); err != nil {
					return fmt.Errorf("generating workflow: %w", err)
				}
			}
			return nil
		},
	}
	workflowSubCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (defaults to stdout)")

	cmd.AddCommand(
		imageSubCmd,
		workflowSubCmd,
	)
	root.Cmd = cmd
	return root
}
