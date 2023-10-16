package main

import (
	"context"
	"flag"
	"log"
	"opentofu-stats/internal/db"
	"opentofu-stats/internal/github"
	"os"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func main() {
	client, err := github.NewClient(context.Background(), os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		log.Fatalf("cannot init github client: %v\n", err)
	}

	var (
		dbPath    string
		repoOwner string
		repoName  string
	)
	flag.StringVar(&dbPath, "db", "", "Path to the SQLLite DB file.")
	flag.StringVar(&repoOwner, "owner", "opentofu", "GitHub repository owner.")
	flag.StringVar(&repoName, "name", "opentofu", "GitHub repository name.")
	flag.Parse()

	if dbPath == "" {
		println("ERROR: specify SQLLite DB file")
		flag.Usage()
		os.Exit(1)
	}

	c, err := sqlite.OpenConn(dbPath, sqlite.OpenReadWrite)
	if err != nil {
		log.Fatalf("cannot init SQLLite client: %v\n", err)
	}

	defer func() { _ = c.Close() }()

	const table = "main.assets"

	if err := db.CreateTableIfNotExists(c, table, []db.ColumnDefinition{
		{
			Name: "id",
			Type: "INTEGER PRIMARY KEY AUTOINCREMENT",
		},
		{
			Name: "release_name",
			Type: "TEXT NOT NULL",
		},
		{
			Name: "asset_name",
			Type: "TEXT NOT NULL",
		},
		{
			Name: "download_count",
			Type: "INTEGER NOT NULL",
		},
		{
			Name: "fetched_at",
			Type: "TEXT NOT NULL",
		},
	}); err != nil {
		log.Fatalf("cannot create the output table '%s': %v\n", table, err)
	}

	o, err := github.FetchReleaseAssets(client, repoOwner, repoName)
	if err != nil {
		log.Fatalln(err)
	}

	ts := time.Now().UTC().Format(time.RFC3339)

	for _, row := range o {
		err := sqlitex.ExecuteTransient(c, "INSERT INTO "+table+
			" (release_name, asset_name, download_count, fetched_at) VALUES(?, ?, ?, ?)",
			&sqlitex.ExecOptions{
				Args: []interface{}{row.ReleaseName, row.AssetName, row.DownloadCount, ts},
			},
		)
		if err != nil {
			log.Fatalf("cannot write to table: %v\n", err)
		}
	}
}
