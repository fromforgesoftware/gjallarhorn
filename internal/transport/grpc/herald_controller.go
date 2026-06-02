// Package grpc holds Gjallarhorn's gRPC controllers.
package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fromforgesoftware/go-kit/filter"
	"github.com/fromforgesoftware/go-kit/search"
	"github.com/fromforgesoftware/go-kit/search/query"
	kitgrpc "github.com/fromforgesoftware/go-kit/transport/grpc"

	"github.com/fromforgesoftware/gjallarhorn/internal/app"
	"github.com/fromforgesoftware/gjallarhorn/internal/domain"
	"github.com/fromforgesoftware/gjallarhorn/internal/fields"
	gjallarhornv1 "github.com/fromforgesoftware/gjallarhorn/pkg/api/gjallarhorn/v1"
)

type gjallarhornController struct {
	notifications app.NotificationUsecase
}

func NewGjallarhornController(notifications app.NotificationUsecase) kitgrpc.Controller {
	return &gjallarhornController{notifications: notifications}
}

func (c *gjallarhornController) SD() kitgrpc.ServiceDesc {
	return &gjallarhornv1.GjallarhornService_ServiceDesc
}

func (c *gjallarhornController) Send(ctx context.Context, req *gjallarhornv1.SendRequest) (*gjallarhornv1.SendResponse, error) {
	n := domain.NewNotification(req.GetRecipient(), req.GetChannel(),
		domain.WithNotificationRealmID(req.GetRealmId()),
		domain.WithNotificationTemplate(req.GetTemplate()),
		domain.WithNotificationData(structToMap(req.GetData())),
		domain.WithNotificationSubject(req.GetSubject()),
		domain.WithNotificationBody(req.GetBody()),
	)
	sent, err := c.notifications.Send(ctx, n)
	if err != nil {
		return nil, err
	}
	return &gjallarhornv1.SendResponse{Notification: notificationToProto(sent)}, nil
}

func (c *gjallarhornController) GetNotification(ctx context.Context, req *gjallarhornv1.GetNotificationRequest) (*gjallarhornv1.GetNotificationResponse, error) {
	n, err := c.notifications.Get(ctx, search.WithQueryOpts(query.FilterBy(filter.OpEq, fields.ID, req.GetId())))
	if err != nil {
		return nil, err
	}
	return &gjallarhornv1.GetNotificationResponse{Notification: notificationToProto(n)}, nil
}

func notificationToProto(n domain.Notification) *gjallarhornv1.Notification {
	return &gjallarhornv1.Notification{
		Id:        n.ID(),
		RealmId:   n.RealmID(),
		Recipient: n.Recipient(),
		Channel:   n.Channel(),
		Template:  n.Template(),
		Data:      mapToStruct(n.Data()),
		Subject:   n.Subject(),
		Body:      n.Body(),
		Status:    n.Status().String(),
		LastError: n.LastError(),
		CreatedAt: timestamppb.New(n.CreatedAt()),
		UpdatedAt: timestamppb.New(n.UpdatedAt()),
	}
}

func structToMap(s *structpb.Struct) map[string]any {
	if s == nil {
		return nil
	}
	return s.AsMap()
}

func mapToStruct(m map[string]any) *structpb.Struct {
	if len(m) == 0 {
		return nil
	}
	s, err := structpb.NewStruct(m)
	if err != nil {
		return nil
	}
	return s
}
