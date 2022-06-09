package main

import (
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

var assignRe = regexp.MustCompile(`(?mi)^/(un)?Ass(( @?[-\w]+?)*)\s*$`)

func parseCmd(comment, commenter string) (sets.String, sets.String) {
	assign := sets.NewString()
	unassign := sets.NewString()

	matches := assignRe.FindAllStringSubmatch(comment, -1)
	for _, re := range matches {
		v := unassign
		if re[1] == "" {
			v = assign
		}

		if re[2] == "" {
			v.Insert(commenter)
		} else {
			v.Insert(parseLogins(re[2])...)
		}
	}

	return assign, unassign
}

func parseLogins(text string) []string {
	var parts []string
	for _, s := range strings.Split(text, " ") {
		if v := strings.Trim(s, "@ "); v != "" {
			parts = append(parts, v)
		}
	}
	return parts
}
