package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
)

type Project struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	PathWithNamespace string    `json:"path_with_namespace"`
	WebURL            string    `json:"web_url"`
	Namespace         Namespace `json:"namespace"`
}

type Namespace struct {
	Name string `json:"name"`
}

type Variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {
	projects := getAllProjects()

	for _, project := range projects {
		fmt.Println("Processing:", project.WebURL)
		cloneRepo(project)

		projectVariables := getAllVariables(fmt.Sprintf("/projects/%d", project.ID))
		saveVariables(project.PathWithNamespace, projectVariables, "project")

		groupVariables := getAllVariables(fmt.Sprintf("/groups/%s", project.Namespace.Name))
		saveVariables(project.PathWithNamespace, groupVariables, "group")
	}
}

func getAllProjects() []Project {
	var allProjects []Project
	page := 1
	for {
		url := fmt.Sprintf("%s/api/v4/projects?private_token=%s&per_page=100&page=%d", GitLabURL, AccessToken, page)
		projects, err := getProjects(url)
		if err != nil || len(projects) == 0 {
			break
		}
		allProjects = append(allProjects, projects...)
		page++
	}
	return allProjects
}

func getProjects(url string) ([]Project, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting projects:", err)
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	var projects []Project
	err = json.Unmarshal(data, &projects)
	if err != nil {
		fmt.Println("Error unmarshalling projects:", err)
		return nil, err
	}
	return projects, nil
}

func cloneRepo(project Project) {
	path := filepath.Join(TargetDir, strings.ReplaceAll(project.PathWithNamespace, "/", string(filepath.Separator)))

	// Check if the directory already exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		fmt.Println("Directory exists, skipping clone:", path)
		return
	}

	// Ensure the directory structure is created
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory structure:", err)
		return
	}

	// Clone the repo
	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		URL: project.WebURL + ".git",
		Auth: &githttp.BasicAuth{
			Username: "api",
			Password: AccessToken,
		},
	})
	if err != nil {
		fmt.Println("Error cloning", project.WebURL, ":", err)
		return
	}

	// Fetch all references from remote, including all branches and tags
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*", "+refs/tags/*:refs/tags/*"},
		Auth: &githttp.BasicAuth{
			Username: "api",
			Password: AccessToken,
		},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		fmt.Println("Error fetching", project.WebURL, ":", err)
		return
	}

	// List all branches
	refs, err := repo.Branches()
	if err != nil {
		fmt.Println("Error fetching branches", project.WebURL, ":", err)
		return
	}

	w, err := repo.Worktree()
	if err != nil {
		fmt.Println("Error getting worktree:", err)
		return
	}

	// Iterate through each branch
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()
		fmt.Printf("Processing branch: %s\n", branchName)

		// Check if branch already exists
		_, err := repo.Branch(branchName)
		if err == nil {
			fmt.Printf("Branch %s already exists, skipping...\n", branchName)
			return nil
		}

		// Checkout each branch individually
		err = w.Checkout(&git.CheckoutOptions{
			Branch: ref.Name(),
			Force:  true, // Use force to avoid detection of possible conflicts
			Create: true, // Create local branch if not exists
		})
		if err != nil {
			fmt.Printf("Error checking out branch %s: %v\n", branchName, err)
			return err
		}

		// Here, you might copy the checked-out branch to another location if needed
		// Ensure you check out another branch before doing file operations,
		// as the file system changes will reflect the checked-out branch.
		return nil
	})
	if err != nil {
		fmt.Println("Error iterating branches", project.WebURL, ":", err)
		return
	}
}

func getAllVariables(apiEndpoint string) []Variable {
	var allVariables []Variable
	page := 1
	for {
		url := fmt.Sprintf("%s/api/v4%s/variables?private_token=%s&per_page=100&page=%d", GitLabURL, apiEndpoint, AccessToken, page)
		variables, err := getVariables(url)
		if err != nil || len(variables) == 0 {
			break
		}
		allVariables = append(allVariables, variables...)
		page++
	}
	return allVariables
}
func getVariables(url string) ([]Variable, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting variables:", err)
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	if resp.StatusCode == 404 {
		// Group not found, return an empty slice of variables
		fmt.Println("Warning: received 404 response:", string(data))
		return []Variable{}, nil
	}

	if string(data) == "[]" {
		// No variables found, return an empty slice of variables
		return []Variable{}, nil
	}

	var variables []Variable
	err = json.Unmarshal(data, &variables)
	if err != nil {
		fmt.Println("Error unmarshalling variables:", err)
		fmt.Println("Received data:", string(data))
		return nil, err
	}
	return variables, nil
}

func saveVariables(projectPath string, variables []Variable, variableType string) {
	path := filepath.Join(TargetDir, strings.ReplaceAll(projectPath, "/", string(filepath.Separator)), fmt.Sprintf("variables_%s.json", variableType))

	// Ensure the directory structure is created
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory structure:", err)
		return
	}

	data, err := json.Marshal(variables)
	if err != nil {
		fmt.Println("Error marshalling variables:", err)
		return
	}
	err = ioutil.WriteFile(path, data, os.ModePerm)
	if err != nil {
		fmt.Println("Error writing variables to file:", err)
	}
}
