package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
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

	targDir := os.Getenv("HOME") + "/sdk"
	unzip(localPath, targDir)
	log.Println("(main) Un-zipped into:", targDir)
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

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
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
