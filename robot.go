package main

import (
	"fmt"
	"github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/gitlabclient"
	framework "github.com/opensourceways/community-robot-lib/robot-gitlab-framework"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

const (
	botName = "issue-assign"
)

type iClient interface {
	UnAssignIssue(projectID interface{}, issueID int, assignees []int) error
	CreateIssueComment(projectID interface{}, issueID int, comment string) error
	AssignIssue(projectID interface{}, issueID int, assignees []int) error
	ListCollaborators(projectID interface{}) ([]*gitlab.ProjectMember, error)
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewConfig() config.Config {
	return &configuration{}
}

func (bot *robot) RobotName() string {
	return botName
}

func (bot *robot) getConfig(cfg config.Config, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}

	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}

	return nil, fmt.Errorf("no config for this repo: %s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(f framework.HandlerRegister) {
	f.RegisterIssueCommentHandler(bot.handleIssueCommentEvent)
	f.RegisterIssueHandler(bot.handleIssueEvent)
}

func (bot *robot) handleIssueCommentEvent(e *gitlab.IssueCommentEvent, cfg config.Config, log *logrus.Entry) error {
	if e.ObjectKind != "note" || e.Issue.State != "opened" {
		return nil
	}

	org, repo := gitlabclient.GetIssueCommentOrgAndRepo(e)
	c, err := bot.getConfig(cfg, org, repo)
	if err != nil {
		return err
	}

	if c == nil {
		return nil
	}

	return bot.handleAssign(e)
}

func (bot *robot) handleIssueEvent(e *gitlab.IssueEvent, cfg config.Config, log *logrus.Entry) error {
	return nil
}
