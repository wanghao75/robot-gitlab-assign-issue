package main

import (
	"fmt"
	"github.com/opensourceways/community-robot-lib/gitlabclient"
	"github.com/xanzy/go-gitlab"
	"strings"
)

const (
	msgMultipleAssignee           = "Can only assign one assignee to the issue."
	msgAssignRepeatedly           = "This issue is already assigned to ***%s***. Please do not assign repeatedly."
	msgNotAllowAssign             = "This issue can not be assigned to ***%s***. Please try to assign to the repository collaborators."
	msgNotAllowUnassign           = "***%s*** can not be unassigned from this issue. Please try to unassign the assignee of this issue."
	msgCollaboratorCantAsAssignee = "The issue collaborator ***%s*** cannot be assigned as the assignee at the same time."
)

func (bot *robot) handleAssign(e *gitlab.IssueCommentEvent) error {
	number := e.Issue.IID
	pid := e.ProjectID
	body, author := gitlabclient.GetIssueCommentBody(e), gitlabclient.GetIssueCommentAuthor(e)

	currentAssignee := ""

	userNameID := map[string]int{}
	members, err := bot.cli.ListCollaborators(pid)
	if err != nil {
		return err
	}
	for _, m := range members {
		userNameID[m.Username] = m.ID
	}
	if e.Issue.AssigneeID != 0 {
		for _, m := range members {
			userNameID[m.Username] = m.ID
			if m.ID == e.Issue.AssigneeID {
				currentAssignee = m.Username
			}
		}
	}

	writeComment := func(s string) error {
		return bot.cli.CreateIssueComment(pid, number, s)
	}

	assign, unassign := parseCmd(body, author)
	fmt.Println(assign, unassign, number, pid, body, author, currentAssignee, e.Issue.AssigneeID)
	if n := assign.Len(); n > 0 {
		if n > 1 {
			return writeComment(msgMultipleAssignee)
		}

		if assign.Has(currentAssignee) {
			return writeComment(fmt.Sprintf(msgAssignRepeatedly, currentAssignee))
		}

		newOne := assign.UnsortedList()[0]
		fmt.Println(newOne, userNameID)
		if isIssueCollaborator(userNameID, e.Issue.AssigneeID, newOne) {
			return writeComment(fmt.Sprintf(msgCollaboratorCantAsAssignee, newOne))
		}

		err := bot.cli.AssignIssue(pid, number, []int{userNameID[newOne]})
		fmt.Println("error ", err)
		if err == nil {
			return nil
		}
		if _, ok := err.(gitlabclient.ErrorForbidden); ok {
			return writeComment(fmt.Sprintf(msgNotAllowAssign, newOne))
		}
		return err
	}

	if unassign.Len() > 0 {
		if unassign.Has(currentAssignee) {
			return bot.cli.UnAssignIssue(pid, number, []int{})
		} else {
			return writeComment(fmt.Sprintf(msgNotAllowUnassign, strings.Join(unassign.UnsortedList(), ",")))
		}
	}

	return nil
}

func isIssueCollaborator(collaborators map[string]int, assignID int, assignee string) bool {
	for v := range collaborators {
		if v == assignee && collaborators[v] == assignID {
			return true
		}
	}

	return false
}
