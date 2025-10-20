package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/util/git"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

type (
	Slack struct {
		l         log.Logger
		op        *onepassword.OnePassword
		cfg       Config
		token     string
		configKey string
	}
	Option func(*Slack) error
)

type WorkflowPayload struct {
	Msg string `json:"msg"`
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfigKey(v string) Option {
	return func(o *Slack) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(l log.Logger, op *onepassword.OnePassword, opts ...Option) (*Slack, error) {
	inst := &Slack{
		l:         l,
		op:        op,
		configKey: "slack",
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(inst); err != nil {
				return nil, err
			}
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (s *Slack) Client(ctx context.Context) (*slack.Client, error) {
	if s.token == "" {
		if value, err := s.op.Get(ctx, s.cfg.Token); err != nil {
			return nil, err
		} else {
			s.token = value
		}
	}

	return slack.New(s.token, slack.OptionDebug(s.l.IsLevel(log.LevelTrace))), nil
}

func (s *Slack) Channel(id string) string {
	if value, ok := s.cfg.Channels[id]; ok {
		return value
	} else {
		return "general"
	}
}

func (s *Slack) Webhook(ctx context.Context, id string) (string, error) {
	value, ok := s.cfg.Webhooks[id]
	if !ok {
		return "", errors.Errorf("missing webhook configuration for %s", id)
	}

	return s.op.Get(ctx, value)
}

func (s *Slack) SendUserMessage(ctx context.Context, markdown, channel string, annotate bool) error {
	ch, ok := s.cfg.Channels[channel]
	if !ok {
		return errors.Errorf("channel not found: %s", channel)
	}

	user, err := git.ConfigUserName(ctx, s.l)
	if err != nil {
		return errors.Wrap(err, "failed to get git user")
	}

	if !annotate {
		markdown = fmt.Sprintf("*%s*: %s", user, markdown)
	}

	blocks := []slack.Block{s.MarkdownSection(markdown)}
	if annotate {
		blocks = append(blocks, slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", "by "+user, false, false)))
	}

	fallbackOpt := slack.MsgOptionText(markdown, false)

	return s.Send(ctx, ch, slack.MsgOptionCompose(fallbackOpt, slack.MsgOptionBlocks(blocks...)))
}

func (s *Slack) SendETCDUpdateMessage(ctx context.Context, cluster string) error {
	user, err := git.ConfigUserName(ctx, s.l)
	if err != nil {
		pterm.Debug.Println("failed to get git user: " + err.Error())

		user = "unknown"
	}

	msg := s.MarkdownSection(fmt.Sprintf("📝 *ETCD* config update on *%s*", cluster))

	blockOpt := slack.MsgOptionBlocks(
		msg,
		slack.NewContextBlock("", slack.NewTextBlockObject("mrkdwn", "by "+user, false, false)),
		s.DividerSection(),
	)
	fallbackOpt := slack.MsgOptionText(fmt.Sprintf("ETCD config update on %s", cluster), false)

	if err := s.Send(ctx,
		s.cfg.Channels["releases"],
		slack.MsgOptionCompose(fallbackOpt, blockOpt),
	); err != nil {
		return err
	}

	s.l.Info("💌 sent slack notification")

	return nil
}

func (s *Slack) Send(ctx context.Context, channel string, opts ...slack.MsgOption) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := s.Client(ctx)
	if err != nil {
		return err
	}

	if _, _, _, err = client.SendMessageContext(ctx, channel, opts...); err != nil {
		return err
	}

	s.l.Info("💌 sent slack notification")

	return nil
}

func (s *Slack) SendWebhook(ctx context.Context, webhook string, blocks []slack.Block) error {
	url, err := s.Webhook(ctx, webhook)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := slack.PostWebhookContext(ctx, url, &slack.WebhookMessage{
		Blocks: &slack.Blocks{
			BlockSet: blocks,
		},
	}); err != nil {
		return err
	}

	s.l.Info("💌 sent slack notification")

	return nil
}

func (s *Slack) MarkdownSection(text string) *slack.SectionBlock {
	txt := slack.NewTextBlockObject("mrkdwn", text, false, false)
	return slack.NewSectionBlock(txt, nil, nil)
}

func (s *Slack) DividerSection() *slack.DividerBlock {
	return slack.NewDividerBlock()
}
