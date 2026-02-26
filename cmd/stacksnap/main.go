package main

import (
	"flag"
	"log"
	"os"

	"stacksnap/internal/docker"
	"stacksnap/internal/portainer"
	"stacksnap/internal/snapshot"
)

func main() {
	outputDir := flag.String("output", "./snapshots", "Directory to store stack snapshots")
	keep := flag.Int("keep", 5, "Number of historical versions to keep per stack (min 1)")
	portainerURL := flag.String("portainer-url", "", "Portainer base URL (e.g. http://localhost:9000)")
	portainerToken := flag.String("portainer-token", "", "Portainer API access token")
	flag.Parse()

	// Environment variable fallback (safer than CLI flags for secrets)
	if *portainerURL == "" {
		*portainerURL = os.Getenv("STACKSNAP_PORTAINER_URL")
	}
	if *portainerToken == "" {
		*portainerToken = os.Getenv("STACKSNAP_PORTAINER_TOKEN")
	}

	if *keep < 1 {
		log.Fatal("--keep must be at least 1")
	}

	store := snapshot.NewStore(*outputDir, *keep)
	portainerStacks := make(map[string]bool)

	// ── 1. Portainer (preferred source) ────────────────────────────────────────
	if *portainerURL != "" && *portainerToken != "" {
		client := portainer.NewClient(*portainerURL, *portainerToken)

		stacks, err := client.GetStacks()
		if err != nil {
			log.Fatalf("Portainer: failed to list stacks: %v", err)
		}

		for _, stack := range stacks {
			portainerStacks[stack.Name] = true

			yamlData, err := client.GetStackFile(stack.ID)
			if err != nil {
				log.Printf("Portainer: skipping stack %q: %v", stack.Name, err)
				continue
			}

			if err := store.Save(stack.Name, yamlData); err != nil {
				log.Printf("Snapshot: failed to save %q: %v", stack.Name, err)
			}
		}
	}

	// ── 2. Docker fallback (stacks not managed by Portainer) ───────────────────
	projects, err := docker.DiscoverProjects()
	if err != nil {
		log.Printf("Docker: discovery failed (is Docker running?): %v", err)
	}

	for _, project := range projects {
		if portainerStacks[project.Name] {
			continue // already handled above
		}

		yamlData := project.ToCompose()

		if err := store.Save(project.Name, yamlData); err != nil {
			log.Printf("Snapshot: failed to save %q: %v", project.Name, err)
		}
	}

	log.Println("StackSnap: done.")
}
