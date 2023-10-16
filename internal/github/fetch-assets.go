package github

import "encoding/json"

type ReleaseAsset struct {
	ReleaseName   string
	AssetName     string
	DownloadCount int
}

func FetchReleaseAssets(client *Client, repoOwner, repoName string) ([]ReleaseAsset, error) {
	var query = `{
  repository(owner: "` + repoOwner + `", name: "` + repoName + `") {
    releases(first: 100) {
      nodes {
        name
        releaseAssets(first: 100) {
          nodes {
            name
            downloadCount
          }
        }
      }
      pageInfo{
        hasPreviousPage
        startCursor
      }
    }
  }
}`
	data, err := client.Query(query)
	if err != nil {
		return nil, err
	}

	var res struct {
		Repository struct {
			Releases struct {
				Nodes []struct {
					Name          string `json:"name"`
					ReleaseAssets struct {
						Nodes []struct {
							Name          string `json:"name"`
							DownloadCount int    `json:"downloadCount"`
						} `json:"nodes"`
					} `json:"releaseAssets"`
				} `json:"nodes"`
			} `json:"releases"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	var o []ReleaseAsset
	for _, release := range res.Repository.Releases.Nodes {
		for _, asset := range release.ReleaseAssets.Nodes {
			o = append(o, ReleaseAsset{
				ReleaseName:   release.Name,
				AssetName:     asset.Name,
				DownloadCount: asset.DownloadCount,
			})
		}
	}

	return o, nil
}
