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
	// Slack sends channel messages and webhook notifications, resolving its API
	// token and webhook URLs through onepassword.
	Slack struct {
		l         log.Logger
		op        *onepassword.OnePassword
		cfg       Config
		token     string
		configKey string
	}
	// Option configures a Slack built by New.
	Option func(*Slack) error
)

// WorkflowPayload is the JSON body sent to a Slack workflow webhook.
type WorkflowPayload struct {
	Msg string `json:"msg"`
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// WithConfigKey overrides the viper key the Slack config is read from, which
// defaults to "slack".
func WithConfigKey(v string) Option {
	return func(o *Slack) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New builds a Slack, unmarshalling its config from viper under configKey. It
// returns an error if the unmarshalling or an option fails.
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

// Client returns a slack.Client authenticated with the API token resolved
// from onepassword, caching the token after the first lookup.
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

// Channel resolves id to its configured Slack channel ID, falling back to
// "general" if id is not configured.
func (s *Slack) Channel(id string) string {
	if value, ok := s.cfg.Channels[id]; ok {
		return value
	} else {
		return "general"
	}
}

// Webhook resolves id to its configured webhook URL via onepassword. It
// returns an error if id has no webhook configured.
func (s *Slack) Webhook(ctx context.Context, id string) (string, error) {
	value, ok := s.cfg.Webhooks[id]
	if !ok {
		return "", errors.Errorf("missing webhook configuration for %s", id)
	}

	return s.op.Get(ctx, value)
}

// SendUserMessage posts markdown to the configured channel as the current git
// user. If annotate is false, the username is prefixed into the message
// text; if true, it is appended as a context block instead. It returns an
// error if channel is not configured or the git user cannot be determined.
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

// SendETCDUpdateMessage posts a notification about an ETCD config update on
// cluster to the "releases" channel, attributed to the current git user.
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

// Send posts a message built from opts to channel, using a 5 second timeout.
// It returns an error if the client cannot be built or the message fails to
// send.
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

// SendWebhook posts blocks to the configured incoming webhook named webhook,
// using a 5 second timeout. It returns an error if webhook is not configured
// or the request fails.
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

// MarkdownSection builds a Slack section block rendering text as mrkdwn.
func (s *Slack) MarkdownSection(text string) *slack.SectionBlock {
	txt := slack.NewTextBlockObject("mrkdwn", text, false, false)
	return slack.NewSectionBlock(txt, nil, nil)
}

// DividerSection builds a Slack divider block.
func (s *Slack) DividerSection() *slack.DividerBlock {
	return slack.NewDividerBlock()
}
