package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"

	medium "github.com/readme-update-actions/pkg/structs"
	helpers "github.com/readme-update-actions/pkg/utils"
)

func main() {
	// get the rss list from the actions env
	// rss_medium, _ := helpers.GetEnvString("INPUT_RSS_LIST")
	rss_medium := "https://imskr.medium.com/feed"

	// get the number of posts or stories to commit
	max_post, _ := helpers.GetEnvInteger("INPUT_MAX_POST")

	// if max_post not in env var set default to 3
	if max_post == 0 {
		max_post = 3
	}

	// get readme path from the actions env
	readme_path, _ := helpers.GetEnvString("INPUT_README_PATH")

	// if path not provided default to root readme
	if readme_path == "" {
		readme_path = "./README.md"
	}

	// get username
	commit_user, _ := helpers.GetEnvString("INPUT_COMMIT_USER")
	if commit_user == "" {
		commit_user = "readme-update-bot"
	}

	// git user email
	commit_email, _ := helpers.GetEnvString("INPUT_COMMIT_EMAIL")
	if commit_email == "" {
		commit_email = "readme-update-actions@example.com"
	}

	// git personal access token
	git_token, _ := helpers.GetEnvString("INPUT_GIT_TOKEN")
	if git_token == "" {
		log.Fatal("GIT_TOKEN not provided")
	}

	// git commit message
	commit_message, _ := helpers.GetEnvString("INPUT_COMMIT_MESSAGE")
	if commit_message == "" {
		commit_message = "Update readme with latest blogs"
	}

	// get medium.com rss feed
	mediumResponse, err := http.Get(rss_medium)
	if err != nil {
		log.Println("Error making request to medium", err)
	}

	defer mediumResponse.Body.Close()

	responseBody, err := ioutil.ReadAll(mediumResponse.Body)
	if err != nil {
		log.Println("Error reading response body", err)
	}

	// use RSS structs
	var rss medium.RSS
	errXMLParse := xml.Unmarshal(responseBody, &rss)
	if errXMLParse != nil {
		log.Println("Error xml parse", errXMLParse)
	}

	// store the posts
	var items []string

	// get the posts
	// format it according to readme links format
	for i := 0; i < max_post; i++ {
		item := fmt.Sprintf("- [%s](%s)\n", rss.Channel.Item[i].Title, rss.Channel.Item[i].Link)
		items = append(items, item)
	}

	// find readme and replace with our result
	err = helpers.ReplaceFile(readme_path, items)
	if err != nil {
		log.Fatalf("Error updating readme %s", err)
	}

	// set git user name
	nameCmd := exec.Command("git", "config", "user.name", commit_user)
	err = nameCmd.Run()
	if err != nil {
		log.Println("Error setting git user", err)
	}

	// set git user email
	emailCmd := exec.Command("git", "config", "user.email", commit_email)
	err = emailCmd.Run()
	if err != nil {
		log.Println("Error setting git email", err)
	}

	// add personal access token (git_token) to git
	tokenCmd := exec.Command("git", "config", "credential.helper", "store --file", "git-credential-store", "--force")
	err = tokenCmd.Run()
	if err != nil {
		log.Println("Error setting git token", err)
	}

	// store the token in git-credential-store
	tokenFile := "git-credential-store"
	tokenCmd = exec.Command("git", "credential-store", "store", "--file", tokenFile, "https://github.com", "token", git_token)
	err = tokenCmd.Run()
	if err != nil {
		log.Println("Error setting git token", err)
	}

	// add to staging area
	addCmd := exec.Command("git", "add", readme_path)
	err = addCmd.Run()
	if err != nil {
		log.Println("Error adding to staging area", err)
		return
	}

	// do git commit
	commitCmd := exec.Command("git", "commit", "-m", commit_message)
	err = commitCmd.Run()
	if err != nil {
		log.Println("Error commiting to repo", err)
		return
	}

	// do git push
	pushCmd := exec.Command("git", "push")
	err = pushCmd.Run()
	if err != nil {
		log.Println("Error pushing to repo", err)
		return
	}
}
