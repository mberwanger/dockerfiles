package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mberwanger/dockerfiles/tooling/internal/builder"
)

func main() {
	var (
		generateAll  = flag.Bool("generate-all", false, "Generate all Dockerfiles")
		generateImg  = flag.String("generate", "", "Generate Dockerfiles for specific image")
		depIndex     = flag.Bool("dependency-index", false, "Generate dependency index for CI")
		manifestFile = flag.String("manifest", "dockerfiles/manifest.yaml", "Path to manifest file")
	)
	flag.Parse()

	manifest, err := builder.LoadManifest(*manifestFile)
	if err != nil {
		log.Fatalf("Failed to load manifest: %v", err)
	}

	switch {
	case *generateAll:
		if err := builder.GenerateAll(manifest); err != nil {
			log.Fatalf("Failed to generate all: %v", err)
		}
		fmt.Println("✅ All Dockerfiles generated successfully")

	case *generateImg != "":
		if err := builder.GenerateImage(manifest, *generateImg); err != nil {
			log.Fatalf("Failed to generate %s: %v", *generateImg, err)
		}
		fmt.Printf("✅ %s Dockerfiles generated successfully\n", *generateImg)

	case *depIndex:
		index, err := builder.GenerateDependencyIndex(".")
		if err != nil {
			log.Fatalf("Failed to generate dependency index: %v", err)
		}

		jsonOutput, err := json.Marshal(index)
		if err != nil {
			log.Fatalf("Failed to marshal index to JSON: %v", err)
		}

		fmt.Print(string(jsonOutput))

	default:
		flag.Usage()
		os.Exit(1)
	}
}
