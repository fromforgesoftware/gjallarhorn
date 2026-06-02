// Package internal wires Gjallarhorn's components into a single fx module that
// cmd/server composes alongside the kit's defaults.
package internal

import (
	"context"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/fx"

	"github.com/fromforgesoftware/go-kit/monitoring/logger"
	kitgrpc "github.com/fromforgesoftware/go-kit/transport/grpc"
	kitrest "github.com/fromforgesoftware/go-kit/transport/rest"

	"github.com/fromforgesoftware/gjallarhorn/internal/app"
	"github.com/fromforgesoftware/gjallarhorn/internal/db"
	gjallarhorngrpc "github.com/fromforgesoftware/gjallarhorn/internal/transport/grpc"
	gjallarhornhttp "github.com/fromforgesoftware/gjallarhorn/internal/transport/http"
)

const Version = "0.1.0"

// scheduleDispatchInterval is how often the dispatcher claims and sends due
// scheduled notifications; dispatchBatch caps each tick's claim.
const (
	scheduleDispatchInterval = 30 * time.Second
	dispatchBatch            = 100
)

func FxModule() fx.Option {
	return fx.Module("gjallarhorn",
		channelsFxModule(),
		repositoriesFxModule(),
		usecasesFxModule(),
		transportFxModule(),
	)
}

func newSMTPConfig() (app.SMTPConfig, error) {
	var cfg app.SMTPConfig
	if err := envconfig.Process("", &cfg); err != nil {
		return app.SMTPConfig{}, err
	}
	return cfg, nil
}

func newWebhookConfig() (app.WebhookConfig, error) {
	var cfg app.WebhookConfig
	if err := envconfig.Process("", &cfg); err != nil {
		return app.WebhookConfig{}, err
	}
	return cfg, nil
}

func newSMSConfig() (app.SMSConfig, error) {
	var cfg app.SMSConfig
	if err := envconfig.Process("", &cfg); err != nil {
		return app.SMSConfig{}, err
	}
	return cfg, nil
}

func newPushConfig() (app.PushConfig, error) {
	var cfg app.PushConfig
	if err := envconfig.Process("", &cfg); err != nil {
		return app.PushConfig{}, err
	}
	return cfg, nil
}

func newChannelRegistry(smtp *app.SMTPSender, webhook *app.WebhookSender, sms *app.SMSSender, push *app.PushSender) *app.ChannelRegistry {
	return app.NewChannelRegistry(smtp, webhook, sms, push)
}

func channelsFxModule() fx.Option {
	return fx.Module("gjallarhorn:channels",
		fx.Provide(
			newSMTPConfig,
			app.NewSMTPSender,
			newWebhookConfig,
			app.NewWebhookSender,
			newSMSConfig,
			app.NewSMSSender,
			newPushConfig,
			app.NewPushSender,
			newChannelRegistry,
		),
	)
}

func newNotificationUsecase(
	notifications app.NotificationRepository,
	attempts app.DeliveryAttemptRepository,
	channels *app.ChannelRegistry,
	templates app.TemplateUsecase,
	prefs app.NotificationPreferenceRepository,
) app.NotificationUsecase {
	return app.NewNotificationUsecase(notifications, attempts, channels,
		app.WithRenderer(templates), app.WithPreferences(prefs))
}

func repositoriesFxModule() fx.Option {
	return fx.Module("gjallarhorn:repositories",
		fx.Provide(
			fx.Annotate(db.NewNotificationRepository, fx.As(new(app.NotificationRepository))),
			fx.Annotate(db.NewDeliveryAttemptRepository, fx.As(new(app.DeliveryAttemptRepository))),
			fx.Annotate(db.NewNotificationTemplateRepository, fx.As(new(app.NotificationTemplateRepository))),
			fx.Annotate(db.NewNotificationPreferenceRepository, fx.As(new(app.NotificationPreferenceRepository))),
		),
	)
}

func usecasesFxModule() fx.Option {
	return fx.Module("gjallarhorn:usecases",
		fx.Provide(
			fx.Annotate(app.NewTemplateUsecase, fx.As(new(app.TemplateUsecase))),
			fx.Annotate(app.NewPreferenceUsecase, fx.As(new(app.PreferenceUsecase))),
			newNotificationUsecase,
		),
	)
}

func transportFxModule() fx.Option {
	return fx.Module("gjallarhorn:transport",
		kitrest.NewFxMiddleware(kitrest.NewGatewayMiddleware),
		fx.Invoke(registerScheduleDispatcher),
		kitgrpc.NewFxController(gjallarhorngrpc.NewGjallarhornController),
		kitrest.NewFxController(gjallarhornhttp.NewNotificationController),
		kitrest.NewFxController(gjallarhornhttp.NewNotificationTemplateController),
		kitrest.NewFxController(gjallarhornhttp.NewNotificationPreferenceController),
	)
}

// registerScheduleDispatcher runs the scheduled-notification dispatcher on an
// interval for the life of the process; OnStop cancels the loop.
func registerScheduleDispatcher(lc fx.Lifecycle, notifications app.NotificationUsecase) {
	ctx, cancel := context.WithCancel(context.Background())
	log := logger.New()
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go runScheduleDispatcher(ctx, notifications, log)
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})
}

func runScheduleDispatcher(ctx context.Context, notifications app.NotificationUsecase, log logger.Logger) {
	ticker := time.NewTicker(scheduleDispatchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := notifications.DispatchDue(ctx, dispatchBatch); err != nil {
				log.ErrorContext(ctx, "scheduled dispatch failed", "error", err)
			}
		}
	}
}
