package install

import (
	"fmt"
)

func makeGHJsonFileName(pack, ver string) string {
	return fmt.Sprintf("gh.jpm.%s-%s.json", pack, ver)
}

func disectGithubDepString(d string) (jpmRepo, error) {
	return jpmRepo{}, nil
}

func figureOutLatestGithub(p string) (string, error) {

	return "latest", nil
}
