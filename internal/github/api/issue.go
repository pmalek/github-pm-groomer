package api

import "github.com/google/go-github/v40/github"

type Issue github.Issue

func (i *Issue) HasLabel(label string) bool {
	for _, v := range i.Labels {
		if *v.Name == label {
			return true
		}
	}
	return false
}

func (i *Issue) RemoveLabel(label string) []string {
	var newLabels []string
	for _, v := range i.Labels {
		if *v.Name != label {
			newLabels = append(newLabels, *v.Name)
		}
	}
	return newLabels
}

func (i *Issue) ReplaceLabel(label string, newLabel string) []string {
	var newLabels []string
	for _, v := range i.Labels {
		if *v.Name == label {
			newLabels = append(newLabels, newLabel)
		} else {
			newLabels = append(newLabels, *v.Name)
		}
	}
	return newLabels
}

func (i *Issue) AddLabel(label string) []string {
	seen := false
	var newLabels []string
	for _, v := range i.Labels {
		if *v.Name == label {
			seen = true
		}
		newLabels = append(newLabels, *v.Name)
	}
	if !seen {
		newLabels = append(newLabels, label)
	}
	return newLabels
}
