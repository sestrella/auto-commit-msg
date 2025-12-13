package git

import (
	"errors"
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

	re := regexp.MustCompile(`\s+(\d+)\s+files? changed,\s+(\d+)\s+insertions\(\+\),\s+(\d+)\s+deletions\(\-\)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 4 {
		return nil, errors.New("TODO")
	}

	fileChanged, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}

	insertions, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, err
	}

	deletions, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, err
	}

	return &Stats{
		FileChanged: fileChanged,
		Insertions:  insertions,
		Deletions:   deletions,
	}, nil
}
