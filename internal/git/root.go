package git

import (
	"os/exec"
	"regexp"
	"strconv"
)

type Stats struct {
	FileChanged int
	Insertions  int
	Deletions   int
}

func DiffCached() (string, error) {
	output, err := exec.Command("git", "diff", "--cached").Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func DiffCachedStats() (*Stats, error) {
	output, err := exec.Command("git", "diff", "--cached", "--shortstat").Output()
	if err != nil {
		return nil, err
	}

	return parseDiffStats(string(output))
}

func parseDiffStats(output string) (*Stats, error) {
	fileChanged, err := parseInt(output, `(\d+)\sfiles? changed`)
	if err != nil {
		return nil, err
	}

	insertions, err := parseInt(output, `(\d+)\sinsertions?\(\+\)`)
	if err != nil {
		return nil, err
	}

	deletions, err := parseInt(output, `(\d+)\sdeletions?\(\-\)`)
	if err != nil {
		return nil, err
	}

	return &Stats{
		FileChanged: fileChanged,
		Insertions:  insertions,
		Deletions:   deletions,
	}, nil
}

func parseInt(str string, re string) (int, error) {
	matches := regexp.MustCompile(re).FindStringSubmatch(str)
	if len(matches) != 2 {
		return 0, nil
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, err
	}

	return value, nil
}
