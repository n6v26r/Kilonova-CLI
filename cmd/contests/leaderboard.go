// Copyright (c) 2025 @drclcomputers. All rights reserved.
//
// This work is licensed under the terms of the MIT license.
// For a copy, see <https://opensource.org/licenses/MIT>.

package contests

import (
	"encoding/json"
	"fmt"
	"kncli/internal"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/huh/spinner"
)

type Leaderboard struct {
	Status string `json:"status"`
	Data   struct {
		ProblemNames map[string]string `json:"problem_names"`
		Entries      []struct {
			User struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"user"`
			Scores map[string]int `json:"scores"`
			Total  int            `json:"total"`
		} `json:"entries"`
	} `json:"data"`
}

func downloadLeaderboard(contestID string) {
	resp, err := internal.MakeGetRequest(fmt.Sprintf(internal.URL_CONTEST_ASSETS, contestID), nil, internal.RequestDownloadZip)
	if err != nil {
		internal.LogError(err)
		return
	}

	homedir, err := os.Getwd()
	if err != nil {
		internal.LogError(fmt.Errorf("failed to get current working directory: %w", err))
		return
	}

	downFile := filepath.Join(homedir, "leaderboard_"+contestID+".csv")
	outFile, err := os.Create(downFile)
	if err != nil {
		internal.LogError(fmt.Errorf("failed to create file %q: %w", downFile, err))
		return
	}
	defer outFile.Close()

	if err := os.WriteFile(downFile, resp, 0644); err != nil {
		internal.LogError(fmt.Errorf("failed to write to file %q: %w", downFile, err))
		return
	}

	fmt.Printf("Leaderboard to contest #%s saved to %q\n", contestID, downFile)
}

func leaderboard(contestID string) {
	url := fmt.Sprintf(internal.URL_CONTEST_LEADERBOARD, contestID)
	body, err := internal.MakeGetRequest(url, nil, internal.RequestNone)
	if err != nil {
		internal.LogError(err)
		return
	}
	var data Leaderboard
	if err = json.Unmarshal(body, &data); err != nil {
		internal.LogError(err)
		return
	}

	var Rows []table.Row

	for _, entry := range data.Data.Entries {
		var scores string
		for _, score := range entry.Scores {
			scores += fmt.Sprintf("%d  ", score)
		}
		Rows = append(Rows, table.Row{
			fmt.Sprintf("%d   %s ", entry.User.ID, entry.User.Name),
			scores,
			fmt.Sprintf("%d", entry.Total),
		})
	}

	var problemNamesTitle string
	for id, name := range data.Data.ProblemNames {
		problemNamesTitle += "| #" + id + " " + name + " "
	}

	problemNamesTitle += "|"

	Columns := []table.Column{
		{Title: "ID | Name", Width: 25},
		{Title: problemNamesTitle, Width: 50},
		{Title: "Total", Width: 5},
	}

	internal.RenderTable(Columns, Rows, 1)

	if shouldDownload {
		action := func() { downloadLeaderboard(contestID) }
		if err := spinner.New().Title("Downloading...").Action(action).Run(); err != nil {
			internal.LogError(err)
			return
		}
	}
}
