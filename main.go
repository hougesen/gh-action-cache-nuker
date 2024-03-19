package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type GithubRepoCacheResponse struct {
	TotalCount    int            `json:"total_count"`
	ActionsCaches []ActionsCache `json:"actions_caches"`
}
type ActionsCache struct {
	ID             int    `json:"id"`
	Ref            string `json:"ref"`
	Key            string `json:"key"`
	Version        string `json:"version"`
	LastAccessedAt string `json:"last_accessed_at"`
	CreatedAt      string `json:"created_at"`
	SizeInBytes    int    `json:"size_in_bytes"`
}

func warnInvalidToken(statusCode int) {
	fmt.Printf("[WARN] Just received a %d from GitHub - is the token valid?\n", statusCode)
}

func getRepoCacheList(token string, repo string) ([]ActionsCache, error) {
	var actionCaches GithubRepoCacheResponse

	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/caches?per_page=100&sort=created_at&direction=asc", repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []ActionsCache{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return []ActionsCache{}, err
	}

	if res.StatusCode == 401 || res.StatusCode == 403 {
		warnInvalidToken(res.StatusCode)
	}

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return []ActionsCache{}, err
	}

	decodeErr := json.Unmarshal(bodyBytes, &actionCaches)

	if decodeErr != nil {
		return []ActionsCache{}, decodeErr
	}

	return actionCaches.ActionsCaches, nil
}

func deleteRepoCacheByKey(token string, repo string, key string) (bool, error) {
	u := fmt.Sprintf("https://api.github.com/repos/%s/actions/caches?key=%s", repo, url.QueryEscape(key))

	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	if res.StatusCode == 401 || res.StatusCode == 403 {
		warnInvalidToken(res.StatusCode)
	}

	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

func deleteRepoCaches(token string, repo string) (uint, error) {
	var deletedCaches uint = 0

	actionCaches, err := getRepoCacheList(token, repo)
	if err != nil {
		return deletedCaches, err
	}

	for _, cache := range actionCaches {
		deleteResult, deleteError := deleteRepoCacheByKey(token, repo, cache.Key)

		if deleteError != nil {
			return deletedCaches, deleteError
		}

		if deleteResult {
			deletedCaches += 1
		}

	}

	return deletedCaches, nil
}

type GithubOrganizationCacheUsageRepos struct {
	TotalCount            int                    `json:"total_count"`
	RepositoryCacheUsages []RepositoryCacheUsage `json:"repository_cache_usages"`
}

type RepositoryCacheUsage struct {
	FullName                string `json:"full_name"`
	ActiveCachesSizeInBytes int    `json:"active_caches_size_in_bytes"`
	ActiveCachesCount       int    `json:"active_caches_count"`
}

func getUsageByRepository(token string, owner string) ([]string, error) {
	repoNames := []string{}

	page := 1

	for {
		var usage GithubOrganizationCacheUsageRepos

		url := fmt.Sprintf("https://api.github.com/orgs/%s/actions/cache/usage-by-repository?per_page=100&page=%d", owner, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return repoNames, err
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return repoNames, err
		}

		if res.StatusCode == 401 || res.StatusCode == 403 {
			warnInvalidToken(res.StatusCode)
		}

		defer res.Body.Close()

		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return repoNames, err
		}

		decodeErr := json.Unmarshal(bodyBytes, &usage)

		if decodeErr != nil {
			return repoNames, decodeErr
		}

		if len(usage.RepositoryCacheUsages) == 0 {
			break
		}

		for _, repo := range usage.RepositoryCacheUsages {
			repoNames = append(repoNames, repo.FullName)
		}
		page += 1

	}

	return repoNames, nil
}

func printHelp() {
	fmt.Println("Usage: <repo | org> <OWNER/NAME> <GITHUB_TOKEN>")
}

func main() {
	if len(os.Args) < 4 {
		fmt.Print("Invalid amount of input args\n\n")

		printHelp()

		os.Exit(1)
	}

	repoNames := []string{}

	action := os.Args[1]
	resource := os.Args[2]

	token := os.Args[3]

	switch action {
	case "repo":
		repoNames = append(repoNames, resource)

	case "org":
		r, err := getUsageByRepository(token, resource)
		if err != nil {
			fmt.Print("Error getting cache usage\n\n")
			fmt.Print(err)
			os.Exit(1)
		}

		if len(r) == 0 {
			fmt.Printf("'%s' has no action caches\n", resource)
		}

		repoNames = r

	default:
		fmt.Printf("Received unknown action input: '%s'\n\n", action)
		printHelp()
		os.Exit(1)
	}

	for _, repo := range repoNames {
		fmt.Printf(repo)

		keys, err := deleteRepoCaches(token, repo)
		if err != nil {
			fmt.Println("Error deleting cache")
			fmt.Print(err)
			os.Exit(1)
		}

		fmt.Printf(": removed %d caches\n", keys)
	}
}
