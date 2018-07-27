package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
)

var regexDownloadUrl = regexp.MustCompile(`"browser_download_url":"(.*?\.zip)"`)

func init() {
	log.SetFlags(0)
}

func main() {
	downloadLinks, err := getLatestDownloadLinks("goreliu", "wsl-terminal")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("(main) downloadlinks: ", downloadLinks)
	if len(downloadLinks) == 0 {
		log.Fatal("No dowload links available")
	}
	downloadLink := downloadLinks[0]
	localPath, err := downloadFile(downloadLink)
	log.Println("(main) Downloaded", localPath)
}

// DownloadFile will download a url to a local file in the system temp
// directory and return file path of the downloaded file.  It's efficient
// because it will write as it downloads and not load the whole file into
// memory.
func downloadFile(downloadLink string) (localPath string, err error) {
	downloadURL, err := url.Parse(downloadLink)
	if err != nil {
		return
	}
	filename := path.Base(downloadURL.Path)
	localPath = path.Join(os.TempDir(), filename)
	fout, err := os.Create(localPath)
	if err != nil {
		return
	}
	defer fout.Close()
	resp, err := http.Get(downloadLink)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(fout, resp.Body)
	return
}

func unzip(zipPath string, targetDir string) error {

}

func getLatestDownloadLinks(owner string, repo string) ([]string, error) {
	latestReleaseUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	log.Println("(getLatestDownloadLinks) Getting release info from:", latestReleaseUrl)
	resp, err := http.Get(latestReleaseUrl)
	if err != nil {
		return nil, fmt.Errorf("Getting release info at %s: %v", latestReleaseUrl, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Parsing release info: %v", err)
	}
	releaseInfo := string(body)
	// ioutil.WriteFile("/tmp/releaseInfo.json", body, 0644)

	matches := regexDownloadUrl.FindAllStringSubmatch(releaseInfo, -1)
	//fmt.Printf("%q\n", matches)
	if matches == nil {
		return nil, fmt.Errorf("No match for download url at %s", latestReleaseUrl)
	}
	var links []string
	for _, m := range matches[1:] {
		if len(m) < 2 {
			return nil, fmt.Errorf("No submatch for link in release info at %s: %v", latestReleaseUrl, matches)
		}
		if m[1] == "" {
			return nil, fmt.Errorf("Emtpy download link in release info at %s", latestReleaseUrl)
		}
		links = append(links, m[1])
	}
	return links, nil
}
