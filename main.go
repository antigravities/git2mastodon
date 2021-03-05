package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/mattn/go-mastodon"
)

type mastoCfg struct {
	Server       string
	ClientID     string
	ClientSecret string
	AccessToken  string
	Commit       string
}

type tmplInput struct {
	Commit     string
	Repository string
}

func run(repo *string, refspec *string, instance *string, storage *string, force *bool, tmpl *string) {
	if repo == nil || *repo == "" {
		log.Fatalf("No repo provided. Provide a repo using -repo.")
	}

	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		URLs: []string{
			*repo,
		},
	})

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		log.Fatalf("Error fetching remote: %v", err)
	}

	commit := ""

	for _, ref := range refs {
		if ref.Name().String() == *refspec && !ref.Hash().IsZero() {
			commit = ref.Hash().String()
		}
	}

	if commit == "" {
		log.Printf("No commits matching refspec")
		return
	}

	if _, err := os.Stat(*storage); os.IsNotExist(err) {
		app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
			Server:     *instance,
			ClientName: "git2mastodon",
			Scopes:     "read write",
			Website:    "https://get.cutie.cafe/git2mastodon",
		})
		if err != nil {
			log.Fatalf("Could not register app: %v", err)
		}

		cfg := &mastoCfg{}
		cfg.ClientID = app.ClientID
		cfg.ClientSecret = app.ClientSecret
		cfg.Server = *instance
		cfg.Commit = commit

		for {
			log.Printf("Click or copy the following link to authorize with your instance:")
			log.Printf("%s/oauth/authorize?client_id=%s&scope=read+write&redirect_uri=urn:ietf:wg:oauth:2.0:oob&response_type=code", *instance, cfg.ClientID)
			log.Printf("Then paste the token code provided:")

			token := ""
			fmt.Scanf("%s", &token)

			client := mastodon.NewClient(&mastodon.Config{
				Server:       *instance,
				ClientID:     app.ClientID,
				ClientSecret: app.ClientSecret,
			})

			err := client.AuthenticateToken(context.Background(), token, "urn:ietf:wg:oauth:2.0:oob")

			if err != nil {
				log.Printf("Error validating token: %v", err)
			} else {
				cfg.AccessToken = client.Config.AccessToken
				jx, _ := json.Marshal(cfg)

				if err := os.WriteFile(*storage, jx, 0660); err != nil {
					log.Fatalf("Could not write %s: %v", *storage, err)
				}
				break
			}
		}

		log.Printf("Credentials saved.")
	}

	cfg := &mastoCfg{}
	file, err := os.Open(*storage)
	if err != nil {
		log.Fatalf("Error opening %s: %v", *storage, err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	if err := dec.Decode(cfg); err != nil {
		log.Fatalf("Error decoding %s: %v", *storage, err)
	}

	if cfg.Commit == commit && !*force {
		log.Printf("Commit %s has already been announced", commit)
		return
	}

	client := mastodon.NewClient(&mastodon.Config{
		Server:       cfg.Server,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		AccessToken:  cfg.AccessToken,
	})

	st := fmt.Sprintf("Repository %s was pushed to %s", *repo, commit)

	if *tmpl != "" {
		var buf bytes.Buffer

		x, err := os.ReadFile(*tmpl)
		if err != nil {
			log.Fatalf("Error reading template: %v", err)
		}

		template.Must(template.New("status").Parse(string(x))).Execute(&buf, tmplInput{commit, *repo})
		st = buf.String()
	}

	status, err := client.PostStatus(context.Background(), &mastodon.Toot{
		Status:     st,
		Visibility: "unlisted",
	})

	cfg.Commit = commit

	jx, _ := json.Marshal(cfg)
	os.WriteFile(*storage, jx, 0660)

	if err != nil {
		log.Fatalf("Could not post status: %v", err)
	}

	log.Printf("Posted status %s", status.URL)
}

func main() {
	repo := flag.String("repo", "", "The repository to fetch.")
	refspec := flag.String("refspec", "refs/heads/master", "The refspec to compare commits with.")
	instance := flag.String("instance", "https://mastodon.social", "The Mastodon (or Mastodon-compatible) instance to interface with (only required on first run)")
	storage := flag.String("storage", "masto.cfg", "File to store settings/data in.")
	force := flag.Bool("force", false, "Post a commit again even if it's already been posted.")
	tmpl := flag.String("tmpl", "", "A Go template used to encode a status to post. {{ .Repository }} and {{ .Commit }} are available.")
	runEvery := flag.Uint("run-every", 0, "If > 0, git2mastodon will run in the foreground and run a check every X seconds.")

	flag.Parse()

	if *runEvery == 0 {
		run(repo, refspec, instance, storage, force, tmpl)
	} else {
		for {
			run(repo, refspec, instance, storage, force, tmpl)
			time.Sleep(time.Duration(*runEvery) * time.Second)
		}
	}
}
