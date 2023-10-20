package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Data struct {
	Date                   []string  `json:"date"`
	Commits                []int     `json:"commits"`
	Committers             []int     `json:"committers"`
	CommittersNew          []int     `json:"committers_new"`
	CommittersRecurrent    []int     `json:"committers_recurrent"`
	IssuesNew              []int     `json:"issues_new"`
	IssuesClosed           []int     `json:"issues_closed"`
	IssuesOpenTotal        []int     `json:"issues_open_total"`
	IssuesClosedTotal      []int     `json:"issues_closed_total"`
	PrNew                  []int     `json:"pr_new"`
	PrClosed               []int     `json:"pr_closed"`
	PrMerged               []int     `json:"pr_merged"`
	PrOpenTotal            []int     `json:"pr_open_total"`
	PrClosedTotal          []int     `json:"pr_closed_total"`
	PrTimeToMergeMeanHours []float64 `json:"pr_time_to_merge_mean_hours"`
	StarsNew               []int     `json:"stars_new"`
	StarsTotal             []int     `json:"stars_total"`
	DownloadsTotal         []int     `json:"downloads_total"`
}

type Summary struct {
	Stars               int `json:"stars"`
	Downloads           int `json:"downloads"`
	PrOpen              int `json:"pr_open"`
	IssuesOpen          int `json:"issues_open"`
	Committers          int `json:"committers"`
	CommittersRecurrent int `json:"committers_recurrent"`
}

type Output struct {
	Summary    Summary         `json:"summary"`
	Timeseries map[string]Data `json:"timeseries"`
	UpdatedAt  string          `json:"updated_at"`
}

func main() {
	var (
		dbPath   string
		outPath  string
		htmlPath string
	)
	flag.StringVar(&dbPath, "db", "", "Path to the SQLLite DB file.")
	flag.StringVar(&outPath, "out", "/tmp/output.json", "Path to store the calculation results.")
	flag.StringVar(&htmlPath, "html", "public/index.html", "Path to store generated HTML page.")
	flag.Parse()

	if dbPath == "" {
		println("ERROR: specify SQLLite DB file")
		flag.Usage()
		os.Exit(1)
	}

	c, err := sqlite.OpenConn(dbPath, sqlite.OpenReadOnly)
	if err != nil {
		log.Fatalf("cannot init SQLLite client: %v\n", err)
	}
	defer func() { _ = c.Close() }()

	o := Output{
		Summary:    Summary{},
		Timeseries: map[string]Data{},
	}

	for key, query := range queries {
		d := Data{
			Date:                   []string{},
			Commits:                []int{},
			Committers:             []int{},
			CommittersNew:          []int{},
			CommittersRecurrent:    []int{},
			IssuesNew:              []int{},
			IssuesClosed:           []int{},
			IssuesOpenTotal:        []int{},
			IssuesClosedTotal:      []int{},
			PrNew:                  []int{},
			PrClosed:               []int{},
			PrMerged:               []int{},
			PrOpenTotal:            []int{},
			PrClosedTotal:          []int{},
			PrTimeToMergeMeanHours: []float64{},
			StarsNew:               []int{},
			StarsTotal:             []int{},
			DownloadsTotal:         []int{},
		}

		if err := sqlitex.ExecuteTransient(c, query, &sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				d.Date = append(d.Date, convertDateMonth(stmt.ColumnText(0)))
				d.Commits = append(d.Commits, stmt.ColumnInt(1))
				d.Committers = append(d.Committers, stmt.ColumnInt(2))
				d.CommittersNew = append(d.CommittersNew, stmt.ColumnInt(3))
				d.CommittersRecurrent = append(d.CommittersRecurrent, stmt.ColumnInt(4))
				d.IssuesNew = append(d.IssuesNew, stmt.ColumnInt(5))
				d.IssuesClosed = append(d.IssuesClosed, stmt.ColumnInt(6))
				d.IssuesOpenTotal = append(d.IssuesOpenTotal, stmt.ColumnInt(7))
				d.IssuesClosedTotal = append(d.IssuesClosedTotal, stmt.ColumnInt(8))
				d.PrNew = append(d.PrNew, stmt.ColumnInt(9))
				d.PrClosed = append(d.PrClosed, stmt.ColumnInt(10))
				d.PrMerged = append(d.PrMerged, stmt.ColumnInt(11))
				d.PrOpenTotal = append(d.PrOpenTotal, stmt.ColumnInt(12))
				d.PrClosedTotal = append(d.PrClosedTotal, stmt.ColumnInt(13))
				d.PrTimeToMergeMeanHours = append(d.PrTimeToMergeMeanHours, stmt.ColumnFloat(14))
				d.StarsNew = append(d.StarsNew, stmt.ColumnInt(15))
				d.StarsTotal = append(d.StarsTotal, stmt.ColumnInt(16))
				d.DownloadsTotal = append(d.DownloadsTotal, stmt.ColumnInt(17))
				return nil
			},
		}); err != nil {
			log.Fatalf("cannot execute query: %v\n", err)
		}

		o.Timeseries[key] = d
	}

	generateSummary(&o)
	o.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	outBytes, err := json.Marshal(o)
	if err != nil {
		log.Fatalf("cannot marshal output: %v\n", err)
	}

	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("cannot open file for saving %s: %v\n", outPath, err)
	}
	defer func() { _ = f.Close() }()

	_, err = f.Write(outBytes)
	if err != nil {
		log.Fatalf("cannot write to file: %v\n", err)
	}

	fHTML, err := os.OpenFile(htmlPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("cannot open file for saving HTML page %s: %v\n", htmlPath, err)
	}
	defer func() { _ = fHTML.Close() }()

	if err := indexPage.Execute(fHTML, string(outBytes)); err != nil {
		log.Fatalf("cannot generate HTML page: %v\n", err)
	}
}

func convertDateMonth(s string) string {
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return s
	}
	return t.Format("2006-Jan")
}

func generateSummary(v *Output) {
	const weekly = "weekly"
	in := v.Timeseries[weekly]
	lastInd := len(in.Date) - 1

	v.Summary.Stars = in.StarsTotal[lastInd]
	v.Summary.Downloads = in.DownloadsTotal[lastInd]
	v.Summary.IssuesOpen = in.IssuesOpenTotal[lastInd]
	v.Summary.PrOpen = in.PrOpenTotal[lastInd]
	v.Summary.CommittersRecurrent = in.CommittersRecurrent[lastInd]

	for _, row := range in.CommittersNew {
		v.Summary.Committers += row
	}
}

//go:embed query.sql.templ
var queryTemplate string

var queries = map[string]string{}

//go:embed index.html.templ
var indexTemplate string

var indexPage = template.Must(template.New("webpage").Parse(indexTemplate))

func init() {
	queryBuilder := template.Must(template.New("query").Parse(queryTemplate))
	frameDef := map[string]string{
		"weekly":  "%Y-W%W",
		"monthly": "%Y-%m",
	}
	for k, frame := range frameDef {
		q, err := makeQuery(queryBuilder, frame)
		if err != nil {
			log.Fatalf("cannot prepare query: %v\n", err)
		}
		queries[k] = q
	}

}

func makeQuery(queryBuilder *template.Template, frame string) (string, error) {
	var buf strings.Builder
	if err := queryBuilder.Execute(&buf, frame); err != nil {
		return "", err
	}
	return buf.String(), nil
}
